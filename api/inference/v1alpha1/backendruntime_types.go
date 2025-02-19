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
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPATrigger represents the configuration of the HorizontalPodAutoscaler.
// Inspired by kubernetes.io/pkg/apis/autoscaling/types.go#HorizontalPodAutoscalerSpec.
// Note: HPA component should be installed in prior.
type HPATrigger struct {
	// metrics contains the specifications for which to use to calculate the
	// desired replica count (the maximum replica count across all metrics will
	// be used).  The desired replica count is calculated multiplying the
	// ratio between the target value and the current value by the current
	// number of pods.  Ergo, metrics used must decrease as the pod count is
	// increased, and vice-versa.  See the individual metric source types for
	// more information about how each type of metric must respond.
	// +optional
	Metrics []autoscalingv2.MetricSpec `json:"metrics,omitempty"`
	// behavior configures the scaling behavior of the target
	// in both Up and Down directions (scaleUp and scaleDown fields respectively).
	// If not set, the default HPAScalingRules for scale up and scale down are used.
	// +optional
	Behavior *autoscalingv2.HorizontalPodAutoscalerBehavior `json:"behavior,omitempty"`
}

// ScaleTrigger defines the rules to scale the workloads.
// Only one trigger cloud work at a time, mostly used in Playground.
type ScaleTrigger struct {
	// HPA represents the trigger configuration of the HorizontalPodAutoscaler.
	HPA *HPATrigger `json:"hpa,omitempty"`
}

// MultiHostCommands represents leader & worker commands for multiple nodes scenarios.
type MultiHostCommands struct {
	// Leader commands.
	// +optional
	Leader []string `json:"leader,omitempty"`
	// Worker commands.
	// +optional
	Worker []string `json:"worker,omitempty"`
}

// RecommendedConfig represents the recommended configurations for the backendRuntime,
// user can choose one of them to apply.
type RecommendedConfig struct {
	// Name represents the identifier of the config.
	Name string `json:"name"`
	// Args represents all the arguments for the command.
	// Argument around with {{ .CONFIG }} is a configuration waiting for render.
	// +optional
	Args []string `json:"args,omitempty"`
	// Resources represents the resource requirements for backend, like cpu/mem,
	// accelerators like GPU should not be defined here, but at the model flavors,
	// or the values here will be overwritten.
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
	// SharedMemorySize represents the size of /dev/shm required in the runtime of
	// inference workload.
	// +optional
	SharedMemorySize *resource.Quantity `json:"sharedMemorySize,omitempty"`
	// ScaleTrigger defines the rules to scale the workloads.
	// Only one trigger cloud work at a time.
	// +optional
	ScaleTrigger *ScaleTrigger `json:"scaleTrigger,omitempty"`
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
	// Envs represents the environments set to the container.
	// +optional
	Envs []corev1.EnvVar `json:"envs,omitempty"`
	// Periodic probe of backend liveness.
	// Backend will be restarted if the probe fails.
	// Cannot be updated.
	// +optional
	LivenessProbe *corev1.Probe `json:"livenessProbe,omitempty"`
	// Periodic probe of backend readiness.
	// Backend will be removed from service endpoints if the probe fails.
	// +optional
	ReadinessProbe *corev1.Probe `json:"readinessProbe,omitempty"`
	// StartupProbe indicates that the Backend has successfully initialized.
	// If specified, no other probes are executed until this completes successfully.
	// If this probe fails, the backend will be restarted, just as if the livenessProbe failed.
	// This can be used to provide different probe parameters at the beginning of a backend's lifecycle,
	// when it might take a long time to load data or warm a cache, than during steady-state operation.
	// +optional
	StartupProbe *corev1.Probe `json:"startupProbe,omitempty"`
	// RecommendedConfigs represents the recommended configurations for the backendRuntime.
	// +optional
	RecommendedConfigs []RecommendedConfig `json:"recommendedConfigs,omitempty"`
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
