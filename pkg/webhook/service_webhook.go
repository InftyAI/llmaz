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

package webhook

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/modelsource"
)

type ServiceWebhook struct{}

// SetupServiceWebhook will setup the manager to manage the webhooks
func SetupServiceWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&inferenceapi.Service{}).
		WithDefaulter(&ServiceWebhook{}).
		WithValidator(&ServiceWebhook{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-inference-llmaz-io-v1alpha1-service,mutating=true,failurePolicy=fail,sideEffects=None,groups=inference.llmaz.io,resources=services,verbs=create;update,versions=v1alpha1,name=mservice.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &ServiceWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *ServiceWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

//+kubebuilder:webhook:path=/validate-inference-llmaz-io-v1alpha1-service,mutating=false,failurePolicy=fail,sideEffects=None,groups=inference.llmaz.io,resources=services,verbs=create;update,versions=v1alpha1,name=vservice.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &ServiceWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ServiceWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(obj)
	service := obj.(*inferenceapi.Service)
	for _, err := range validation.IsDNS1123Label(service.Name) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("metadata.name"), service.Name, err))
	}

	runnerContainerExists := false
	for _, container := range service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers {
		if container.Name == modelSource.MODEL_RUNNER_CONTAINER_NAME {
			runnerContainerExists = true
			break
		}
	}
	if !runnerContainerExists {
		allErrs = append(allErrs, field.Forbidden(field.NewPath("spec.workloadTemplate.leaderWorkerTemplate.workerTemplate.spec.containers"), "model-runner container doesn't exist"))
	}

	return nil, allErrs.ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ServiceWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(newObj)
	return nil, allErrs.ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *ServiceWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *ServiceWebhook) generateValidate(obj runtime.Object) field.ErrorList {
	service := obj.(*inferenceapi.Service)
	specPath := field.NewPath("spec")
	var allErrs field.ErrorList

	mainModelCount := 0
	var speculativeDecoding bool
	for _, model := range service.Spec.ModelClaims.Models {
		if model.Role == nil || *model.Role == coreapi.MainRole {
			mainModelCount += 1
		}
		if model.Role != nil && *model.Role == coreapi.DraftRole {
			speculativeDecoding = true
		}
	}

	if speculativeDecoding {
		if len(service.Spec.ModelClaims.Models) != 2 {
			allErrs = append(allErrs, field.Forbidden(specPath.Child("modelClaims", "models"), "only two models are allowed in speculativeDecoding mode"))
		}
		if mainModelCount != 1 {
			allErrs = append(allErrs, field.Forbidden(specPath.Child("modelClaims", "models"), "main model is required"))
		}
	}
	return allErrs
}
