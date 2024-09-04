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

var _ Backend = (*SGLANG)(nil)

type SGLANG struct{}

const (
	sglang_image_registry = "lmsysorg/sglang"
)

func (s *SGLANG) Name() inferenceapi.BackendName {
	return inferenceapi.SGLANG
}

func (s *SGLANG) Image(version string) string {
	return sglang_image_registry + ":" + version
}

func (s *SGLANG) DefaultVersion() string {
	return "v0.2.10-cu121"
}

func (s *SGLANG) DefaultResources() inferenceapi.ResourceRequirements {
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

func (s *SGLANG) Command() []string {
	return []string{"python3", "-m", "sglang.launch_server"}
}

func (s *SGLANG) Args(models []*coreapi.OpenModel, mode coreapi.InferenceMode) []string {
	if mode == coreapi.Standard {
		return s.defaultArgs(models[0])
	}
	// We should not reach here.
	return nil
}

func (s *SGLANG) defaultArgs(model *coreapi.OpenModel) []string {
	source := modelSource.NewModelSourceProvider(model)
	return []string{
		"--model-path", source.ModelPath(),
		"--served-model-name", source.ModelName(),
		"--host", "0.0.0.0",
		"--port", strconv.Itoa(DEFAULT_BACKEND_PORT),
	}
}
