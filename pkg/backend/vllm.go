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

package backend

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
)

var _ Backend = (*VLLM)(nil)

type VLLM struct{}

const (
	image_registry = "vllm/vllm-openai"
)

func (v *VLLM) Name() inferenceapi.BackendName {
	return inferenceapi.VLLM
}

func (v *VLLM) Image(version string) string {
	return image_registry + ":" + version
}

func (v *VLLM) DefaultVersion() string {
	return "v0.5.1"
}

func (v *VLLM) DefaultResources() inferenceapi.ResourceRequirements {
	return inferenceapi.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
	}
}

func (v *VLLM) DefaultCommands() []string {
	return []string{"python3", "-m", "vllm.entrypoints.openai.api_server"}
}
