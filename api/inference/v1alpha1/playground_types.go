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
	api "inftyai.com/llmaz/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PlaygroundSpec defines the desired state of Playground
type PlaygroundSpec struct {
	// Replicas represents the replica number of inference workloads.
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// ModelClaim represents one modelClaim, it's a simple configuration
	// compared to multiModelsClaims only work for one model and one claim.
	// ModelClaim and multiModelsClaims are exclusive configured.
	// Note: properties (nodeSelectors, resources, e.g.) of the model flavors
	// will be applied to the workload if not exist.
	ModelClaim api.ModelClaim `json:"modelClaim"`
	// MultiModelsClaims represents multiple modelClaim, which is useful when different
	// sub-workload has different accelerator requirements, like the state-of-the-art
	// technology called splitwise, the workload template is shared by both.
	// ModelClaim and multiModelsClaims are exclusive configured.
	// +kubebuilder:validation:MinItems=1
	MultiModelsClaims []api.MultiModelsClaim `json:"multiModelsClaims"`
	// BackendConfig represents the inference backend configuration
	// under the hood, e.g. vLLM, which is the default backend.
	// +optional
	BackendConfig *BackendConfig `json:"backendConfig,omitempty"`
	// ElasticConfig defines the configuration for elastic usage,
	// e.g. the max/min replicas. Default to 0 ~ Inf+.
	// +optional
	ElasticConfig *ElasticConfig `json:"elasticConfig,omitempty"`
}

// PlaygroundStatus defines the observed state of Playground
type PlaygroundStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Playground is the Schema for the playgrounds API
type Playground struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PlaygroundSpec   `json:"spec,omitempty"`
	Status PlaygroundStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PlaygroundList contains a list of Playground
type PlaygroundList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Playground `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Playground{}, &PlaygroundList{})
}
