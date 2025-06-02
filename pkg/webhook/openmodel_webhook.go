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

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/modelsource"
	"github.com/inftyai/llmaz/pkg/util"
)

type OpenModelWebhook struct{}

// SetupOpenModelWebhook will setup the manager to manage the webhooks
func SetupOpenModelWebhook(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(&coreapi.OpenModel{}).
		WithDefaulter(&OpenModelWebhook{}).
		WithValidator(&OpenModelWebhook{}).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-llmaz-io-v1alpha1-openmodel,mutating=true,failurePolicy=fail,sideEffects=None,groups=llmaz.io,resources=openmodels,verbs=create;update,versions=v1alpha1,name=mopenmodel.kb.io,admissionReviewVersions=v1

var _ webhook.CustomDefaulter = &OpenModelWebhook{}

var SUPPORTED_OBJ_STORES = map[string]struct{}{
	modelSource.GCS:      {},
	modelSource.OSS:      {},
	modelSource.S3:       {},
	modelSource.Ollama:   {},
	modelSource.HostPath: {},
}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (w *OpenModelWebhook) Default(ctx context.Context, obj runtime.Object) error {
	model := obj.(*coreapi.OpenModel)
	if model.Labels == nil {
		model.Labels = map[string]string{}
	}
	model.Labels[coreapi.ModelFamilyNameLabelKey] = string(model.Spec.FamilyName)
	return nil
}

//+kubebuilder:webhook:path=/validate-llmaz-io-v1alpha1-openmodel,mutating=false,failurePolicy=fail,sideEffects=None,groups=llmaz.io,resources=openmodels,verbs=create;update,versions=v1alpha1,name=vopenmodel.kb.io,admissionReviewVersions=v1

var _ webhook.CustomValidator = &OpenModelWebhook{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (w *OpenModelWebhook) ValidateCreate(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(obj)
	model := obj.(*coreapi.OpenModel)
	for _, err := range validation.IsDNS1123Label(model.Name) {
		allErrs = append(allErrs, field.Invalid(field.NewPath("metadata.name"), model.Name, err))
	}
	return nil, allErrs.ToAggregate()
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (w *OpenModelWebhook) ValidateUpdate(ctx context.Context, oldObj, newObj runtime.Object) (admission.Warnings, error) {
	allErrs := w.generateValidate(newObj)
	return nil, allErrs.ToAggregate()
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (w *OpenModelWebhook) ValidateDelete(ctx context.Context, obj runtime.Object) (admission.Warnings, error) {
	return nil, nil
}

// TODO: once namespaced models enabled, we should validate that only such volume path
// /mnt/models/<model-namespace>/ is allowed.
func (w *OpenModelWebhook) generateValidate(obj runtime.Object) field.ErrorList {
	model := obj.(*coreapi.OpenModel)
	specPath := field.NewPath("spec")
	sourcePath := specPath.Child("source")

	var allErrs field.ErrorList
	if model.Spec.Source.ModelHub == nil && model.Spec.Source.URI == nil {
		allErrs = append(allErrs, field.Forbidden(sourcePath, "Source can't be all null"))
	}

	if model.Spec.Source.URI != nil {
		if protocol, address, err := util.ParseURI(string(*model.Spec.Source.URI)); err != nil {
			allErrs = append(allErrs, field.Invalid(sourcePath.Child("uri"), *model.Spec.Source.URI, "URI with wrong format"))
		} else {
			if _, ok := SUPPORTED_OBJ_STORES[protocol]; !ok {
				allErrs = append(allErrs, field.Invalid(sourcePath.Child("uri"), *model.Spec.Source.URI, "URI with unsupported protocol"))
			} else {
				switch protocol {
				case modelSource.OSS:
					if _, _, _, err := util.ParseOSS(address); err != nil {
						allErrs = append(allErrs, field.Invalid(sourcePath.Child("uri"), *model.Spec.Source.URI, "URI with wrong address"))
					}
				case modelSource.S3, modelSource.GCS:
					if _, _, err := util.ParseS3(address); err != nil {
						allErrs = append(allErrs, field.Invalid(sourcePath.Child("uri"), *model.Spec.Source.URI, "URI with wrong address"))
					}
				}
			}
		}
	}

	if model.Spec.Source.ModelHub != nil {
		if model.Spec.Source.ModelHub.Filename != nil && model.Spec.Source.ModelHub.Name != nil && *model.Spec.Source.ModelHub.Name == coreapi.MODEL_SCOPE {
			allErrs = append(allErrs, field.Invalid(sourcePath.Child("modelHub.filename"), *model.Spec.Source.ModelHub.Filename, "Filename can only set once modeHub is Huggingface"))
		}
	}

	if model.Spec.Source.ModelHub != nil && model.Spec.Source.ModelHub.Filename != nil {
		if model.Spec.Source.ModelHub.AllowPatterns != nil {
			allErrs = append(allErrs, field.Invalid(sourcePath.Child("modelHub.allowPatterns"), model.Spec.Source.ModelHub.AllowPatterns, "Once Filename is set, allowPatterns should be nil"))
		}
		if model.Spec.Source.ModelHub.IgnorePatterns != nil {
			allErrs = append(allErrs, field.Invalid(sourcePath.Child("modelHub.ignorePatterns"), model.Spec.Source.ModelHub.IgnorePatterns, "Once Filename is set, ignorePatterns should be nil"))
		}
	}

	return allErrs
}
