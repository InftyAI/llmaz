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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackendRuntimeArg is the preset arguments for easy to use.
// Three preset names are provided: default, speculative-decoding, model-parallelism,
// do not change the name.
type BackendRuntimeArg struct {
	// Name represents the identifier of the backendRuntime argument.
	Name string `json:"name"`
	// Flags represents all the preset configurations.
	// Flag around with {{ .CONFIG }} is a configuration waiting for render.
	Flags []string `json:"flags,omitempty"`
}

// MultiHostCommands represents leader & worker commands for multiple nodes scenarios.
type MultiHostCommands struct {
	Leader []string `json:"leader,omitempty"`
	Worker []string `json:"worker,omitempty"`
}

// BackendRuntimeSpec defines the desired state of BackendRuntime
type BackendRuntimeSpec struct {
	// Commands represents the default commands for the backendRuntime.
	// +optional
	Commands []string `json:"commands,omitempty"`
	// MultiHostCommands represents leader and worker commands for nodes with
	// different roles.
	// +optional
	MultiHostCommands *MultiHostCommands `json:"multiHostCommands,omitempty"`
	// Image represents the default image registry of the backendRuntime.
	// It will work together with version to make up a real image.
	Image string `json:"image"`
	// Version represents the default version of the backendRuntime.
	// It will be appended to the image as a tag.
	Version string `json:"version"`
	// Args represents the preset arguments of the backendRuntime.
	// They can be appended or overwritten by the Playground backendRuntimeConfig.
	Args []BackendRuntimeArg `json:"args,omitempty"`
	// Envs represents the environments set to the container.
	// +optional
	Envs []corev1.EnvVar `json:"envs,omitempty"`
	// Resources represents the resource requirements for backendRuntime, like cpu/mem,
	// accelerators like GPU should not be defined here, but at the model flavors,
	// or the values here will be overwritten.
	Resources ResourceRequirements `json:"resources"`
}

// BackendRuntimeStatus defines the observed state of BackendRuntime
type BackendRuntimeStatus struct {
	// Conditions represents the Inference condition.
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:shortName=br,scope=Cluster

// BackendRuntime is the Schema for the backendRuntime API
type BackendRuntime struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BackendRuntimeSpec   `json:"spec,omitempty"`
	Status BackendRuntimeStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// BackendRuntimeList contains a list of BackendRuntime
type BackendRuntimeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BackendRuntime `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BackendRuntime{}, &BackendRuntimeList{})
}
