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

// python -m vllm.entrypoints.openai.api_server --model NousResearch/Meta-Llama-3-8B-Instruct --dtype auto --api-key token-abc123
//
//	docker run --runtime nvidia --gpus all \
//	    -v ~/.cache/huggingface:/root/.cache/huggingface \
//	    --env "HUGGING_FACE_HUB_TOKEN=<secret>" \
//	    -p 8000:8000 \
//	    --ipc=host \
//	    vllm/vllm-openai:latest \
//	    --model mistralai/Mistral-7B-v0.1

// from transformers import AutoModel

// access_token = "hf_..."

// model = AutoModel.from_pretrained("private/model", token=access_token)
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
