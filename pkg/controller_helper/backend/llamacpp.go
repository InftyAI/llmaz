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

var _ Backend = (*LLAMACPP)(nil)
var _ SpeculativeBackend = (*LLAMACPP)(nil)

type LLAMACPP struct{}

const (
	llama_cpp_image_registry = "ghcr.io/ggerganov/llama.cpp"
)

func (l *LLAMACPP) Name() inferenceapi.BackendName {
	return inferenceapi.LLAMACPP
}

func (l *LLAMACPP) Image(version string) string {
	return llama_cpp_image_registry + ":" + version
}

func (l *LLAMACPP) DefaultVersion() string {
	return "server"
}

func (l *LLAMACPP) DefaultResources() inferenceapi.ResourceRequirements {
	return inferenceapi.ResourceRequirements{
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("2"),
			corev1.ResourceMemory: resource.MustParse("4Gi"),
		},
	}
}

func (l *LLAMACPP) Command() []string {
	return []string{"./llama-server"}
}

func (l *LLAMACPP) Args(models []*coreapi.OpenModel, mode coreapi.InferenceMode) []string {
	if mode == coreapi.Standard {
		return l.defaultArgs(models[0])
	}
	if mode == coreapi.SpeculativeDecoding {
		return l.speculativeArgs(models)
	}
	// We should not reach here.
	return nil
}

func (l *LLAMACPP) defaultArgs(model *coreapi.OpenModel) []string {
	source := modelSource.NewModelSourceProvider(model)
	return []string{
		"-m", source.ModelPath(),
		"--port", strconv.Itoa(DEFAULT_BACKEND_PORT),
		"--host", "0.0.0.0",
	}
}

func (l *LLAMACPP) speculativeArgs(models []*coreapi.OpenModel) []string {
	targetModelSource := modelSource.NewModelSourceProvider(models[0])
	draftModelSource := modelSource.NewModelSourceProvider(models[1])
	return []string{
		"-m", targetModelSource.ModelPath(),
		"-md", draftModelSource.ModelPath(),
		"--port", strconv.Itoa(DEFAULT_BACKEND_PORT),
		"--host", "0.0.0.0",
	}
}
