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
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	metaapplyv1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
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

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	helper "github.com/inftyai/llmaz/pkg/controller_helper"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/modelsource"
	"github.com/inftyai/llmaz/pkg/util"
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
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ServiceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	service := &inferenceapi.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, service); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(10).Info("reconcile Service", "Service", klog.KObj(service))

	models, err := helper.FetchModelsByService(ctx, r.Client, service)
	if err != nil {
		return ctrl.Result{}, err
	}

	workloadApplyConfiguration := buildWorkloadApplyConfiguration(service, models)
	if err := setControllerReferenceForWorkload(service, workloadApplyConfiguration, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	// TODO: handle fungibility

	if err := util.Patch(ctx, r.Client, workloadApplyConfiguration); err != nil {
		return ctrl.Result{}, err
	}

	// Create a service for the leader pods of the lws for loadbalancing.
	if err := CreateServiceIfNotExists(ctx, r.Client, r.Scheme, service); err != nil {
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
		Watches(&lws.LeaderWorkerSet{}, handler.EnqueueRequestForOwner(r.Scheme, r.RESTMapper(), &inferenceapi.Service{}, handler.OnlyControllerOwner()),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldBar := e.ObjectOld.(*lws.LeaderWorkerSet)
					newBar := e.ObjectNew.(*lws.LeaderWorkerSet)
					return !reflect.DeepEqual(oldBar.Status, newBar.Status)
				},
			})).
		Complete(r)
}

func buildWorkloadApplyConfiguration(service *inferenceapi.Service, models []*coreapi.OpenModel) *applyconfigurationv1.LeaderWorkerSetApplyConfiguration {
	workload := applyconfigurationv1.LeaderWorkerSet(service.Name, service.Namespace)

	leaderWorkerTemplate := applyconfigurationv1.LeaderWorkerTemplate()
	if service.Spec.WorkloadTemplate.LeaderTemplate != nil {
		leaderWorkerTemplate.WithLeaderTemplate(*service.Spec.WorkloadTemplate.LeaderTemplate)
	}
	leaderWorkerTemplate.WithWorkerTemplate(service.Spec.WorkloadTemplate.WorkerTemplate)

	// The core logic to inject additional configurations.
	injectModelProperties(leaderWorkerTemplate, models, service)

	spec := applyconfigurationv1.LeaderWorkerSetSpec()
	spec.WithLeaderWorkerTemplate(leaderWorkerTemplate)
	spec.LeaderWorkerTemplate.WithSize(*service.Spec.WorkloadTemplate.Size)
	spec.WithReplicas(*service.Spec.Replicas)
	if service.Spec.RolloutStrategy != nil {
		spec.WithRolloutStrategy(applyconfigurationv1.RolloutStrategy().WithType(service.Spec.RolloutStrategy.Type))
		if service.Spec.RolloutStrategy.RollingUpdateConfiguration != nil {
			spec.RolloutStrategy.WithRollingUpdateConfiguration(
				applyconfigurationv1.RollingUpdateConfiguration().
					WithMaxSurge(service.Spec.RolloutStrategy.RollingUpdateConfiguration.MaxSurge).
					WithMaxUnavailable(service.Spec.RolloutStrategy.RollingUpdateConfiguration.MaxUnavailable),
			)
		}
	}
	spec.WithStartupPolicy(lws.LeaderReadyStartupPolicy)

	workload.WithSpec(spec)
	return workload
}

func injectModelProperties(template *applyconfigurationv1.LeaderWorkerTemplateApplyConfiguration, models []*coreapi.OpenModel, service *inferenceapi.Service) {
	isMultiNodesInference := template.LeaderTemplate != nil

	// Skip model loader if llmaz.io/skip-model-loader annotation is set.
	skipModelLoader := false
	if annotations := service.GetAnnotations(); annotations != nil {
		skipModelLoader = annotations[inferenceapi.SkipModelLoaderAnnoKey] == "true"
	}

	for i, model := range models {
		source := modelSource.NewModelSourceProvider(model)

		if !skipModelLoader {
			if isMultiNodesInference {
				source.InjectModelLoader(template.LeaderTemplate, i)
			}
			source.InjectModelLoader(template.WorkerTemplate, i)
		} else {
			if isMultiNodesInference {
				source.InjectModelEnvVars(template.LeaderTemplate)
			}
			source.InjectModelEnvVars(template.WorkerTemplate)
		}
	}

	// We only consider the main model's requirements for now.
	if isMultiNodesInference {
		template.LeaderTemplate.Labels = util.MergeKVs(template.LeaderTemplate.Labels, modelLabels(models[0]))
		template.LeaderTemplate.Annotations = util.MergeKVs(template.LeaderTemplate.Annotations, modelAnnotations(service))
	} else {
		template.WorkerTemplate.Labels = util.MergeKVs(template.WorkerTemplate.Labels, modelLabels(models[0]))
		template.WorkerTemplate.Annotations = util.MergeKVs(template.WorkerTemplate.Annotations, modelAnnotations(service))
	}

	// Consider main model only.
	injectModelFlavor(template.WorkerTemplate, models[0], service)
	if isMultiNodesInference {
		injectModelFlavor(template.LeaderTemplate, models[0], service)
	}
}

func injectModelFlavor(template *corev1.PodTemplateSpec, model *coreapi.OpenModel, service *inferenceapi.Service) {
	if model.Spec.InferenceConfig == nil || len(model.Spec.InferenceConfig.Flavors) == 0 {
		return
	}

	container := &corev1.Container{}
	for i, c := range template.Spec.Containers {
		if c.Name == modelSource.MODEL_RUNNER_CONTAINER_NAME {
			container = &template.Spec.Containers[i]
		}
	}

	flavorName := model.Spec.InferenceConfig.Flavors[0].Name
	if len(service.Spec.ModelClaims.InferenceFlavors) > 0 {
		// We only support the same resource request right now, so 0-index flavor is enough.
		flavorName = service.Spec.ModelClaims.InferenceFlavors[0]
	}

	for i, flavor := range model.Spec.InferenceConfig.Flavors {
		if flavor.Name == flavorName {
			limits := model.Spec.InferenceConfig.Flavors[i].Limits
			for k, v := range limits {
				if container.Resources.Requests == nil {
					container.Resources.Requests = map[corev1.ResourceName]resource.Quantity{}
				}
				// overwrite the requests and limits.
				container.Resources.Requests[k] = v

				if container.Resources.Limits == nil {
					container.Resources.Limits = map[corev1.ResourceName]resource.Quantity{}
				}
				// overwrite the requests and limits.
				container.Resources.Limits[k] = v
			}
			break
		}
	}
}

func modelLabels(model *coreapi.OpenModel) map[string]string {
	return map[string]string{
		coreapi.ModelNameLabelKey:       model.Name,
		coreapi.ModelFamilyNameLabelKey: string(model.Spec.FamilyName),
	}
}

func modelAnnotations(service *inferenceapi.Service) map[string]string {
	var values string
	for i, value := range service.Spec.ModelClaims.InferenceFlavors {
		if i == len(service.Spec.ModelClaims.InferenceFlavors)-1 {
			values += string(value)
		} else {
			values += string(value) + ","
		}
	}

	if len(values) > 0 {
		return map[string]string{
			inferenceapi.InferenceServiceFlavorsAnnoKey: values,
		}
	}
	return nil
}

func setServiceCondition(service *inferenceapi.Service, workload *lws.LeaderWorkerSet) {
	defer func() {
		if service.Status.Selector != workload.Status.HPAPodSelector {
			service.Status.Selector = workload.Status.HPAPodSelector
		}
		if service.Status.Replicas != workload.Status.Replicas {
			service.Status.Replicas = workload.Status.Replicas
		}
	}()

	if apimeta.IsStatusConditionTrue(workload.Status.Conditions, string(lws.LeaderWorkerSetAvailable)) {
		condition := metav1.Condition{
			Type:    inferenceapi.ServiceAvailable,
			Status:  metav1.ConditionTrue,
			Reason:  "ServiceReady",
			Message: "Inference Service is ready",
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, condition)
	} else {
		condition := metav1.Condition{
			Type:    inferenceapi.ServiceProgressing,
			Status:  metav1.ConditionTrue,
			Reason:  "ServiceInProgress",
			Message: "Inference Service is progressing",
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
			Message: "Waiting for leaderWorkerSet ready",
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, new_condition)
	}
}

// setControllerReferenceForWorkload set service as the owner reference for the workload.
func setControllerReferenceForWorkload(owner metav1.Object, lws *applyconfigurationv1.LeaderWorkerSetApplyConfiguration, scheme *runtime.Scheme) error {
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

func CreateServiceIfNotExists(ctx context.Context, k8sClient client.Client, Scheme *runtime.Scheme, service *inferenceapi.Service) error {
	log := ctrl.LoggerFrom(ctx)
	// The load balancing service name.
	svcName := service.Name + "-lb"

	var svc corev1.Service
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: svcName, Namespace: service.Namespace}, &svc); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		svc = corev1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svcName,
				Namespace: service.Namespace,
			},
			Spec: corev1.ServiceSpec{
				Ports: []corev1.ServicePort{
					{
						Name:       "http",
						Protocol:   corev1.ProtocolTCP,
						Port:       modelSource.DEFAULT_BACKEND_PORT,
						TargetPort: intstr.FromInt(modelSource.DEFAULT_BACKEND_PORT),
					},
				},
				Selector: map[string]string{
					lws.SetNameLabelKey: service.Name,
					// the leader pod.
					lws.WorkerIndexLabelKey: "0",
				},
			},
		}

		// Set the controller owner reference for garbage collection and reconciliation.
		if err := ctrl.SetControllerReference(service, &svc, Scheme); err != nil {
			return err
		}
		// create the service in the cluster
		log.V(2).Info("Creating service.")
		if err := k8sClient.Create(ctx, &svc); err != nil {
			return err
		}
	}
	return nil
}
