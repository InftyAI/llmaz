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
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metaapplyv1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"
	applyconfigurationv1 "sigs.k8s.io/lws/client-go/applyconfiguration/leaderworkerset/v1"

	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	"inftyai.com/llmaz/pkg"
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
	model := &coreapi.Model{}
	// TODO: multiModelsClaim
	modelName := service.Spec.MultiModelsClaims[0].ModelNames[0]
	if err := r.Get(ctx, types.NamespacedName{Name: string(modelName)}, model); err != nil {
		return ctrl.Result{}, err
	}

	workloadApplyConfiguration := buildWorkloadApplyConfiguration(service, model)
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
		Watches(&lws.LeaderWorkerSet{}, &handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldBar := e.ObjectOld.(*lws.LeaderWorkerSet)
					newBar := e.ObjectNew.(*lws.LeaderWorkerSet)
					return !reflect.DeepEqual(oldBar.Status, newBar.Status)
				},
			})).
		Complete(r)
}

func buildWorkloadApplyConfiguration(service *inferenceapi.Service, model *coreapi.Model) *applyconfigurationv1.LeaderWorkerSetApplyConfiguration {
	workload := applyconfigurationv1.LeaderWorkerSet(service.Name, service.Namespace)

	leaderWorkerTemplate := applyconfigurationv1.LeaderWorkerTemplate()
	leaderWorkerTemplate.WithWorkerTemplate(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate)
	injectModelRequirements(leaderWorkerTemplate, model)

	spec := applyconfigurationv1.LeaderWorkerSetSpec()
	spec.WithLeaderWorkerTemplate(leaderWorkerTemplate)
	spec.WithReplicas(*service.Spec.WorkloadTemplate.Replicas)

	workload.WithSpec(spec)
	return workload
}

func injectModelRequirements(template *applyconfigurationv1.LeaderWorkerTemplateApplyConfiguration, model *coreapi.Model) {
	template.WorkerTemplate.Labels = modelLabels(model)
	template.WorkerTemplate.Annotations = modelAnnotations(model)
	injectModelLoader(template)
	injectModelFlavor(template, model)
}

func injectModelLoader(template *applyconfigurationv1.LeaderWorkerTemplateApplyConfiguration) {
	template.WorkerTemplate.Spec.InitContainers = []corev1.Container{
		{
			Name:  pkg.MODEL_LOADER_CONTAINER_NAME,
			Image: pkg.LOADER_IMAGE,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      pkg.MODEL_CACHE_VOLUME_NAME,
					MountPath: pkg.CONTAINER_MODEL_PATH,
				},
			},
		},
	}

	if template.WorkerTemplate.Annotations != nil && template.WorkerTemplate.Annotations[coreapi.ModelHubLabelKey] != "" {
		template.WorkerTemplate.Spec.InitContainers[0].Env = append(
			template.WorkerTemplate.Spec.InitContainers[0].Env,
			corev1.EnvVar{Name: "MODEL_ID", Value: template.WorkerTemplate.Annotations[coreapi.ModelIDLabelKey]},
			corev1.EnvVar{Name: "MODEL_HUB_NAME", Value: template.WorkerTemplate.Annotations[coreapi.ModelHubLabelKey]},
		)
	}
}

func injectModelFlavor(template *applyconfigurationv1.LeaderWorkerTemplateApplyConfiguration, model *coreapi.Model) {
	if len(model.Spec.InferenceFlavors) == 0 {
		return
	}

	// Let's handle the 0-index flavor for the model first.
	// TODO: fungibility support.
	requests := model.Spec.InferenceFlavors[0].Requests
	for k, v := range requests {
		if template.WorkerTemplate.Spec.Containers[0].Resources.Requests == nil {
			template.WorkerTemplate.Spec.Containers[0].Resources.Requests = map[corev1.ResourceName]resource.Quantity{}
		}
		template.WorkerTemplate.Spec.Containers[0].Resources.Requests[k] = v

		if template.WorkerTemplate.Spec.Containers[0].Resources.Limits == nil {
			template.WorkerTemplate.Spec.Containers[0].Resources.Limits = map[corev1.ResourceName]resource.Quantity{}
		}
		template.WorkerTemplate.Spec.Containers[0].Resources.Limits[k] = v
	}

	nodeSelector := model.Spec.InferenceFlavors[0].NodeSelector
	if len(nodeSelector) > 0 {
		template.WorkerTemplate.Spec.Affinity = &corev1.Affinity{
			NodeAffinity: &corev1.NodeAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{},
			},
		}

		term := corev1.NodeSelectorTerm{}
		for k, v := range nodeSelector {
			term.MatchExpressions = append(term.MatchExpressions,
				corev1.NodeSelectorRequirement{
					Key:      k,
					Values:   []string{v},
					Operator: corev1.NodeSelectorOpIn,
				})
		}
		template.WorkerTemplate.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{term}
	}

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

func modelLabels(model *coreapi.Model) map[string]string {
	return map[string]string{
		coreapi.ModelNameLabelKey:       string(model.Name),
		coreapi.ModelFamilyNameLabelKey: string(model.Spec.FamilyName),
	}
}

func modelAnnotations(model *coreapi.Model) map[string]string {
	if model.Spec.DataSource.ModelHub != nil {
		return map[string]string{
			coreapi.ModelIDLabelKey:  string(model.Spec.DataSource.ModelHub.ModelID),
			coreapi.ModelHubLabelKey: string(*model.Spec.DataSource.ModelHub.Name),
		}
	}
	return nil
}
