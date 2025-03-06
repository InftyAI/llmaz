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

package util

import (
	"k8s.io/utils/ptr"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

const (
	sampleModelName = "llama3-8b"
)

func MockASampleModel() *coreapi.OpenModel {
	return wrapper.MakeModel(sampleModelName).FamilyName("llama3").
		ModelSourceWithModelHub("Huggingface").
		ModelSourceWithModelID("meta-llama/Meta-Llama-3-8B", "", "", nil, nil).
		InferenceFlavors(
			*wrapper.MakeFlavor("a100").SetRequest("nvidia.com/gpu", "1").Obj(),
			*wrapper.MakeFlavor("a10").SetRequest("nvidia.com/gpu", "2").Obj()).
		Obj()
}

func MockASamplePlayground(ns string) *inferenceapi.Playground {
	return wrapper.MakePlayground("playground-llama3-8b", ns).ModelClaim(sampleModelName).Label(coreapi.ModelNameLabelKey, sampleModelName).Obj()
}

func MockASampleService(ns string) *inferenceapi.Service {
	return wrapper.MakeService("service-llama3-8b", ns).
		ModelClaims([]string{sampleModelName}, []string{"main"}).
		WorkerTemplate().
		Obj()
}

func MockASampleBackendRuntime() *wrapper.BackendRuntimeWrapper {
	return wrapper.MakeBackendRuntime("vllm").
		Image("vllm/vllm-openai").Version(VllmImageVersion).
		Command([]string{"python3", "-m", "vllm.entrypoints.openai.api_server"}).
		Arg("default", []string{"--model", "{{.ModelPath}}", "--served-model-name", "{{.ModelName}}", "--host", "0.0.0.0", "--port", "8080"}).
		Request("default", "cpu", "4").Limit("default", "cpu", "4")
}

func MockASimpleHPATrigger() *inferenceapi.HPATrigger {
	return &inferenceapi.HPATrigger{
		Metrics: []autoscalingv2.MetricSpec{
			{
				Type: autoscalingv2.ResourceMetricSourceType,
				Resource: &autoscalingv2.ResourceMetricSource{
					Name: corev1.ResourceCPU,
					Target: autoscalingv2.MetricTarget{
						Type:               autoscalingv2.UtilizationMetricType,
						AverageUtilization: ptr.To[int32](50),
					},
				},
			},
		},
	}
}
