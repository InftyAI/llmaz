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
// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/inftyai/llmaz/api/core/v1alpha1"
)

// ModelRepresentativeApplyConfiguration represents an declarative configuration of the ModelRepresentative type for use
// with apply.
type ModelRepresentativeApplyConfiguration struct {
	Name *v1alpha1.ModelName `json:"name,omitempty"`
	Role *v1alpha1.ModelRole `json:"role,omitempty"`
}

// ModelRepresentativeApplyConfiguration constructs an declarative configuration of the ModelRepresentative type for use with
// apply.
func ModelRepresentative() *ModelRepresentativeApplyConfiguration {
	return &ModelRepresentativeApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *ModelRepresentativeApplyConfiguration) WithName(value v1alpha1.ModelName) *ModelRepresentativeApplyConfiguration {
	b.Name = &value
	return b
}

// WithRole sets the Role field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Role field is set to the value of the last call.
func (b *ModelRepresentativeApplyConfiguration) WithRole(value v1alpha1.ModelRole) *ModelRepresentativeApplyConfiguration {
	b.Role = &value
	return b
}