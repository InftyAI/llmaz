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
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	core "inftyai.com/llmaz/api/core/v1alpha1"
)

type ModelWebhook struct{}

// SetupModelWebhook will setup the manager to manage the webhooks
func SetupModelWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&core.Model{}).
		WithDefaulter(&ModelWebhook{}).
		WithValidator(&ModelWebhook{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-llmaz-io-v1alpha1-model,mutating=true,failurePolicy=fail,sideEffects=None,groups=llmaz.io,resources=models,verbs=create;update,versions=v1alpha1,name=mmodel.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &ModelWebhook{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *ModelWebhook) Default(ctx context.Context, obj runtime.Object) error {
	model := obj.(*core.Model)
	if model.Labels == nil {
		model.Labels = map[string]string{}
	}
	model.Labels[core.ModelFamilyNameLabelKey] = string(model.Spec.FamilyName)
	return nil
}

//+kubebuilder:webhook:path=/validate-llmaz-io-v1alpha1-model,mutating=false,failurePolicy=fail,sideEffects=None,groups=llmaz.io,resources=models,verbs=create;update,versions=v1alpha1,name=vmodel.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &ModelWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *ModelWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	warnings, allErrs := w.generateValidate(obj)
	return warnings, allErrs.ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *ModelWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	warnings, allErrs := w.generateValidate(newObj)
	return warnings, allErrs.ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *ModelWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

func (w *ModelWebhook) generateValidate(obj runtime.Object) (admission.Warnings, field.ErrorList) {
	model := obj.(*core.Model)
	dataSourcePath := field.NewPath("spec", "dataSource")

	var allErrs field.ErrorList
	if model.Spec.DataSource.ModelHub == nil {
		allErrs = append(allErrs, field.Forbidden(dataSourcePath, "data source can't be all null"))
	}
	return nil, allErrs
}
