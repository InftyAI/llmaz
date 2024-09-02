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

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
)

// PlaygroundSpec defines the desired state of Playground
type PlaygroundSpec struct {
	// Replicas represents the replica number of inference workloads.
	// +kubebuilder:default=1
	// +optional
	Replicas *int32 `json:"replicas,omitempty"`
	// ModelClaim represents claiming for one model, it's the standard claimMode
	// of multiModelsClaim compared to other modes like SpeculativeDecoding.
	// Most of the time, modelClaim is enough.
	// ModelClaim and multiModelsClaim are exclusive configured.
	// +optional
	ModelClaim *coreapi.ModelClaim `json:"modelClaim,omitempty"`
	// MultiModelsClaim represents claiming for multiple models with different claimModes,
	// like standard or speculative-decoding to support different inference scenarios.
	// ModelClaim and multiModelsClaim are exclusive configured.
	// +optional
	MultiModelsClaim *coreapi.MultiModelsClaim `json:"multiModelsClaim,omitempty"`
	// BackendConfig represents the inference backend configuration
	// under the hood, e.g. vLLM, which is the default backend.
	// +optional
	BackendConfig *BackendConfig `json:"backendConfig,omitempty"`
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
