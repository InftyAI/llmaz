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
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/pkg/util"
)

type BackendRuntimeWebhook struct{}

// SetupBackendRuntimeWebhook will setup the manager to manage the webhooks
func SetupBackendRuntimeWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&inferenceapi.BackendRuntime{}).
		WithDefaulter(&BackendRuntimeWebhook{}).
		WithValidator(&BackendRuntimeWebhook{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-inference-llmaz-io-v1alpha1-backendruntime,mutating=true,failurePolicy=fail,sideEffects=None,groups=inference.llmaz.io,resources=backendruntimes,verbs=create;update,versions=v1alpha1,name=mbackendruntime.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &BackendRuntimeWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *BackendRuntimeWebhook) Default(ctx context.Context, obj runtime.Object) error {
	return nil
}

//+kubebuilder:webhook:path=/validate-inference-llmaz-io-v1alpha1-backendruntime,mutating=false,failurePolicy=fail,sideEffects=None,groups=inference.llmaz.io,resources=backendruntimes,verbs=create;update,versions=v1alpha1,name=vbackendruntime.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &BackendRuntimeWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *BackendRuntimeWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(obj)
	return nil, allErrs.ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *BackendRuntimeWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(newObj)
	return nil, allErrs.ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *BackendRuntimeWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *BackendRuntimeWebhook) generateValidate(obj runtime.Object) field.ErrorList {
	backend := obj.(*inferenceapi.BackendRuntime)
	specPath := field.NewPath("spec")

	var allErrs field.ErrorList

	// Validate resources.
	for _, recommend := range backend.Spec.RecommendedConfigs {
		if recommend.Resources == nil {
			continue
		}
		for k, v := range recommend.Resources.Limits {
			if requestV, ok := recommend.Resources.Requests[k]; ok {
				if v.Cmp(requestV) == -1 {
					allErrs = append(allErrs, field.Forbidden(specPath.Child("resources"), fmt.Sprintf("resource limit of %s is less than resource request", k)))
				}
			}
		}
	}

	names := []string{}
	for _, recommend := range backend.Spec.RecommendedConfigs {
		if util.In(names, recommend.Name) {
			allErrs = append(allErrs, field.Forbidden(specPath.Child("args", "name"), fmt.Sprintf("duplicated name %s", recommend.Name)))
		}
		names = append(names, recommend.Name)
	}
	return allErrs
}
