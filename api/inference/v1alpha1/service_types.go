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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
)

// ServiceSpec defines the desired state of Service.
// Service controller will maintain multi-flavor of workloads with
// different accelerators for cost or performance considerations.
type ServiceSpec struct {
	// MultiModelsClaim represents claiming for multiple models with different claimModes,
	// like standard or speculative-decoding to support different inference scenarios.
	MultiModelsClaim coreapi.MultiModelsClaim `json:"multiModelsClaim,omitempty"`
	// WorkloadTemplate defines the underlying workload layout and configuration.
	// Note: the LWS spec might be twisted with various LWS instances to support
	// accelerator fungibility or other cutting-edge researches.
	// LWS supports both single-host and multi-host scenarios, for single host
	// cases, only need to care about replicas, rolloutStrategy and workerTemplate.
	WorkloadTemplate lws.LeaderWorkerSetSpec `json:"workloadTemplate"`
	// ElasticConfig defines the configuration for elastic usage,
	// e.g. the max/min replicas. Default to 0 ~ Inf+.
	// This requires to install the HPA first or will not work.
	// +optional
	ElasticConfig *ElasticConfig `json:"elasticConfig,omitempty"`
}

const (
	// ServiceAvailable means the inferenceService is available and all the
	// workloads are running as expected.
	ServiceAvailable = "Available"
	// ServiceProgressing means the inferenceService is progressing now, such as
	// in creation, rolling update or scaling up and down.
	ServiceProgressing = "Progressing"
)

// ServiceStatus defines the observed state of Service
type ServiceStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={isvc}

// Service is the Schema for the services API
type Service struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServiceSpec   `json:"spec,omitempty"`
	Status ServiceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServiceList contains a list of Service
type ServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Service `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Service{}, &ServiceList{})
}
