/*
Copyright 2024 The InftyAI Team.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package inference

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	llmazcoreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	llmazcorev1alpha1 "github.com/inftyai/llmaz/api/core/v1alpha1"
)

var (
	activatorControllerLog = ctrl.Log.WithName("activator-controller")
)

const (
	playgroundsResource     = "playgrounds"
	activatorControllerName = "activator-controller"
)

type ActivatorReconciler struct {
	client.Client
	dynamicClient dynamic.Interface
	portManager   *PortManager
	ip            string
}

func NewActivatorReconciler(mgr ctrl.Manager, dynamicClient dynamic.Interface, ip string) *ActivatorReconciler {
	reconciler := &ActivatorReconciler{
		Client:        mgr.GetClient(),
		dynamicClient: dynamicClient,
		ip:            ip,
	}
	reconciler.portManager = NewPortManager(reconciler.scaleUp)
	return reconciler
}

// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;update;patch;delete
// +kubebuilder:rbac:groups="",resources=endpoints,verbs=get;list;watch;update;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ActivatorReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	svc := &corev1.Service{}
	if err := r.Get(ctx, req.NamespacedName, svc); err != nil {
		if errors.IsNotFound(err) {
			r.handleServiceDeletion(req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	if err := r.restoreSelectorIfNeeded(ctx, svc); err != nil {
		return ctrl.Result{}, err
	}

	ep := &corev1.Endpoints{}
	if err := r.Get(ctx, req.NamespacedName, ep); err != nil {
		if errors.IsNotFound(err) {
			activatorControllerLog.Info("Endpoints not found, waiting for creation", "service", svc.Name)
			return ctrl.Result{}, nil
		}
		activatorControllerLog.Error(err, "Failed to get endpoints", "service", svc.Name)
		return ctrl.Result{}, err
	}

	// Check if the service has the activator annotation
	ports, ok := r.needInject(svc)
	if !ok {
		activatorControllerLog.Info("Activator annotation not found, skipping", "service", svc.Name)
		return ctrl.Result{}, nil
	}

	if len(ep.Subsets) == 0 {
		// If the endpoints are empty, inject the activator IP
		return ctrl.Result{}, r.injectEndpoint(ctx, ep, svc, ports)
	} else if ep.Subsets[0].Addresses != nil &&
		len(ep.Subsets[0].Addresses) > 0 &&
		ep.Subsets[0].Addresses[0].IP != r.ip {
		// If the endpoints are not empty and not the activator IP, forward the traffic
		return ctrl.Result{}, r.forwardEndpoint(ctx, ep, ports)
	}

	return ctrl.Result{}, nil
}

func (r *ActivatorReconciler) needInject(svc *corev1.Service) ([]corev1.ServicePort, bool) {
	if svc == nil || svc.Annotations == nil {
		return nil, false
	}
	if _, ok := svc.Annotations[llmazcoreapi.ModelActivatorAnnoKey]; !ok {
		return nil, false
	}
	if len(svc.Spec.Ports) == 0 || svc.Spec.Type != corev1.ServiceTypeClusterIP {
		return nil, false
	}

	validPorts := make([]corev1.ServicePort, 0, len(svc.Spec.Ports))
	for _, port := range svc.Spec.Ports {
		if port.Port == 0 || port.Protocol != corev1.ProtocolTCP {
			continue
		}
		validPorts = append(validPorts, port)
	}
	if len(validPorts) == 0 {
		return nil, false
	}
	return validPorts, true
}

func (r *ActivatorReconciler) restoreSelectorIfNeeded(ctx context.Context, svc *corev1.Service) error {
	selectorStr := svc.Annotations[llmazcoreapi.CachedModelActivatorAnnoKey]
	if selectorStr == "" {
		return nil
	}

	sel := map[string]string{}
	if err := json.Unmarshal([]byte(selectorStr), &sel); err != nil {
		activatorControllerLog.Error(err, "Failed to unmarshal selector")
		return err
	}

	updatedSvc := svc.DeepCopy()
	delete(updatedSvc.Annotations, llmazcoreapi.CachedModelActivatorAnnoKey)
	updatedSvc.Spec.Selector = sel

	if err := r.Update(ctx, updatedSvc); err != nil {
		activatorControllerLog.Error(err, "Failed to restore service selector")
		return err
	}

	activatorControllerLog.Info("Restored service selector", "selector", sel)
	return nil
}

func (r *ActivatorReconciler) injectEndpoint(ctx context.Context, ep *corev1.Endpoints, svc *corev1.Service, ports []corev1.ServicePort) error {
	subsets := make([]corev1.EndpointSubset, 0, len(ports))
	for _, port := range ports {
		ds, err := r.portManager.AddTarget(ep.Name, ep.Namespace, int(port.Port))
		if err != nil {
			return err
		}

		activatorControllerLog.Info("Injecting endpoint",
			"port", port.Port,
			"listenerPort", ds.Listener.Port(),
		)

		subsets = append(subsets, corev1.EndpointSubset{
			Addresses: []corev1.EndpointAddress{{IP: r.ip}},
			Ports: []corev1.EndpointPort{{
				Name: port.Name,
				Port: int32(ds.Listener.Port()),
			}},
		})
	}

	updatedEp := ep.DeepCopy()
	updatedEp.Subsets = subsets
	if err := r.Update(ctx, updatedEp); err != nil {
		activatorControllerLog.Error(err, "Failed to update endpoints")
		return err
	}

	// Save the original selector to annotation and clear the selector
	selectorBytes, _ := json.Marshal(svc.Spec.Selector)
	updatedSvc := svc.DeepCopy()
	if updatedSvc.Annotations == nil {
		updatedSvc.Annotations = make(map[string]string)
	}
	updatedSvc.Annotations[llmazcoreapi.CachedModelActivatorAnnoKey] = string(selectorBytes)
	updatedSvc.Spec.Selector = nil
	return r.Update(ctx, updatedSvc)
}

func (r *ActivatorReconciler) handleServiceDeletion(namespace, name string) {
	pis := r.portManager.RemoveTargetForAllPorts(name, namespace)
	for _, pi := range pis {
		activatorControllerLog.Info("Cleaning up endpoints after service deletion",
			"port", pi.Target.Port,
			"listenerPort", pi.Listener.Port(),
		)
		_ = pi.Listener.Close()
		for _, conn := range pi.Connections {
			_ = conn.Close()
		}
	}
}

func (r *ActivatorReconciler) forwardEndpoint(ctx context.Context, ep *corev1.Endpoints, ports []corev1.ServicePort) error {
	for _, port := range ports {
		ds := r.portManager.RemoveTarget(ep.Name, ep.Namespace, int(port.Port))
		if ds == nil {
			continue
		}

		address, err := r.getEndpointAddress(ep, ports, &ds.Target)
		if err != nil {
			activatorControllerLog.Error(err, "Failed to get endpoint address")
			continue
		}

		activatorControllerLog.Info("Forwarding traffic to real endpoint",
			"port", port.Port,
			"address", address,
			"connections", len(ds.Connections),
		)

		for _, conn := range ds.Connections {
			targetConn, err := net.Dial("tcp", address)
			if err != nil {
				activatorControllerLog.Error(err, "Failed to dial target")
				continue
			}
			tunnel(conn, targetConn)
		}
		err = ds.Listener.Close()
		if err != nil {
			activatorControllerLog.Error(err, "Failed to close listener")
			continue
		}
	}
	return nil
}

func (r *ActivatorReconciler) getEndpointAddress(ep *corev1.Endpoints, ports []corev1.ServicePort, target *Target) (string, error) {
	for _, port := range ports {
		if int(port.Port) != target.Port {
			continue
		}

		for _, subset := range ep.Subsets {
			if len(subset.Addresses) == 0 {
				continue
			}
			for _, p := range subset.Ports {
				if port.TargetPort.Type == intstr.Int && int(p.Port) == int(port.TargetPort.IntVal) {
					return fmt.Sprintf("%s:%d", subset.Addresses[0].IP, p.Port), nil
				}
			}
		}
	}
	return "", fmt.Errorf("address not found for port %d", target.Port)
}

func (r *ActivatorReconciler) scaleUp(pi *PortInformation) {
	ctx := context.Background()
	activatorControllerLog.Info("Scaling up target Playground", "service", pi.Target.Name)

	svc := &corev1.Service{}
	key := types.NamespacedName{Namespace: pi.Target.Namespace, Name: pi.Target.Name}
	if err := r.Get(ctx, key, svc); err != nil {
		activatorControllerLog.Error(err, "Failed to get service")
		return
	}

	name := svc.Annotations[llmazcoreapi.ModelActivatorAnnoKey]
	if name == "" {
		activatorControllerLog.Error(nil, "Scale annotation not found")
		return
	}

	gvr := llmazcorev1alpha1.GroupVersion.WithResource(playgroundsResource)

	activatorControllerLog.Info("Scaling up Playground", "playground", name)
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		playground, err := r.dynamicClient.Resource(gvr).Namespace(pi.Target.Namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return err
		}
		if err := unstructured.SetNestedField(playground.Object, int64(1), "spec", "replicas"); err != nil {
			return err
		}
		_, err = r.dynamicClient.Resource(gvr).Namespace(pi.Target.Namespace).Update(ctx, playground, metav1.UpdateOptions{})
		return err
	})

	if retryErr != nil {
		activatorControllerLog.Error(retryErr, "Failed to scale Playground")
		return
	}

	if err := r.waitUntilPlaygroundPodIsReady(ctx, name, pi.Target.Namespace); err != nil {
		activatorControllerLog.Error(err, "Failed waiting for Playground pod")
		return
	}

	// Restore the service selector
	restoreSelectorIfNeededErr := r.restoreSelectorIfNeeded(ctx, svc)
	if restoreSelectorIfNeededErr != nil {
		activatorControllerLog.Error(restoreSelectorIfNeededErr, "Failed to restore service selector")
		return
	}
}

func (r *ActivatorReconciler) waitUntilPlaygroundPodIsReady(ctx context.Context, name, namespace string) error {
	// The pod name is always playground name + "-0"
	podName := name + "-0"
	return wait.PollUntilContextTimeout(ctx, time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		pod := &corev1.Pod{}
		if err := r.Get(ctx, types.NamespacedName{Namespace: namespace, Name: podName}, pod); err != nil {
			if errors.IsNotFound(err) {
				return false, nil
			}
			return false, err
		}
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				return true, nil
			}
		}
		return false, nil
	})
}

func (r *ActivatorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	hasActivatorAnnotation := func(obj client.Object) bool {
		// Make sure the object has the activator annotation
		annotations := obj.GetAnnotations()
		_, ok := annotations[llmazcoreapi.ModelActivatorAnnoKey]
		if ok {
			activatorControllerLog.V(4).Info("Object has activator annotation", "object", obj.GetName())
		}

		return ok
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(activatorControllerName).
		For(&corev1.Service{}, builder.WithPredicates(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return hasActivatorAnnotation(e.Object)
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return hasActivatorAnnotation(e.ObjectNew) || hasActivatorAnnotation(e.ObjectOld)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return hasActivatorAnnotation(e.Object)
			},
		})).
		Watches(
			&corev1.Endpoints{},
			handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, obj client.Object) []reconcile.Request {
				return []reconcile.Request{
					{NamespacedName: types.NamespacedName{
						Namespace: obj.GetNamespace(),
						Name:      obj.GetName(),
					}},
				}
			}),
			builder.WithPredicates(predicate.Funcs{
				CreateFunc: func(e event.CreateEvent) bool {
					return hasActivatorAnnotation(e.Object)
				},
				UpdateFunc: func(e event.UpdateEvent) bool {
					return hasActivatorAnnotation(e.ObjectNew)
				},
				DeleteFunc: func(e event.DeleteEvent) bool {
					return hasActivatorAnnotation(e.Object)
				},
			}),
		).
		Complete(r)
}

func tunnel(a, b net.Conn) {
	go io.Copy(a, b)
	go io.Copy(b, a)
}

type Listener interface {
	net.Listener
	Port() int
}

type listener struct {
	net.Listener
	port int
}

func NewListener() (Listener, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, err
	}
	return &listener{
		Listener: l,
		port:     l.Addr().(*net.TCPAddr).Port,
	}, nil
}

func (l *listener) Accept() (net.Conn, error) {
	c, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (l *listener) Port() int {
	return l.port
}

type Target struct {
	Name      string
	Namespace string
	Port      int
}

type PortInformation struct {
	Target      Target
	Listener    Listener
	Connections []net.Conn
}

type PortManager struct {
	portMap        map[int]*PortInformation
	reversePortMap map[Target]int
	mut            sync.Mutex

	cb func(*PortInformation)
}

func NewPortManager(cb func(*PortInformation)) *PortManager {
	return &PortManager{
		portMap:        map[int]*PortInformation{},
		reversePortMap: map[Target]int{},
		cb:             cb,
	}
}

func (pm *PortManager) AddTarget(name string, namespace string, port int) (*PortInformation, error) {
	pm.mut.Lock()
	defer pm.mut.Unlock()

	target := Target{
		Name:      name,
		Namespace: namespace,
		Port:      port,
	}

	port, ok := pm.reversePortMap[target]
	if ok {
		return pm.portMap[port], nil
	}

	listener, err := NewListener()
	if err != nil {
		return nil, err
	}
	port = listener.Port()
	downstream := &PortInformation{
		Target:   target,
		Listener: listener,
	}
	pm.portMap[port] = downstream
	pm.reversePortMap[target] = port

	go pm.startListener(downstream)
	return downstream, nil
}

func (pm *PortManager) RemoveTarget(name string, namespace string, port int) *PortInformation {
	pm.mut.Lock()
	defer pm.mut.Unlock()

	target := Target{
		Name:      name,
		Namespace: namespace,
		Port:      port,
	}

	port, ok := pm.reversePortMap[target]
	if !ok {
		return nil
	}
	downstream := pm.portMap[port]
	delete(pm.portMap, port)
	delete(pm.reversePortMap, target)
	return downstream
}

func (pm *PortManager) RemoveTargetForAllPorts(name string, namespace string) []*PortInformation {
	pm.mut.Lock()
	defer pm.mut.Unlock()

	var downstreams []*PortInformation
	for port, downstream := range pm.portMap {
		if downstream.Target.Name == name && downstream.Target.Namespace == namespace {
			delete(pm.portMap, port)
			delete(pm.reversePortMap, downstream.Target)
			downstreams = append(downstreams, downstream)
		}
	}
	return downstreams
}

func (pm *PortManager) startListener(downstream *PortInformation) {
	start := false
	for {
		conn, err := downstream.Listener.Accept()
		if err != nil {
			return
		}
		downstream.Connections = append(downstream.Connections, conn)
		if !start {
			go pm.cb(downstream)
			start = true
		}
	}
}
