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
)

type BackendName string

const (
	DefaultBackend BackendName = "vllm"
)

type BackendRuntimeConfig struct {
	// Name represents the inference backend under the hood, e.g. vLLM.
	// +kubebuilder:default=vllm
	// +optional
	Name *BackendName `json:"name,omitempty"`
	// Version represents the backend version if you want a different one
	// from the default version.
	// +optional
	Version *string `json:"version,omitempty"`
	// Args represents the specified arguments of the backendRuntime,
	// will be append to the backendRuntime.spec.Args.
	Args *BackendRuntimeArg `json:"args,omitempty"`
	// Envs represents the environments set to the container.
	// +optional
	Envs []corev1.EnvVar `json:"envs,omitempty"`
	// Resources represents the resource requirements for backend, like cpu/mem,
	// accelerators like GPU should not be defined here, but at the model flavors,
	// or the values here will be overwritten.
	Resources *ResourceRequirements `json:"resources,omitempty"`
}

// TODO: Do not support DRA yet, we can support that once needed.
type ResourceRequirements struct {
	// Limits describes the maximum amount of compute resources allowed.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Limits corev1.ResourceList `json:"limits,omitempty"`
	// Requests describes the minimum amount of compute resources required.
	// If Requests is omitted for a container, it defaults to Limits if that is explicitly specified,
	// otherwise to an implementation-defined value. Requests cannot exceed Limits.
	// More info: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// +optional
	Requests corev1.ResourceList `json:"requests,omitempty"`
}

// ScaleTriggerRef refers to the configured scaleTrigger in the backendRuntime.
type ScaleTriggerRef struct {
	// Name represents the scale trigger name defined in the backendRuntime.scaleTriggers.
	Name string `json:"name"`
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
	// +optional
	MaxReplicas *int32 `json:"maxReplicas,omitempty"`
	// ScaleTriggerRef refers to the configured scaleTrigger in the backendRuntime
	// with tuned target value.
	// ScaleTriggerRef and ScaleTrigger can't be set at the same time.
	// +optional
	ScaleTriggerRef *ScaleTriggerRef `json:"scaleTriggerRef,omitempty"`
	// ScaleTrigger defines a set of triggers to scale the workloads.
	// If not defined, trigger configured in backendRuntime will be used,
	// otherwise, trigger defined here will overwrite the defaulted ones.
	// ScaleTriggerRef and ScaleTrigger can't be set at the same time.
	// +optional
	ScaleTrigger *ScaleTrigger `json:"scaleTrigger,omitempty"`
}
