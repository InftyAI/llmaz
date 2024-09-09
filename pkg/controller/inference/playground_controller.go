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
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
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

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	coreclientgo "github.com/inftyai/llmaz/client-go/applyconfiguration/core/v1alpha1"
	inferenceclientgo "github.com/inftyai/llmaz/client-go/applyconfiguration/inference/v1alpha1"
	"github.com/inftyai/llmaz/pkg/controller_helper/backend"
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

	var serviceApplyConfiguration *inferenceclientgo.ServiceApplyConfiguration

	models := []*coreapi.OpenModel{}
	if playground.Spec.ModelClaim != nil {
		model := &coreapi.OpenModel{}
		if err := r.Get(ctx, types.NamespacedName{Name: string(playground.Spec.ModelClaim.ModelName)}, model); err != nil {
			if apierrors.IsNotFound(err) && handleUnexpectedCondition(playground, false, false) {
				return ctrl.Result{}, r.Client.Status().Update(ctx, playground)
			}
			return ctrl.Result{}, err
		}
		models = append(models, model)
	} else if playground.Spec.ModelClaims != nil {
		for _, mr := range playground.Spec.ModelClaims.Models {
			model := &coreapi.OpenModel{}
			if err := r.Get(ctx, types.NamespacedName{Name: string(mr.Name)}, model); err != nil {
				if apierrors.IsNotFound(err) && handleUnexpectedCondition(playground, false, false) {
					return ctrl.Result{}, r.Client.Status().Update(ctx, playground)
				}
				return ctrl.Result{}, err
			}
			models = append(models, model)
		}
	}

	serviceApplyConfiguration = buildServiceApplyConfiguration(models, playground)

	if err := setControllerReferenceForService(playground, serviceApplyConfiguration, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if err := util.Patch(ctx, r.Client, serviceApplyConfiguration); err != nil {
		return ctrl.Result{}, err
	}

	// Handle status.
	setPlaygroundCondition(playground, service)
	if err := r.Client.Status().Update(ctx, playground); err != nil {
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

		// TODO: handle MultiModelsClaims in the future.

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

func buildServiceApplyConfiguration(models []*coreapi.OpenModel, playground *inferenceapi.Playground) *inferenceclientgo.ServiceApplyConfiguration {
	// Build metadata
	serviceApplyConfiguration := inferenceclientgo.Service(playground.Name, playground.Namespace)

	// Build spec.
	spec := inferenceclientgo.ServiceSpec()

	claim := &coreclientgo.ModelClaimsApplyConfiguration{}
	if playground.Spec.ModelClaim != nil {
		claim = coreclientgo.ModelClaims().
			WithModels(coreclientgo.ModelRepresentative().WithName(playground.Spec.ModelClaim.ModelName).WithRole(coreapi.MainRole)).
			WithInferenceFlavors(playground.Spec.ModelClaim.InferenceFlavors...)
	} else if playground.Spec.ModelClaims != nil {
		mrs := []*coreclientgo.ModelRepresentativeApplyConfiguration{}
		for _, model := range playground.Spec.ModelClaims.Models {
			role := coreapi.MainRole
			if model.Role != nil {
				role = *model.Role
			}
			mr := coreclientgo.ModelRepresentative().WithName(model.Name).WithRole(role)
			mrs = append(mrs, mr)
		}

		claim = coreclientgo.ModelClaims().
			WithModels(mrs...).
			WithInferenceFlavors(playground.Spec.ModelClaims.InferenceFlavors...)
	}

	spec.WithModelClaims(claim)
	spec.WithWorkloadTemplate(buildWorkloadTemplate(models, playground))
	serviceApplyConfiguration.WithSpec(spec)

	return serviceApplyConfiguration

	// TODO: handle MultiModelsClaims in the future.
}

// We do not want to maintain another workload like deployment for single-host cases so we choose lws here
// to cover both single-host and multi-host cases. There're some shortages for lws like can not force rolling
// update when one replica failed, we'll fix this in the kubernetes upstream.
// Model flavors will not be considered but in inferenceService controller to support accelerator fungibility.
func buildWorkloadTemplate(models []*coreapi.OpenModel, playground *inferenceapi.Playground) lws.LeaderWorkerSetSpec {
	// TODO: this should be leaderWorkerSetTemplateSpec, we should support in the lws upstream.
	workload := lws.LeaderWorkerSetSpec{
		// Use the default policy defined in lws.
		StartupPolicy: lws.LeaderCreatedStartupPolicy,
		RolloutStrategy: lws.RolloutStrategy{
			Type: lws.RollingUpdateStrategyType,
		},
	}

	workload.Replicas = playground.Spec.Replicas

	// TODO: handle multi-host scenarios, e.g. nvidia.com/gpu: 32, means we'll split into 4 hosts.
	// Do we need another configuration for playground for multi-host use case? I guess no currently.
	workload.LeaderWorkerTemplate.WorkerTemplate = buildWorkerTemplate(models, playground)

	return workload
}

func involveRole(playground *inferenceapi.Playground) coreapi.ModelRole {
	if playground.Spec.ModelClaim != nil {
		return coreapi.MainRole
	} else if playground.Spec.ModelClaims != nil {
		for _, mr := range playground.Spec.ModelClaims.Models {
			if *mr.Role != coreapi.MainRole {
				return *mr.Role
			}
		}
	}

	return coreapi.MainRole
}

func buildWorkerTemplate(models []*coreapi.OpenModel, playground *inferenceapi.Playground) corev1.PodTemplateSpec {
	backendName := inferenceapi.DefaultBackend
	if playground.Spec.BackendConfig != nil && playground.Spec.BackendConfig.Name != nil {
		backendName = *playground.Spec.BackendConfig.Name
	}
	bkd := backend.SwitchBackend(backendName)

	version := bkd.DefaultVersion()
	if playground.Spec.BackendConfig != nil && playground.Spec.BackendConfig.Version != nil {
		version = *playground.Spec.BackendConfig.Version
	}

	args := bkd.Args(models, involveRole(playground))

	var envs []corev1.EnvVar
	if playground.Spec.BackendConfig != nil {
		args = append(args, playground.Spec.BackendConfig.Args...)
		envs = playground.Spec.BackendConfig.Envs
	}

	resources := corev1.ResourceRequirements{
		Limits:   bkd.DefaultResources().Limits,
		Requests: bkd.DefaultResources().Requests,
	}
	if playground.Spec.BackendConfig != nil && playground.Spec.BackendConfig.Resources != nil {
		limits := util.MergeResources(playground.Spec.BackendConfig.Resources.Limits, resources.Limits)
		requests := util.MergeResources(playground.Spec.BackendConfig.Resources.Requests, resources.Requests)

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

	template := corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			// TODO: should we support image pull secret here?
			// TODO: support readiness/liveness
			Containers: []corev1.Container{
				{
					Name:      modelSource.MODEL_RUNNER_CONTAINER_NAME,
					Image:     bkd.Image(version),
					Resources: resources,
					Command:   bkd.DefaultCommand(),
					Args:      args,
					Env:       envs,
					Ports: []corev1.ContainerPort{
						{
							Name:          "http",
							Protocol:      corev1.ProtocolTCP,
							ContainerPort: modelSource.DEFAULT_BACKEND_PORT,
						},
					},
				},
			},
		},
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
