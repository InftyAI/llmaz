/*
Copyright 2023.

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

package controller

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	llmazv1alpha1 "inftyai.io/llmaz/api/v1alpha1"
)

// ServeReconciler reconciles a Serve object
type ServeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
	ctx    context.Context
}

//+kubebuilder:rbac:groups=llmaz.inftyai.io,resources=serves,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=llmaz.inftyai.io,resources=serves/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=llmaz.inftyai.io,resources=serves/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Serve object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ServeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.ctx = ctx
	//log := log.FromContext(ctx, "Serve", req.NamespacedName)
	log := r.Log.WithValues("Serve", req.NamespacedName)
	log.Info("Reconciliation start")

	// TODO(user): your logic here
	serve := &llmazv1alpha1.Serve{}
	if err := r.Get(ctx, req.NamespacedName, serve); err != nil {
		log.Error(err, "unable to fetch Serve")
		// we'll ignore not-found errors, since there is nothing to do.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO: add finalizer here

	if !r.needToCreateWorkload(serve) {
		log.V(1).Info("No need to create Workload")
		return ctrl.Result{}, nil
	}

	if err := r.Status().Update(r.ctx, serve); err != nil {
		log.Error(err, "Failed to update serve status")
		return ctrl.Result{}, err
	}

	// create deployment
	deployment, err := r.generateDeployment(serve)
	if err != nil {
		log.Error(err, "Failed to generate deployment")
		return ctrl.Result{}, err
	}
	serve.Status.ResourceRef = make(map[string]string)
	serve.Status.ResourceRef["workload"] = deployment.GetName()

	if err := controllerutil.SetControllerReference(serve, deployment, r.Scheme); err != nil {
		log.Error(err, "Failed to SetControllerReference for deployment")
		return ctrl.Result{}, err
	}

	if err := r.Create(r.ctx, deployment); err != nil {
		log.Error(err, "Failed to create deployment", "deployment", deployment.GetName())
		return ctrl.Result{}, err
	}

	log.V(1).Info("Deployment created", "Deployment", deployment.GetName())
	serve.Status.State = "Running"

	// create service
	service, err := r.generateService(serve)
	if err != nil {
		log.Error(err, "Failed to create service")
		return ctrl.Result{}, err
	}

	service.SetOwnerReferences(nil)
	if err := controllerutil.SetControllerReference(serve, service, r.Scheme); err != nil {
		log.Error(err, "Failed to SetControllerReference for service", "Service", service.GetName())
		return ctrl.Result{}, err
	}

	if err := r.Create(r.ctx, service); err != nil {
		log.Error(err, "Failed to create service", "service", service.GetName())
		return ctrl.Result{}, err
	}

	if err := r.Status().Update(r.ctx, serve); err != nil {
		log.Error(err, "Failed to update serve status")
		return ctrl.Result{}, err
	}

	log.V(1).Info("Service created", "Service", service.GetName())

	// TODO: add Scaler for deployment
	//if err := r.createScaler(serve, deployment, service); err != nil {
	//	log.Error(err, "Failed to create Keda scaler")
	//	return ctrl.Result{}, nil
	//}

	return ctrl.Result{}, nil
}

func (r *ServeReconciler) generateDeployment(s *llmazv1alpha1.Serve) (*appsv1.Deployment, error) {
	if &s.Spec.Template == nil {
		return &appsv1.Deployment{}, nil
	}

	labels := map[string]string{
		"llmaz.inftyai.io/managed": "true",
		"llmaz.inftyai.io/serve":   s.Name,
	}
	selector := &metav1.LabelSelector{
		MatchLabels: labels,
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-deployment-", s.Name),
			Namespace:    s.Namespace,
			Labels:       labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &s.Spec.Replicas,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: *s.Spec.Template,
			},
		},
	}
	if workload, exists := s.Status.ResourceRef["workload"]; exists && workload == deployment.GetName() {
		return &appsv1.Deployment{}, nil
	}
	s.Status.ResourceRef = make(map[string]string)
	s.Status.ResourceRef["workload"] = deployment.GetName()

	return deployment, nil
}

func (r *ServeReconciler) generateService(s *llmazv1alpha1.Serve) (*corev1.Service, error) {
	labels := map[string]string{
		"llmaz.inftyai.io/managed": "true",
		"llmaz.inftyai.io/serve":   s.Name,
	}

	var svcPort []corev1.ServicePort
	for i, container := range s.Spec.Template.Containers {
		//svcPort[i].Port = container.Ports[i].ContainerPort
		svcPort = append(svcPort, corev1.ServicePort{
			Port:       container.Ports[i].ContainerPort,
			TargetPort: intstr.IntOrString{IntVal: container.Ports[i].ContainerPort},
		})
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: fmt.Sprintf("%s-service-", s.Name),
			Namespace:    s.Namespace,
			Labels:       labels,
		},
		Spec: corev1.ServiceSpec{
			Ports:    svcPort,
			Selector: labels,
		},
	}
	s.Status.Service = service.GetName()

	return service, nil
}

func (r *ServeReconciler) needToCreateWorkload(s *llmazv1alpha1.Serve) bool {
	log := r.Log.WithName("NeedToCreateWorkload").
		WithValues("Serve", fmt.Sprintf("%s/%s", s.Namespace, s.Name))

	// Workload had not created, need to create.
	if len(s.Status.State) == 0 || s.Status.ResourceRef == nil {
		log.V(1).Info("Workload not created")
		return true
	}

	return false
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&llmazv1alpha1.Serve{}).
		Complete(r)
}
