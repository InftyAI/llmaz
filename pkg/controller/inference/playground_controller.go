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

	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	metaapplyv1 "k8s.io/client-go/applyconfigurations/meta/v1"
	"k8s.io/client-go/tools/record"
	"k8s.io/klog/v2"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	coreclientgo "github.com/inftyai/llmaz/client-go/applyconfiguration/core/v1alpha1"
	inferenceclientgo "github.com/inftyai/llmaz/client-go/applyconfiguration/inference/v1alpha1"
	helper "github.com/inftyai/llmaz/pkg/controller_helper"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/model_source"
	"github.com/inftyai/llmaz/pkg/util"
)

// PlaygroundReconciler reconciles a Playground object
type PlaygroundReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Record record.EventRecorder
}

func NewPlaygroundReconciler(client client.Client, scheme *runtime.Scheme, record record.EventRecorder) *PlaygroundReconciler {
	return &PlaygroundReconciler{
		Client: client,
		Scheme: scheme,
		Record: record,
	}
}

//+kubebuilder:rbac:groups=inference.llmaz.io,resources=playgrounds,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=inference.llmaz.io,resources=playgrounds/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=inference.llmaz.io,resources=playgrounds/finalizers,verbs=update
//+kubebuilder:rbac:groups=autoscaling,resources=horizontalpodautoscalers,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *PlaygroundReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithName("Playground")

	playground := &inferenceapi.Playground{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, playground); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.V(10).Info("reconcile Playground", "Playground", klog.KObj(playground))

	service := &inferenceapi.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service); err == nil {
		if !metav1.IsControlledBy(service, playground) {
			logger.Info("failed to construct inference Service as a Service with the same exists", "Playground", klog.KObj(playground))
			if changed := handleUnexpectedCondition(playground, true, true); changed {
				err = r.Client.Status().Update(ctx, playground)
			}
			// if update successfully, err will be nil and we'll hanging here until Playground or Service deleted.
			return ctrl.Result{}, err
		}
	}

	models, err := helper.FetchModelsByPlayground(ctx, r.Client, playground)
	if err != nil {
		if apierrors.IsNotFound(err) && handleUnexpectedCondition(playground, false, false) {
			return ctrl.Result{}, r.Client.Status().Update(ctx, playground)
		}
		return ctrl.Result{}, err
	}

	backendRuntimeName := inferenceapi.DefaultBackend
	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Name != nil {
		backendRuntimeName = *playground.Spec.BackendRuntimeConfig.Name
	}
	backendRuntime := &inferenceapi.BackendRuntime{}
	if err := r.Get(ctx, types.NamespacedName{Name: string(backendRuntimeName)}, backendRuntime); err != nil {
		logger.Error(err, "failed to get backendRuntime", "BackendRuntime", backendRuntimeName)
		return ctrl.Result{}, err
	}

	serviceApplyConfiguration, err := buildServiceApplyConfiguration(models, playground, backendRuntime)
	if err != nil {
		logger.Error(err, "failed to build inference Service")
		return ctrl.Result{}, err
	}
	if err := setControllerReferenceForService(playground, serviceApplyConfiguration, r.Scheme); err != nil {
		logger.Error(err, "failed to set OwnerReference for Service", "Service", fmt.Sprintf("%s/%s", playground.Namespace, playground.Name))
		return ctrl.Result{}, err
	}
	if err := util.Patch(ctx, r.Client, serviceApplyConfiguration); err != nil {
		logger.Error(err, "failed to patch Service", "Service", fmt.Sprintf("%s/%s", playground.Namespace, playground.Name))
		return ctrl.Result{}, err
	}

	scalingConfiguration := buildScalingConfiguration(playground, backendRuntime)
	if scalingConfiguration != nil {
		if err := setControllerReferenceForScalingConfiguration(playground, scalingConfiguration, r.Scheme); err != nil {
			logger.Error(err, "failed to set OwnerReference for scaling workload", "workload", fmt.Sprintf("%s/%s", playground.Namespace, playground.Name), "kind", scalingConfiguration.Kind)
			return ctrl.Result{}, err
		}
		if err := util.Patch(ctx, r.Client, scalingConfiguration); err != nil {
			logger.Error(err, "failed to patch scaling workload", "workload", fmt.Sprintf("%s/%s", playground.Namespace, playground.Name), "kind", scalingConfiguration.Kind)
			return ctrl.Result{}, err
		}
	}

	// Handle status.
	setPlaygroundCondition(playground, service)
	if err := r.Client.Status().Update(ctx, playground); err != nil {
		logger.Error(err, "failed to update Playground status", "Playground", klog.KObj(playground))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlaygroundReconciler) SetupWithManager(mgr ctrl.Manager) error {
	mapFn := func(ctx context.Context, obj client.Object) []ctrl.Request {
		logger := log.FromContext(ctx)

		modelName := obj.GetName()

		playgrounds := &inferenceapi.PlaygroundList{}
		err := r.List(ctx, playgrounds, client.MatchingLabels{coreapi.ModelNameLabelKey: modelName})
		if err != nil {
			logger.Error(err, "failed to list playgrounds")
			return nil
		}

		var reqs []ctrl.Request
		for _, playground := range playgrounds.Items {
			reqs = append(reqs, ctrl.Request{
				NamespacedName: types.NamespacedName{
					Name: playground.Name,
				},
			})
		}

		return reqs
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&inferenceapi.Playground{}).
		Watches(&inferenceapi.Service{}, &handler.EnqueueRequestForObject{},
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc: func(e event.UpdateEvent) bool {
					oldBar := e.ObjectOld.(*inferenceapi.Service)
					newBar := e.ObjectNew.(*inferenceapi.Service)
					return !reflect.DeepEqual(oldBar.Status, newBar.Status)
				},
			})).
		Watches(&coreapi.OpenModel{}, handler.EnqueueRequestsFromMapFunc(mapFn),
			builder.WithPredicates(predicate.Funcs{
				UpdateFunc:  func(e event.UpdateEvent) bool { return false },
				DeleteFunc:  func(e event.DeleteEvent) bool { return false },
				GenericFunc: func(e event.GenericEvent) bool { return false },
			})).
		Complete(r)
}

func buildServiceApplyConfiguration(models []*coreapi.OpenModel, playground *inferenceapi.Playground, backendRuntime *inferenceapi.BackendRuntime) (*inferenceclientgo.ServiceApplyConfiguration, error) {
	// Build metadata
	serviceApplyConfiguration := inferenceclientgo.Service(playground.Name, playground.Namespace)

	// Build spec.
	spec := inferenceclientgo.ServiceSpec()

	var claim *coreclientgo.ModelClaimsApplyConfiguration
	if playground.Spec.ModelClaim != nil {
		claim = coreclientgo.ModelClaims().
			WithModels(coreclientgo.ModelRef().WithName(playground.Spec.ModelClaim.ModelName).WithRole(coreapi.MainRole)).
			WithInferenceFlavors(playground.Spec.ModelClaim.InferenceFlavors...)
	} else {
		mrs := []*coreclientgo.ModelRefApplyConfiguration{}
		for _, model := range playground.Spec.ModelClaims.Models {
			role := coreapi.MainRole
			if model.Role != nil {
				role = *model.Role
			}
			mr := coreclientgo.ModelRef().WithName(model.Name).WithRole(role)
			mrs = append(mrs, mr)
		}

		claim = coreclientgo.ModelClaims().
			WithModels(mrs...).
			WithInferenceFlavors(playground.Spec.ModelClaims.InferenceFlavors...)
	}

	spec.WithModelClaims(claim)
	template, err := buildWorkloadTemplate(models, playground, backendRuntime)
	if err != nil {
		return nil, err
	}

	spec.WithWorkloadTemplate(template)
	serviceApplyConfiguration.WithSpec(spec)

	return serviceApplyConfiguration, nil

	// TODO: handle MultiModelsClaims in the future.
}

// We do not want to maintain another workload like deployment for single-host cases so we choose lws here
// to cover both single-host and multi-host cases. There're some shortages for lws like can not force rolling
// update when one replica failed, we'll fix this in the kubernetes upstream.
// Model flavors will not be considered but in inferenceService controller to support accelerator fungibility.
func buildWorkloadTemplate(models []*coreapi.OpenModel, playground *inferenceapi.Playground, backendRuntime *inferenceapi.BackendRuntime) (lws.LeaderWorkerSetSpec, error) {
	workload := lws.LeaderWorkerSetSpec{
		// Use the default policy defined in lws.
		StartupPolicy: lws.LeaderCreatedStartupPolicy,
		RolloutStrategy: lws.RolloutStrategy{
			Type: lws.RollingUpdateStrategyType,
		},
	}

	workload.Replicas = playground.Spec.Replicas

	nodeSize, multiHost := helper.MultiHostInference(models[0], playground)
	if multiHost {
		workload.LeaderWorkerTemplate.Size = &nodeSize
	}

	template, err := buildTemplate(models, playground, backendRuntime, multiHost)
	if err != nil {
		return lws.LeaderWorkerSetSpec{}, err
	}

	if multiHost {
		workload.LeaderWorkerTemplate.LeaderTemplate = &template
		workload.LeaderWorkerTemplate.WorkerTemplate = buildWorkerTemplate(models, playground, backendRuntime)
	} else {
		workload.LeaderWorkerTemplate.WorkerTemplate = template
	}

	return workload, nil
}

func buildTemplate(models []*coreapi.OpenModel, playground *inferenceapi.Playground, backendRuntime *inferenceapi.BackendRuntime, multiHost bool) (corev1.PodTemplateSpec, error) {
	parser := helper.NewBackendRuntimeParser(backendRuntime)

	commands := parser.Commands()
	if multiHost {
		commands = parser.LeaderCommands()
	}

	args, err := parser.Args(playground, models, multiHost)
	if err != nil {
		return corev1.PodTemplateSpec{}, err
	}
	envs := parser.Envs()

	if playground.Spec.BackendRuntimeConfig != nil {
		envs = append(envs, playground.Spec.BackendRuntimeConfig.Envs...)
		if playground.Spec.BackendRuntimeConfig.Args != nil {
			args = append(args, playground.Spec.BackendRuntimeConfig.Args.Flags...)
		}
	}

	resources := corev1.ResourceRequirements{
		Requests: parser.Resources().Requests,
		Limits:   parser.Resources().Limits,
	}
	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Resources != nil {
		limits := util.MergeResources(playground.Spec.BackendRuntimeConfig.Resources.Limits, parser.Resources().Limits)
		requests := util.MergeResources(playground.Spec.BackendRuntimeConfig.Resources.Requests, parser.Resources().Requests)

		resources = corev1.ResourceRequirements{
			Limits:   limits,
			Requests: requests,
		}

		// Make sure the limits are always greater than requests.
		for k, v := range resources.Limits {
			if k == corev1.ResourceCPU || k == corev1.ResourceMemory {
				if v.Cmp(requests[k]) == -1 {
					resources.Limits[k] = requests[k]
				}
			}
		}
	}

	version := parser.Version()
	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Version != nil {
		version = *playground.Spec.BackendRuntimeConfig.Version
	}

	// Pod can not accept shell commands with args together, merge the args with the commands.
	if multiHost {
		commands = util.MergeArgsWithCommands(commands, args)
		args = nil
	}

	var livenessProbe, readinessProbe, startupProbe *corev1.Probe
	if backendRuntime.Spec.StartupProbe != nil {
		startupProbe = backendRuntime.Spec.StartupProbe
	}
	if backendRuntime.Spec.LivenessProbe != nil {
		livenessProbe = backendRuntime.Spec.LivenessProbe
	}
	if backendRuntime.Spec.ReadinessProbe != nil {
		readinessProbe = backendRuntime.Spec.ReadinessProbe
	}

	template := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			// TODO: should we support image pull secret here?
			// TODO: support readiness/liveness
			Containers: []corev1.Container{
				{
					Name:      modelSource.MODEL_RUNNER_CONTAINER_NAME,
					Image:     parser.Image(version),
					Resources: resources,
					Command:   commands,
					Args:      args,
					Env:       envs,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							Protocol:      corev1.ProtocolTCP,
							ContainerPort: modelSource.DEFAULT_BACKEND_PORT,
						},
					},
					StartupProbe:   startupProbe,
					LivenessProbe:  livenessProbe,
					ReadinessProbe: readinessProbe,
				},
			},
		},
	}

	// construct /dev/shm size
	if models[0].Spec.InferenceConfig != nil && models[0].Spec.InferenceConfig.SharedMemorySize != nil {
		template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMediumMemory,
					SizeLimit: models[0].Spec.InferenceConfig.SharedMemorySize,
				},
			},
		})

		template.Spec.Containers[0].VolumeMounts = append(template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "dshm",
			MountPath: "/dev/shm",
		})
	}

	return template, nil
}

// This is a copy of buildTemplate with some refactors, only used in multi-nodes cases.
// Worker template has no args, no contain port.
func buildWorkerTemplate(models []*coreapi.OpenModel, playground *inferenceapi.Playground, backendRuntime *inferenceapi.BackendRuntime) corev1.PodTemplateSpec {
	parser := helper.NewBackendRuntimeParser(backendRuntime)

	envs := parser.Envs()
	if playground.Spec.BackendRuntimeConfig != nil {
		envs = append(envs, playground.Spec.BackendRuntimeConfig.Envs...)
	}

	resources := corev1.ResourceRequirements{
		Requests: parser.Resources().Requests,
		Limits:   parser.Resources().Limits,
	}
	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Resources != nil {
		limits := util.MergeResources(playground.Spec.BackendRuntimeConfig.Resources.Limits, parser.Resources().Limits)
		requests := util.MergeResources(playground.Spec.BackendRuntimeConfig.Resources.Requests, parser.Resources().Requests)

		resources = corev1.ResourceRequirements{
			Limits:   limits,
			Requests: requests,
		}

		// Make sure the limits are always greater than requests.
		for k, v := range resources.Limits {
			if k == corev1.ResourceCPU || k == corev1.ResourceMemory {
				if v.Cmp(requests[k]) == -1 {
					resources.Limits[k] = requests[k]
				}
			}
		}
	}

	version := parser.Version()
	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Version != nil {
		version = *playground.Spec.BackendRuntimeConfig.Version
	}

	template := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			// TODO: should we support image pull secret here?
			// TODO: support readiness/liveness
			Containers: []corev1.Container{
				{
					Name:      modelSource.MODEL_RUNNER_CONTAINER_NAME,
					Image:     parser.Image(version),
					Resources: resources,
					Command:   parser.WorkerCommands(),
					Env:       envs,
				},
			},
		},
	}

	// construct /dev/shm size
	if models[0].Spec.InferenceConfig != nil && models[0].Spec.InferenceConfig.SharedMemorySize != nil {
		template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
			Name: "dshm",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{
					Medium:    corev1.StorageMediumMemory,
					SizeLimit: models[0].Spec.InferenceConfig.SharedMemorySize,
				},
			},
		})

		template.Spec.Containers[0].VolumeMounts = append(template.Spec.Containers[0].VolumeMounts, corev1.VolumeMount{
			Name:      "dshm",
			MountPath: "/dev/shm",
		})
	}

	return template
}

func handleUnexpectedCondition(playground *inferenceapi.Playground, modelExists bool, serviceWithSameNameExists bool) (changed bool) {
	// Put it in the first place as more serious.
	if serviceWithSameNameExists {
		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundProgressing,
			Status:  metav1.ConditionFalse,
			Reason:  "AbortProcessing",
			Message: "Playground owns the same name with an existing Service",
		}
		return apimeta.SetStatusCondition(&playground.Status.Conditions, condition)
	}

	if !modelExists {
		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundProgressing,
			Status:  metav1.ConditionFalse,
			Reason:  "AbortProcessing",
			Message: "Waiting for model creation",
		}
		return apimeta.SetStatusCondition(&playground.Status.Conditions, condition)
	}
	return false
}

func setPlaygroundCondition(playground *inferenceapi.Playground, service *inferenceapi.Service) (changed bool) {
	defer func() {
		if playground.Status.Selector != service.Status.Selector {
			playground.Status.Selector = service.Status.Selector
			changed = true
		}
		if playground.Status.Replicas != service.Status.Replicas {
			playground.Status.Replicas = service.Status.Replicas
			changed = true
		}
	}()

	// For the start up or Playground is recovered from AbortProcessing.
	if len(playground.Status.Conditions) == 0 || apimeta.IsStatusConditionFalse(playground.Status.Conditions, inferenceapi.PlaygroundProgressing) {
		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundProgressing,
			Status:  metav1.ConditionTrue,
			Reason:  "Pending",
			Message: "Waiting for inference Service ready",
		}
		return apimeta.SetStatusCondition(&playground.Status.Conditions, condition)
	}

	if apimeta.IsStatusConditionTrue(service.Status.Conditions, inferenceapi.ServiceAvailable) {
		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundAvailable,
			Status:  metav1.ConditionTrue,
			Reason:  "PlaygroundReady",
			Message: "Playground is ready",
		}
		return apimeta.SetStatusCondition(&playground.Status.Conditions, condition)
	} else {
		// Still in starting up, no need to populate the condition.
		if apimeta.FindStatusCondition(playground.Status.Conditions, inferenceapi.PlaygroundAvailable) == nil {
			return false
		}

		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundProgressing,
			Status:  metav1.ConditionTrue,
			Reason:  "PlaygroundInProgress",
			Message: "Waiting for inference Service progressing",
		}
		changed = apimeta.SetStatusCondition(&playground.Status.Conditions, condition) || changed

		// Set the available to false
		new_condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundAvailable,
			Status:  metav1.ConditionFalse,
			Reason:  "PlaygroundNotReady",
			Message: "Waiting for inference Service ready",
		}
		changed = apimeta.SetStatusCondition(&playground.Status.Conditions, new_condition) || changed
		return changed
	}
}

// setControllerReferenceForService set playground as the owner reference for inferenceService.
func setControllerReferenceForService(owner metav1.Object, saf *inferenceclientgo.ServiceApplyConfiguration, scheme *runtime.Scheme) error {
	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("%T is not a runtime.Object, cannot call SetOwnerReference", owner)
	}
	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}
	saf.WithOwnerReferences(metaapplyv1.OwnerReference().
		WithAPIVersion(gvk.GroupVersion().String()).
		WithKind(gvk.Kind).
		WithName(owner.GetName()).
		WithUID(owner.GetUID()).
		WithBlockOwnerDeletion(true).
		WithController(true))
	return nil
}

// buildScalingConfiguration supports HPA only now.
func buildScalingConfiguration(playground *inferenceapi.Playground, backend *inferenceapi.BackendRuntime) *autoscalingv2.HorizontalPodAutoscaler {
	if playground.Spec.ElasticConfig == nil {
		return nil
	}

	// Handle HPA.
	if playground.Spec.ElasticConfig.ScaleTrigger != nil && playground.Spec.ElasticConfig.ScaleTrigger.HPA != nil {
		hpa := newHPA(playground)
		hpa.Spec.Metrics = playground.Spec.ElasticConfig.ScaleTrigger.HPA.Metrics
		hpa.Spec.Behavior = playground.Spec.ElasticConfig.ScaleTrigger.HPA.Behavior
		return hpa
	}

	if len(backend.Spec.ScaleTriggers) > 0 {
		hpa := newHPA(playground)
		if playground.Spec.ElasticConfig.ScaleTriggerRef != nil {
			for _, trigger := range backend.Spec.ScaleTriggers {
				if trigger.Name == playground.Spec.ElasticConfig.ScaleTriggerRef.Name {
					hpa.Spec.Metrics = trigger.HPA.Metrics
					hpa.Spec.Behavior = trigger.HPA.Behavior
					return hpa
				}
			}
		} else {
			// use the 0-index as the default value.
			hpa.Spec.Metrics = backend.Spec.ScaleTriggers[0].HPA.Metrics
			hpa.Spec.Behavior = backend.Spec.ScaleTriggers[0].HPA.Behavior
			return hpa
		}
	}

	return nil
}

func setControllerReferenceForScalingConfiguration(owner metav1.Object, hpa *autoscalingv2.HorizontalPodAutoscaler, scheme *runtime.Scheme) error {
	if hpa == nil {
		return nil
	}

	ro, ok := owner.(runtime.Object)
	if !ok {
		return fmt.Errorf("%T is not a runtime.Object, cannot call SetOwnerReference", owner)
	}
	gvk, err := apiutil.GVKForObject(ro, scheme)
	if err != nil {
		return err
	}
	hpa.OwnerReferences = []metav1.OwnerReference{
		{
			APIVersion:         gvk.GroupVersion().String(),
			Kind:               gvk.Kind,
			Name:               owner.GetName(),
			UID:                owner.GetUID(),
			BlockOwnerDeletion: ptr.To[bool](true),
			Controller:         ptr.To[bool](true),
		},
	}
	return nil
}

func newHPA(playground *inferenceapi.Playground) *autoscalingv2.HorizontalPodAutoscaler {
	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: autoscalingv2.SchemeGroupVersion.String(),
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      playground.Name,
			Namespace: playground.Namespace,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: inferenceapi.SchemeGroupVersion.String(),
				Kind:       "Playground",
				Name:       playground.Name,
			},
		},
	}

	hpa.Spec.MinReplicas = playground.Spec.ElasticConfig.MinReplicas
	if playground.Spec.ElasticConfig.MaxReplicas == nil {
		// The value is hardcoded, because maxReplicas is required by HPA.
		hpa.Spec.MaxReplicas = 99999
	} else {
		hpa.Spec.MaxReplicas = *playground.Spec.ElasticConfig.MaxReplicas
	}

	return hpa
}
