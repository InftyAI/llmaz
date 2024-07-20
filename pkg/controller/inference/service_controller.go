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
	"fmt"

	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metaapplyv1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"
	applyconfigurationv1 "sigs.k8s.io/lws/client-go/applyconfiguration/leaderworkerset/v1"

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
	_ = log.FromContext(ctx)

	service := &inferenceapi.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, service); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	workloadApplyConfiguration := buildWorkloadApplyConfiguration(service)
	if err := setControllerReferenceForLWS(service, workloadApplyConfiguration, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// TODO: handle fungibility

	if err := util.Patch(ctx, r.Client, workloadApplyConfiguration); err != nil {
		return ctrl.Result{}, err
	}

	// Handle status.

	workload := &lws.LeaderWorkerSet{}
	if err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, workload); err != nil {
		return ctrl.Result{}, err
	}
	setServiceCondition(service, workload)
	if err := r.Status().Update(ctx, service); err != nil {
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

func buildWorkloadApplyConfiguration(service *inferenceapi.Service) *applyconfigurationv1.LeaderWorkerSetApplyConfiguration {
	workload := applyconfigurationv1.LeaderWorkerSet(service.Name, service.Namespace)

	leaderWorkerTemplate := applyconfigurationv1.LeaderWorkerTemplate()
	leaderWorkerTemplate.WithWorkerTemplate(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate)

	spec := applyconfigurationv1.LeaderWorkerSetSpec()
	spec.WithLeaderWorkerTemplate(leaderWorkerTemplate)
	spec.WithReplicas(*service.Spec.WorkloadTemplate.Replicas)

	workload.WithSpec(spec)
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

		if apimeta.FindStatusCondition(service.Status.Conditions, inferenceapi.ServiceAvailable) == nil {
			return
		}

		// Set the available to false
		new_condition := metav1.Condition{
			Type:    inferenceapi.ServiceAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "ServiceNotReady",
			Message: "InferenceService not ready",
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, new_condition)
	}
}

// setControllerReferenceForLWS set service as the owner reference for lws.
func setControllerReferenceForLWS(owner metav1.Object, lws *applyconfigurationv1.LeaderWorkerSetApplyConfiguration, scheme *runtime.Scheme) error {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("%T is not a runtime.Object, cannot call SetOwnerReference", owner)
	}
	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}
	lws.WithOwnerReferences(metaapplyv1.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().String()).
		WithKind(gvk.Kind).
		WithName(owner.GetName()).
		WithUID(owner.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true))
	return nil
}
