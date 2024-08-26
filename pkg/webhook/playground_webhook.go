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

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
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
	if playground.Spec.ModelClaim == nil && len(playground.Spec.MultiModelsClaims) == 0 {
		allErrs = append(allErrs, field.Forbidden(specPath, "modelClaim and multiModelsClaims couldn't be both empty"))
	}
	return allErrs
}
