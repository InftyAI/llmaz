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

package modelSource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

func Test_ModelSourceProvider(t *testing.T) {
	testCases := []struct {
		name          string
		model         *coreapi.OpenModel
		wantModelName string
		wantModelPath string
	}{
		{
			name:          "model with modelhub configured",
			model:         util.MockASampleModel(),
			wantModelName: "llama3-8b",
			wantModelPath: "/workspace/models/models--meta-llama--Meta-Llama-3-8B",
		},
		{
			name:          "modelhub with GGUF file",
			model:         wrapper.MakeModel("test-7b").FamilyName("test").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", nil, nil).Obj(),
			wantModelName: "test-7b",
			wantModelPath: "/workspace/models/qwen2-0_5b-instruct-q5_k_m.gguf",
		},
		{
			name:          "model with URI configured",
			model:         wrapper.MakeModel("test-7b").FamilyName("test").ModelSourceWithURI("oss://bucket.endpoint/modelPath/subPath").Obj(),
			wantModelName: "test-7b",
			wantModelPath: "/workspace/models/models--subPath",
		},
		{
			name:          "URI with GGUF model",
			model:         wrapper.MakeModel("test-7b").FamilyName("test").ModelSourceWithURI("oss://bucket.endpoint/modelPath/weight.gguf").Obj(),
			wantModelName: "test-7b",
			wantModelPath: "/workspace/models/weight.gguf",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := NewModelSourceProvider(tc.model)
			if tc.wantModelName != provider.ModelName() {
				t.Fatalf("unexpected model name, want %s, got %s", tc.wantModelName, provider.ModelName())
			}
			if tc.wantModelPath != provider.ModelPath() {
				t.Fatalf("unexpected model path, want %s, got %s", tc.wantModelPath, provider.ModelPath())
			}
		})
	}
}

func TestEnvInjectModelLoader(t *testing.T) {
	tests := []struct {
		name     string
		provider ModelSourceProvider
		template *corev1.PodTemplateSpec
	}{
		{
			name: "Spread container env to initContiner using modelhub",
			provider: &ModelHubProvider{
				modelName: "test-model",
			},
			template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "model-runner",
							Image: "vllm:test",
							Env: []corev1.EnvVar{
								{Name: "http_proxy", Value: "1.1.1.1:1234"},
								{Name: "https_proxy", Value: "1.1.1.1:1234"},
							},
						},
					},
				},
			},
		},
		{
			name: "Spread container env to initContiner using objstores",
			provider: &URIProvider{
				modelName: "test-model",
			},
			template: &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "model-runner",
							Image: "vllm:test",
							Env: []corev1.EnvVar{
								{Name: "http_proxy", Value: "1.1.1.1:1234"},
								{Name: "https_proxy", Value: "1.1.1.1:1234"},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.provider.InjectModelLoader(tt.template, 0)
			initContainer := tt.template.Spec.InitContainers[0]
			assert.Subset(t, initContainer.Env, tt.template.Spec.Containers[0].Env)
		})
	}
}
