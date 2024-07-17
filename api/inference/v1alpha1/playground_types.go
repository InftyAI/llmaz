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

	api "inftyai.com/llmaz/api/core/v1alpha1"
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
	// +optional
	ModelClaim *api.ModelClaim `json:"modelClaim,omitempty"`
	// MultiModelsClaims represents multiple modelClaim, which is useful when different
	// sub-workload has different accelerator requirements, like the state-of-the-art
	// technology called splitwise, the workload template is shared by both.
	// ModelClaim and multiModelsClaims are exclusive configured.
	// +optional
	MultiModelsClaims []api.MultiModelsClaim `json:"multiModelsClaims,omitempty"`
	// BackendConfig represents the inference backend configuration
	// under the hood, e.g. vLLM, which is the default backend.
	// +optional
	BackendConfig *BackendConfig `json:"backendConfig,omitempty"`
	// ElasticConfig defines the configuration for elastic usage,
	// e.g. the max/min replicas. Default to 0 ~ Inf+.
	// +optional
	ElasticConfig *ElasticConfig `json:"elasticConfig,omitempty"`
}

const (
	// PlaygroundProgressing means the Playground is progressing now, such as waiting for the
	// inference service creation, rolling update or scaling up and down.
	PlaygroundProgressing = "Progressing"
	// PlaygroundAvailable indicates the corresponding inference service is available now.
	PlaygroundAvailable string = "Available"
)

// PlaygroundStatus defines the observed state of Playground
type PlaygroundStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+genclient
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
