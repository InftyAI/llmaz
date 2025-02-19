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
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	corev1alpha1 "github.com/inftyai/llmaz/api/core/v1alpha1"
)

// ModelSpecApplyConfiguration represents a declarative configuration of the ModelSpec type for use
// with apply.
type ModelSpecApplyConfiguration struct {
	FamilyName      *corev1alpha1.ModelName            `json:"familyName,omitempty"`
	Source          *ModelSourceApplyConfiguration     `json:"source,omitempty"`
	InferenceConfig *InferenceConfigApplyConfiguration `json:"inferenceConfig,omitempty"`
}

// ModelSpecApplyConfiguration constructs a declarative configuration of the ModelSpec type for use with
// apply.
func ModelSpec() *ModelSpecApplyConfiguration {
	return &ModelSpecApplyConfiguration{}
}

// WithFamilyName sets the FamilyName field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FamilyName field is set to the value of the last call.
func (b *ModelSpecApplyConfiguration) WithFamilyName(value corev1alpha1.ModelName) *ModelSpecApplyConfiguration {
	b.FamilyName = &value
	return b
}

// WithSource sets the Source field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Source field is set to the value of the last call.
func (b *ModelSpecApplyConfiguration) WithSource(value *ModelSourceApplyConfiguration) *ModelSpecApplyConfiguration {
	b.Source = value
	return b
}

// WithInferenceConfig sets the InferenceConfig field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the InferenceConfig field is set to the value of the last call.
func (b *ModelSpecApplyConfiguration) WithInferenceConfig(value *InferenceConfigApplyConfiguration) *ModelSpecApplyConfiguration {
	b.InferenceConfig = value
	return b
}
