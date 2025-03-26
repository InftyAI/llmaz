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
	// ModelClaim represents claiming for one model, it's a simplified use case
	// of modelClaims. Most of the time, modelClaim is enough.
	// ModelClaim and modelClaims are exclusive configured.
	// +optional
	ModelClaim *coreapi.ModelClaim `json:"modelClaim,omitempty"`
	// ModelClaims represents claiming for multiple models for more complicated
	// use cases like speculative-decoding.
	// ModelClaims and modelClaim are exclusive configured.
	// +optional
	ModelClaims *coreapi.ModelClaims `json:"modelClaims,omitempty"`
	// BackendRuntimeConfig represents the inference backendRuntime configuration
	// under the hood, e.g. vLLM, which is the default backendRuntime.
	// +optional
	BackendRuntimeConfig *BackendRuntimeConfig `json:"backendRuntimeConfig,omitempty"`
	// ElasticConfig defines the configuration for elastic usage,
	// e.g. the max/min replicas.
	// +optional
	ElasticConfig *ElasticConfig `json:"elasticConfig,omitempty"`
}

type ElasticConfig struct {
	// MinReplicas indicates the minimum number of inference workloads based on the traffic.
	// Default to 1.
	// MinReplicas couldn't be 0 now, will support serverless in the future.
	// +kubebuilder:default=1
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// MaxReplicas indicates the maximum number of inference workloads based on the traffic.
	// Default to nil means there's no limit for the instance number.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Minimum:=1
	MaxReplicas int32 `json:"maxReplicas,omitempty"`
	// ScaleTrigger defines the rules to scale the workloads.
	// Only one trigger cloud work at a time, mostly used in Playground.
	// ScaleTrigger defined here will "overwrite" the scaleTrigger in the recommendedConfig.
	// +optional
	ScaleTrigger *ScaleTrigger `json:"scaleTrigger,omitempty"`
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
	// Replicas track the replicas that have been created, whether ready or not.
	Replicas int32 `json:"replicas"`
	// Selector points to the string form of a label selector which will be used by HPA.
	Selector string `json:"selector,omitempty"`
}

//+genclient
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName={pl}
//+kubebuilder:subresource:scale:specpath=.spec.replicas,statuspath=.status.replicas,selectorpath=.status.selector

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
