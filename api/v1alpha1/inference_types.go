/*
Copyright 2023.

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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InferenceSpec defines the desired state of Inference
type InferenceSpec struct {
	// ModelNameOrPath represents the model name or the local path.
	ModelNameOrPath string `json:"modelNameOrPath,omitempty"`
	// Backend indicates the inference backend under the hood, e.g. vLLM.
	// Default to use huggingface library.
	//
	// +optional
	// +kubebuilder:validation:Enum={vllm,huggingface}
	// +kubebuilder:default=huggingface
	Backend *string `json:"backend,omitempty"`
	// WorkloadTemplate defines the underlying workload layout and configuration,
	// e.g. the leader/worker templates and replicas.
	WorkloadTemplate lws.LeaderWorkerSetSpec `json:"workloadTemplate"`
	// ElasticConfig defines the configuration for elastic usage,
	// e.g. the max/min replicas.
	// Default to 0 ~ Inf+.
	// +optional
	ElasticConfig *ElasticConfig `json:"elasticConfig,omitempty"`
}

type ElasticConfig struct {
	// MinReplicas indicates the minimum number of Inference instances based on the traffic.
	// Default to nil means we can scale down the instances to 0.
	// +optional
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// MaxReplicas indicates the maximum number of Inference instances based on the traffic.
	// Default to nil means there's no limit for the instance number.
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
}

// InferenceStatus defines the observed state of Inference
type InferenceStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Inference is the Schema for the inferences API
type Inference struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InferenceSpec   `json:"spec,omitempty"`
	Status InferenceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// InferenceList contains a list of Inference
type InferenceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Inference `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Inference{}, &InferenceList{})
}
