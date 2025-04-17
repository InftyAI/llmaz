/*
Copyright 2025 The InftyAI Team.

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
	corev1alpha1 "github.com/inftyai/llmaz/client-go/applyconfiguration/core/v1alpha1"
	v1 "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

// ServiceSpecApplyConfiguration represents a declarative configuration of the ServiceSpec type for use
// with apply.
type ServiceSpecApplyConfiguration struct {
	ModelClaims      *corev1alpha1.ModelClaimsApplyConfiguration `json:"modelClaims,omitempty"`
	Replicas         *int32                                      `json:"replicas,omitempty"`
	WorkloadTemplate *v1.LeaderWorkerTemplate                    `json:"workloadTemplate,omitempty"`
	RolloutStrategy  *v1.RolloutStrategy                         `json:"rolloutStrategy,omitempty"`
}

// ServiceSpecApplyConfiguration constructs a declarative configuration of the ServiceSpec type for use with
// apply.
func ServiceSpec() *ServiceSpecApplyConfiguration {
	return &ServiceSpecApplyConfiguration{}
}

// WithModelClaims sets the ModelClaims field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the ModelClaims field is set to the value of the last call.
func (b *ServiceSpecApplyConfiguration) WithModelClaims(value *corev1alpha1.ModelClaimsApplyConfiguration) *ServiceSpecApplyConfiguration {
	b.ModelClaims = value
	return b
}

// WithReplicas sets the Replicas field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Replicas field is set to the value of the last call.
func (b *ServiceSpecApplyConfiguration) WithReplicas(value int32) *ServiceSpecApplyConfiguration {
	b.Replicas = &value
	return b
}

// WithWorkloadTemplate sets the WorkloadTemplate field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the WorkloadTemplate field is set to the value of the last call.
func (b *ServiceSpecApplyConfiguration) WithWorkloadTemplate(value v1.LeaderWorkerTemplate) *ServiceSpecApplyConfiguration {
	b.WorkloadTemplate = &value
	return b
}

// WithRolloutStrategy sets the RolloutStrategy field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the RolloutStrategy field is set to the value of the last call.
func (b *ServiceSpecApplyConfiguration) WithRolloutStrategy(value v1.RolloutStrategy) *ServiceSpecApplyConfiguration {
	b.RolloutStrategy = &value
	return b
}
