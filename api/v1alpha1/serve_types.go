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
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ServeSpec defines the desired state of Serve
type ServeSpec struct {
	// ModelNameOrPath represents the model name or the local path.
	ModelNameOrPath string `json:"modelNameOrPath,omitempty"`
	// Replicas indicates the number of Serve instances.
	Replicas int32 `json:"replicas,omitempty"`
	// MinReplicas indicates the minimum number of Serve instances based on the traffic.
	// Default to nil means we can scale down the instances to 0.
	MinReplicas *int32 `json:"minReplicas,omitempty"`
	// MaxReplicas indicates the maximum number of Serve instances based on the traffic.
	// Default to nil means there's no limit for the instance number.
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// Backend indicates the inference backend under the hood, e.g. vLLM, text-generation-inference.
	// Default to use huggingface library.
	Backend *string `json:"backend,omitempty"`
}

// ServeStatus defines the observed state of Serve
type ServeStatus struct {
	// Conditions represents the Serve condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Serve is the Schema for the serves API
type Serve struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ServeSpec   `json:"spec,omitempty"`
	Status ServeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ServeList contains a list of Serve
type ServeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Serve `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Serve{}, &ServeList{})
}
