/*
Copyright 2024.

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

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	"inftyai.com/llmaz/pkg/util"
)

// ServiceReconciler reconciles a Service object
type ServiceReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Record record.EventRecorder
}

func NewServiceReconciler(client client.Client, scheme *runtime.Scheme, record record.EventRecorder) *ServiceReconciler {
	return &ServiceReconciler{
		Client: client,
		Scheme: scheme,
		Record: record,
	}
}

//+kubebuilder:rbac:groups=inference.llmaz.io,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inference.llmaz.io,resources=services/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inference.llmaz.io,resources=services/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	service := &inferenceapi.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, service); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	workload := buildWorkload(service)
	if err := ctrl.SetControllerReference(service, workload, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// TODO: handle fungibility

	if err := util.Patch(ctx, r.Client, workload); err != nil {
		log.Error(err, "failed to create leaderworkerset")
		return ctrl.Result{}, err
	}

	// Handle status.

	workload = &lws.LeaderWorkerSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, workload); err != nil {
		log.Error(err, "failed to get leaderworkerset")
		return ctrl.Result{}, err
	}
	setServiceCondition(service, workload)
	if err := r.Status().Update(ctx, service); err != nil {
		log.Error(err, "failed to update Service status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ServiceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inferenceapi.Service{}).
		Complete(r)
}

func buildWorkload(service *inferenceapi.Service) *lws.LeaderWorkerSet {
	workload := &lws.LeaderWorkerSet{}
	workload.Kind = "LeaderWorkerSet"
	workload.APIVersion = lws.GroupVersion.Group + "/" + lws.GroupVersion.Version
	workload.Name = service.Name
	workload.Namespace = service.Namespace
	workload.Spec = service.Spec.WorkloadTemplate
	return workload
}

func setServiceCondition(service *inferenceapi.Service, workload *lws.LeaderWorkerSet) {
	if apimeta.IsStatusConditionTrue(workload.Status.Conditions, string(lws.LeaderWorkerSetAvailable)) {
		condition := metav1.Condition{
			Type:    inferenceapi.ServiceAvailable,
			Status:  metav1.ConditionTrue,
			Reason:  "ServiceReady",
			Message: "InferenceService is ready",
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, condition)
	} else {
		condition := metav1.Condition{
			Type:    inferenceapi.ServiceProgressing,
			Status:  metav1.ConditionTrue,
			Reason:  "ServiceInProgress",
			Message: "InferenceService is progressing",
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, condition)

		// Set the available to false
		new_condition := metav1.Condition{
			Type:   inferenceapi.ServiceAvailable,
			Status: metav1.ConditionFalse,
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, new_condition)
	}
}
