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

package webhook

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	helper "github.com/inftyai/llmaz/pkg/controller_helper"
)

type PlaygroundWebhook struct{}

// SetupPlaygroundWebhook will setup the manager to manage the webhooks
func SetupPlaygroundWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&inferenceapi.Playground{}).
		WithDefaulter(&PlaygroundWebhook{}).
		WithValidator(&PlaygroundWebhook{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-inference-llmaz-io-v1alpha1-playground,mutating=true,failurePolicy=fail,sideEffects=None,groups=inference.llmaz.io,resources=playgrounds,verbs=create;update,versions=v1alpha1,name=mplayground.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &PlaygroundWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *PlaygroundWebhook) Default(ctx context.Context, obj runtime.Object) error {
	playground := obj.(*inferenceapi.Playground)

	var modelName string
	if playground.Spec.ModelClaim != nil {
		modelName = string(playground.Spec.ModelClaim.ModelName)
	} else if playground.Spec.ModelClaims != nil {
		for _, model := range playground.Spec.ModelClaims.Models {
			if model.Role == nil || *model.Role == coreapi.MainRole {
				modelName = string(model.Name)
			}
		}
	}

	if playground.Labels == nil {
		playground.Labels = map[string]string{}
	}
	playground.Labels[coreapi.ModelNameLabelKey] = modelName

	return nil
}

//+kubebuilder:webhook:path=/validate-inference-llmaz-io-v1alpha1-playground,mutating=false,failurePolicy=fail,sideEffects=None,groups=inference.llmaz.io,resources=playgrounds,verbs=create;update,versions=v1alpha1,name=vplayground.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &PlaygroundWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *PlaygroundWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(obj)
	playground := obj.(*inferenceapi.Playground)
	for _, err := range validation.IsDNS1123Label(playground.Name) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("metadata.name"), playground.Name, err))
	}
	return nil, allErrs.ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *PlaygroundWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(newObj)
	return nil, allErrs.ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *PlaygroundWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *PlaygroundWebhook) generateValidate(obj runtime.Object) field.ErrorList {
	playground := obj.(*inferenceapi.Playground)
	specPath := field.NewPath("spec")

	var allErrs field.ErrorList
	if playground.Spec.ModelClaim == nil && playground.Spec.ModelClaims == nil {
		allErrs = append(allErrs, field.Forbidden(specPath, "modelClaim and modelClaims couldn't be both nil"))
	}
	if playground.Spec.ModelClaims != nil {
		mainModelCount := 0

		for _, model := range playground.Spec.ModelClaims.Models {
			if model.Name == coreapi.ModelName(coreapi.MainRole) {
				mainModelCount += 1
			}
		}

		// We only have to detect whether this is speculativeDecoding mode, so set the second argument to false is ok.
		arg := helper.DetectArgFrom(playground)
		if arg == helper.SpeculativeDecodingArg {
			if len(playground.Spec.ModelClaims.Models) != 2 {
				allErrs = append(allErrs, field.Forbidden(specPath.Child("modelClaims", "models"), "only two models are allowed in speculativeDecoding mode"))
			}
		}

		if mainModelCount > 1 {
			allErrs = append(allErrs, field.Forbidden(specPath.Child("modelClaims", "models"), "only one main model is allowed"))
		}
	}

	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Resources != nil {
		requirements := playground.Spec.BackendRuntimeConfig.Resources
		for resourceName, limit := range requirements.Limits {
			request, ok := requirements.Requests[resourceName]
			if !ok {
				continue
			}

			if limit.Cmp(request) == -1 {
				allErrs = append(allErrs, field.Forbidden(specPath.Child("backendRuntimeConfig", "resources"),
					fmt.Sprintf("limit (%v) for %s must be greater than or equal to request (%v)", limit, resourceName, request)))
			}
		}
	}

	if playground.Spec.ElasticConfig != nil {
		if playground.Spec.ElasticConfig.MinReplicas != nil && *playground.Spec.ElasticConfig.MinReplicas == 0 {
			allErrs = append(allErrs, field.Forbidden(specPath.Child("elasticConfig.minReplicas"), "minReplicas couldn't be 0"))
		}

		if playground.Spec.ElasticConfig.MinReplicas != nil && playground.Spec.ElasticConfig.MaxReplicas != nil {
			if *playground.Spec.ElasticConfig.MinReplicas >= *playground.Spec.ElasticConfig.MaxReplicas {
				allErrs = append(allErrs, field.Invalid(specPath.Child("elasticConfig.scaleTrigger.hpa"), *playground.Spec.ElasticConfig.MinReplicas, "minReplicas must be less than maxReplicas"))
			}
		}
	}

	if playground.Spec.ElasticConfig != nil && playground.Spec.ElasticConfig.ScaleTrigger != nil {
		if playground.Spec.ElasticConfig.ScaleTrigger.HPA == nil {
			allErrs = append(allErrs, field.Forbidden(specPath.Child("backendRuntime.scaleTrigger.hpa"), "hpa couldn't be nil"))
		}
	}

	return allErrs
}
