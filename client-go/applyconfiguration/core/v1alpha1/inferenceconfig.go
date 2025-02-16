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
	resource "k8s.io/apimachinery/pkg/api/resource"
)

// InferenceConfigApplyConfiguration represents a declarative configuration of the InferenceConfig type for use
// with apply.
type InferenceConfigApplyConfiguration struct {
	Flavors          []FlavorApplyConfiguration `json:"flavors,omitempty"`
	SharedMemorySize *resource.Quantity         `json:"sharedMemorySize,omitempty"`
}

// InferenceConfigApplyConfiguration constructs a declarative configuration of the InferenceConfig type for use with
// apply.
func InferenceConfig() *InferenceConfigApplyConfiguration {
	return &InferenceConfigApplyConfiguration{}
}

// WithFlavors adds the given value to the Flavors field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Flavors field.
func (b *InferenceConfigApplyConfiguration) WithFlavors(values ...*FlavorApplyConfiguration) *InferenceConfigApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithFlavors")
		}
		b.Flavors = append(b.Flavors, *values[i])
	}
	return b
}

// WithSharedMemorySize sets the SharedMemorySize field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the SharedMemorySize field is set to the value of the last call.
func (b *InferenceConfigApplyConfiguration) WithSharedMemorySize(value resource.Quantity) *InferenceConfigApplyConfiguration {
	b.SharedMemorySize = &value
	return b
}
