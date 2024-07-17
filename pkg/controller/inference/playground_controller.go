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

	corev1 "k8s.io/api/core/v1"
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

	core "inftyai.com/llmaz/api/core/v1alpha1"
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	coreclientgo "inftyai.com/llmaz/client-go/applyconfiguration/core/v1alpha1"
	inferenceclientgo "inftyai.com/llmaz/client-go/applyconfiguration/inference/v1alpha1"
	"inftyai.com/llmaz/pkg/backend"
	"inftyai.com/llmaz/pkg/util"
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
	log := log.FromContext(ctx)

	playground := &inferenceapi.Playground{}
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name, Namespace: req.Namespace}, playground); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var serviceApplyConfiguration *inferenceclientgo.ServiceApplyConfiguration

	if playground.Spec.ModelClaim != nil {
		modelName := playground.Spec.ModelClaim.ModelName
		model := &core.Model{}

		if err := r.Get(ctx, types.NamespacedName{Name: string(modelName)}, model); err != nil {
			log.Error(err, "model doesn't exist")
			return ctrl.Result{}, err
		}
		serviceApplyConfiguration = buildServiceApplyConfiguration(model, playground)
	}

	// TODO: handle MultiModelsClaims in the future.

	if err := setControllerReferenceForService(playground, serviceApplyConfiguration, r.Scheme); err != nil {
		return ctrl.Result{}, err
	}

	if err := util.Patch(ctx, r.Client, serviceApplyConfiguration); err != nil {
		log.Error(err, "failed to create inferenceService")
		return ctrl.Result{}, err
	}

	// Handle status.

	var service *inferenceapi.Service
	if err := r.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service); err != nil {
		log.Error(err, "failed to get inferenceService")
		return ctrl.Result{}, err
	}
	setPlaygroundCondition(playground, service)
	if err := r.Status().Update(ctx, playground); err != nil {
		log.Error(err, "failed to update Playground status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PlaygroundReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&inferenceapi.Playground{}).
		Complete(r)
}

func buildServiceApplyConfiguration(model *core.Model, playground *inferenceapi.Playground) *inferenceclientgo.ServiceApplyConfiguration {
	// Build metadata
	serviceApplyConfiguration := inferenceclientgo.Service(playground.Name, playground.Namespace)

	// Build spec.
	spec := inferenceclientgo.ServiceSpec()
	if playground.Spec.ElasticConfig != nil {
		config := inferenceclientgo.ElasticConfig()
		if max := playground.Spec.ElasticConfig.MaxReplicas; max != nil {
			config.WithMaxReplicas(*max)
		}
		if min := playground.Spec.ElasticConfig.MinReplicas; min != nil {
			config.WithMinReplicas(*min)
		}
		spec.WithElasticConfig(config)
	}

	if playground.Spec.ModelClaim != nil {
		claim := coreclientgo.MultiModelsClaim().
			WithModelNames(playground.Spec.ModelClaim.ModelName).
			WithInferenceFlavors(playground.Spec.ModelClaim.InferenceFlavors...)
		spec.WithMultiModelsClaims(claim)
	}

	spec.WithWorkloadTemplate(buildWorkloadTemplate(model, playground))
	serviceApplyConfiguration.WithSpec(spec)

	return serviceApplyConfiguration

	// TODO: handle MultiModelsClaims in the future.
}

// We do not want to maintain another workload like deployment for single-host cases so we choose lws here
// to cover both single-host and multi-host cases. There're some shortages for lws like can not force rolling
// update when one replica failed, we'll fix this in the kubernetes upstream.
// Model flavors will not be considered but in inferenceService controller to support accelerator fungibility.
func buildWorkloadTemplate(model *core.Model, playground *inferenceapi.Playground) lws.LeaderWorkerSetSpec {
	// FIXME: this should be leaderWorkerSetTemplateSpec, we should support in the lws upstream.
	workload := lws.LeaderWorkerSetSpec{}
	workload.Replicas = playground.Spec.Replicas

	backendName := inferenceapi.DefaultBackend
	if playground.Spec.BackendConfig != nil && playground.Spec.BackendConfig.Name != nil {
		backendName = *playground.Spec.BackendConfig.Name
	}
	bkd := backend.SwitchBackend(backendName)

	version := bkd.DefaultVersion()
	if playground.Spec.BackendConfig != nil && playground.Spec.BackendConfig.Version != nil {
		version = *playground.Spec.BackendConfig.Version
	}

	// TODO: should we also support secret here?
	args := []string{
		"--model", backend.DEFAULT_MODEL_PATH + model.Name,
		"--port", backend.DEFAULT_PORT,
	}
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
		// FIXME: we should merge the resources rather than simply replace.
		resources = corev1.ResourceRequirements{
			Limits:   playground.Spec.BackendConfig.Resources.Limits,
			Requests: playground.Spec.BackendConfig.Resources.Requests,
		}
	}

	// TODO: handle multi-host scenarios, e.g. nvidia.com/gpu: 32, means we'll split into 4 hosts.
	// Do we need another configuration for playground for multi-host use case? I guess no currently.
	workload.LeaderWorkerTemplate.WorkerTemplate = corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: map[string]string{
				core.ModelNameLabelKey:       string(model.Name),
				core.ModelFamilyNameLabelKey: string(model.Spec.FamilyName),
			},
		},
		Spec: corev1.PodSpec{
			// FIXME: add initContainer to handler data source
			// TODO: should we support image pull secret here?
			// TODO: support readiness/liveness
			Containers: []corev1.Container{
				{
					Name:      string(bkd.Name()),
					Image:     bkd.Image(version),
					Resources: resources,
					Command:   bkd.DefaultCommands(),
					Args:      args,
					Env:       envs,
				},
			},
		},
	}

	return workload
}

func setPlaygroundCondition(playground *inferenceapi.Playground, service *inferenceapi.Service) {
	// For the start up.
	if len(playground.Status.Conditions) == 0 {
		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundProgressing,
			Status:  metav1.ConditionTrue,
			Reason:  "Pending",
			Message: "Waiting for inferenceService ready",
		}
		apimeta.SetStatusCondition(&playground.Status.Conditions, condition)
		return
	}

	if apimeta.IsStatusConditionTrue(service.Status.Conditions, inferenceapi.ServiceAvailable) {
		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundAvailable,
			Status:  metav1.ConditionTrue,
			Reason:  "PlaygroundReady",
			Message: "Playground is ready",
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, condition)
	} else {
		// Still in starting up, no need to populate the condition.
		if apimeta.FindStatusCondition(playground.Status.Conditions, inferenceapi.PlaygroundAvailable) == nil {
			return
		}

		condition := metav1.Condition{
			Type:    inferenceapi.PlaygroundProgressing,
			Status:  metav1.ConditionTrue,
			Reason:  "PlaygroundInProgress",
			Message: "Playground is progressing",
		}
		apimeta.SetStatusCondition(&service.Status.Conditions, condition)

		// Set the available to false
		new_condition := metav1.Condition{
			Type:   inferenceapi.PlaygroundAvailable,
			Status: metav1.ConditionFalse,
		}
		apimeta.SetStatusCondition(&playground.Status.Conditions, new_condition)
	}
}

// setControllerReferenceForService set playground as the owner reference for inferenceService.
func setControllerReferenceForService(owner metav1.Object, saf *inferenceclientgo.ServiceApplyConfiguration, scheme *runtime.Scheme) error {
	// Validate the owner.
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
