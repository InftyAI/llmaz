/*
Copyright 2025 The InftyAI Team.

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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	aggregator "github.com/inftyai/router/pkg/metrics-aggregator"
	"github.com/inftyai/router/pkg/util"
)

// PodReconciler reconciles a Model object
type PodReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Record record.EventRecorder
	Agg    *aggregator.Aggregator
}

func NewPodReconciler(client client.Client, scheme *runtime.Scheme, record record.EventRecorder, agg *aggregator.Aggregator) *PodReconciler {
	return &PodReconciler{
		Client: client,
		Scheme: scheme,
		Record: record,
		Agg:    agg,
	}
}

//+kubebuilder:rbac:groups="",resources=events,verbs=create;watch;update;patch
//+kubebuilder:rbac:groups=core,resources=pods,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *PodReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var pod corev1.Pod
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, &pod); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(10).Info("reconcile Pod", "Pod", klog.KObj(&pod))

	if isPodTerminating(&pod) {
		r.Agg.DeletePod(r.Agg.KeyFunc(&pod))
		// TODO: this is only for debug, remove it later.
		logger.V(0).Info("Pod terminating", "PodMap length", r.Agg.Len())
		return ctrl.Result{}, nil
	}

	if isPodReady(&pod) {
		r.Agg.AddPod(&pod)
		// TODO: this is only for debug, remove it later.
		logger.V(0).Info("Pod Ready", "PodMap length", r.Agg.Len())
		return ctrl.Result{}, nil
	}

	// If Pod not ready, remove from the store.
	// TODO: should we mark it as not ready in the store rather than delete it?
	r.Agg.DeletePod(r.Agg.KeyFunc(&pod))
	// TODO: this is only for debug, remove it later.
	logger.V(0).Info("Pod not ready", "PodMap length", r.Agg.Len())
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Pod{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return hasLabel(e.Object, util.ModelNameLabelKey)
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return hasLabel(e.ObjectOld, util.ModelNameLabelKey)
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return hasLabel(e.Object, util.ModelNameLabelKey)
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return hasLabel(e.Object, util.ModelNameLabelKey)
			},
		}).
		Complete(r)
}

func hasLabel(obj client.Object, key string) bool {
	_, ok := obj.GetLabels()[key]
	return ok
}

func isPodReady(pod *corev1.Pod) bool {
	if !pod.DeletionTimestamp.IsZero() {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			if condition.Status == corev1.ConditionTrue {
				return true
			}
			break
		}
	}
	return false
}

func isPodTerminating(pod *corev1.Pod) bool {
	return !pod.DeletionTimestamp.IsZero()
}
