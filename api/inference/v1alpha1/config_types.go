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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

type BackendName string

const (
	DefaultBackend BackendName = "vllm"
)

type BackendRuntimeConfig struct {
	// BackendName represents the inference backend under the hood, e.g. vLLM.
	// +kubebuilder:default=vllm
	// +optional
	BackendName *BackendName `json:"backendName,omitempty"`
	// Version represents the backend version if you want a different one
	// from the default version.
	// +optional
	Version *string `json:"version,omitempty"`
	// Envs represents the environments set to the container.
	// +optional
	Envs []corev1.EnvVar `json:"envs,omitempty"`
	// ConfigName represents the recommended configuration name for the backend,
	// It will be inferred from the models in the runtime if not specified, e.g. default,
	// speculative-decoding.
	ConfigName *string `json:"configName,omitempty"`
	// Args defined here will "append" the args defined in the recommendedConfig,
	// either explicitly configured in configName or inferred in the runtime.
	// +optional
	Args []string `json:"args,omitempty"`
	// Resources represents the resource requirements for backend, like cpu/mem,
	// accelerators like GPU should not be defined here, but at the model flavors,
	// or the values here will be overwritten.
	// Resources defined here will "overwrite" the resources in the recommendedConfig.
	// +optional
	Resources *ResourceRequirements `json:"resources,omitempty"`
	// SharedMemorySize represents the size of /dev/shm required in the runtime of
	// inference workload.
	// SharedMemorySize defined here will "overwrite" the sharedMemorySize in the recommendedConfig.
	// +optional
	SharedMemorySize *resource.Quantity `json:"sharedMemorySize,omitempty"`
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
