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
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/model_source"
)

var _ Backend = (*VLLM)(nil)
var _ SpeculativeBackend = (*VLLM)(nil)

type VLLM struct{}

const (
	vllm_image_registry = "vllm/vllm-openai"
)

func (v *VLLM) Name() inferenceapi.BackendName {
	return inferenceapi.VLLM
}

func (v *VLLM) Image(version string) string {
	return vllm_image_registry + ":" + version
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
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		},
	}
}

func (v *VLLM) DefaultCommands() []string {
	return []string{"python3", "-m", "vllm.entrypoints.openai.api_server"}
}

func (v *VLLM) Args(models []*coreapi.OpenModel, mode coreapi.InferenceMode) []string {
	if mode == coreapi.Standard {
		return v.defaultArgs(models[0])
	}
	if mode == coreapi.SpeculativeDecoding {
		return v.speculativeArgs(models)
	}
	// We should not reach here.
	return nil
}

func (v *VLLM) defaultArgs(model *coreapi.OpenModel) []string {
	source := modelSource.NewModelSourceProvider(model)
	return []string{
		"--model", source.ModelPath(),
		"--served-model-name", source.ModelName(),
		"--host", "0.0.0.0",
		"--port", strconv.Itoa(DEFAULT_BACKEND_PORT),
	}
}

func (v *VLLM) speculativeArgs(models []*coreapi.OpenModel) []string {
	targetModelSource := modelSource.NewModelSourceProvider(models[0])
	draftModelSource := modelSource.NewModelSourceProvider(models[1])
	return []string{
		"--model", targetModelSource.ModelPath(),
		"--speculative_model", draftModelSource.ModelPath(),
		"--served-model-name", targetModelSource.ModelName(),
		"--host", "0.0.0.0",
		"--port", strconv.Itoa(DEFAULT_BACKEND_PORT),
	}
}
