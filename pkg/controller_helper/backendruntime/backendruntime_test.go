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

package helper

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	"github.com/inftyai/llmaz/test/util/wrapper"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
)

func TestRenderFlags(t *testing.T) {
	testCases := []struct {
		name      string
		flags     []string
		modelInfo map[string]string
		wantFlags []string
		wantError bool
	}{
		{
			name:  "normal parse long args",
			flags: []string{"run {{ .ModelPath }};sleep 5", "--host", "0.0.0.0"},
			modelInfo: map[string]string{
				"ModelPath": "path/to/model",
				"ModelName": "foo",
			},
			wantFlags: []string{"run path/to/model;sleep 5", "--host", "0.0.0.0"},
		},
		{
			name:  "normal parse",
			flags: []string{"-m", "{{ .ModelPath }}", "--served-model-name", "{{ .ModelName }}", "--host", "0.0.0.0"},
			modelInfo: map[string]string{
				"ModelPath": "path/to/model",
				"ModelName": "foo",
			},
			wantFlags: []string{"-m", "path/to/model", "--served-model-name", "foo", "--host", "0.0.0.0"},
		},
		{
			name:  "miss some info",
			flags: []string{"-m", "{{ .ModelPath }}", "--served-model-name", "{{ .ModelName }}", "--host", "0.0.0.0"},
			modelInfo: map[string]string{
				"ModelPath": "path/to/model",
			},
			wantError: true,
		},
		{
			name:  "missing . with flag",
			flags: []string{"-m", "{{ ModelPath }}", "--served-model-name", "{{ .ModelName }}", "--host", "0.0.0.0"},
			modelInfo: map[string]string{
				"ModelPath": "path/to/model",
				"ModelName": "foo",
			},
			wantFlags: []string{"-m", "{{ ModelPath }}", "--served-model-name", "foo", "--host", "0.0.0.0"},
		},
		{
			name:  "no empty space between {{}}",
			flags: []string{"-m", "{{.ModelPath}}", "--served-model-name", "{{.ModelName}}", "--host", "0.0.0.0"},
			modelInfo: map[string]string{
				"ModelPath": "path/to/model",
				"ModelName": "foo",
			},
			wantFlags: []string{"-m", "path/to/model", "--served-model-name", "foo", "--host", "0.0.0.0"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotFlags, err := renderFlags(tc.flags, tc.modelInfo)
			if tc.wantError && err == nil {
				t.Fatal("test should fail")
			}

			if !tc.wantError && cmp.Diff(tc.wantFlags, gotFlags) != "" {
				t.Fatalf("want flags %v, got flags %v", tc.wantFlags, gotFlags)
			}
		})
	}
}

func TestBackendRuntimeParser_BasicFields(t *testing.T) {
	type want struct {
		cmd       []string
		args      []string
		envs      []corev1.EnvVar
		lifecycle *corev1.Lifecycle
		image     string
		version   string
		resources *inferenceapi.ResourceRequirements
		shm       *resource.Quantity
	}

	testCases := []struct {
		name   string
		parser *BackendRuntimeParser
		want   want
	}{
		{
			name: "normal case has recommendConfig",
			parser: func() *BackendRuntimeParser {
				cmd := []string{"python", "serve.py"}
				args := []string{"--port=8080"}
				envs := []corev1.EnvVar{{Name: "MODE", Value: "release"}}
				lc := &corev1.Lifecycle{
					PostStart: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{Command: []string{"echo", "started"}},
					},
				}
				shm := resource.NewQuantity(1*1024*1024*1024, resource.BinarySI)
				res := &inferenceapi.ResourceRequirements{}

				backend := &inferenceapi.BackendRuntime{
					ObjectMeta: metav1.ObjectMeta{Name: "default"},
					Spec: inferenceapi.BackendRuntimeSpec{
						Command:   cmd,
						Envs:      envs,
						Lifecycle: lc,
						Image:     "inftyai/llama",
						Version:   "v0.1.0",
						RecommendedConfigs: []inferenceapi.RecommendedConfig{
							{
								Name:             "default",
								Args:             args,
								Resources:        res,
								SharedMemorySize: shm,
							},
						},
					},
				}

				models := []*coreapi.OpenModel{}
				qwenModel := wrapper.MakeModel("qwen2-0--5b").FamilyName("qwen2").ModelSourceWithURI("ollama://qwen2:0.5b").Obj()
				models = append(models, qwenModel)

				playground := wrapper.MakePlayground("qwen2-0--5b", corev1.NamespaceDefault).ModelClaim("qwen2-0--5b").BackendRuntime("llmaz-ollama").BackendRuntimeEnv("OLLAMA_HOST", "0.0.0.0:8080").Replicas(1).Obj()

				return NewBackendRuntimeParser(backend, models, playground)
			}(),
			want: want{
				cmd:  []string{"python", "serve.py"},
				args: []string{"--port=8080"},
				envs: []corev1.EnvVar{{Name: "MODE", Value: "release"}},
				lifecycle: &corev1.Lifecycle{
					PostStart: &corev1.LifecycleHandler{
						Exec: &corev1.ExecAction{Command: []string{"echo", "started"}},
					},
				},
				image:     "inftyai/llama:v0.1.0",
				version:   "v0.1.0",
				resources: &inferenceapi.ResourceRequirements{},
				shm:       resource.NewQuantity(1*1024*1024*1024, resource.BinarySI),
			},
		},
		{
			name: "recommendConfigName not found and resources and SharedMemorySize return nil",
			parser: func() *BackendRuntimeParser {
				backend := &inferenceapi.BackendRuntime{
					ObjectMeta: metav1.ObjectMeta{Name: "default"},
					Spec: inferenceapi.BackendRuntimeSpec{
						Command: []string{"some"},
						Image:   "repo/img",
						Version: "latest",
						RecommendedConfigs: []inferenceapi.RecommendedConfig{
							{
								Name: "default",
							},
						},
					},
				}

				models := []*coreapi.OpenModel{}
				qwenModel := wrapper.MakeModel("qwen2-0--5b").FamilyName("qwen2").ModelSourceWithURI("ollama://qwen2:0.5b").Obj()
				models = append(models, qwenModel)

				playground := wrapper.MakePlayground("qwen2-0--5b", corev1.NamespaceDefault).ModelClaim("qwen2-0--5b").BackendRuntime("llmaz-ollama").BackendRuntimeEnv("OLLAMA_HOST", "0.0.0.0:8080").Replicas(1).Obj()

				return NewBackendRuntimeParser(backend, models, playground)
			}(),
			want: want{
				cmd:       []string{"some"},
				image:     "repo/img:latest",
				version:   "latest",
				args:      []string{},
				resources: nil,
				shm:       nil,
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			p := tc.parser
			if diff := cmp.Diff(tc.want.cmd, p.Command()); diff != "" {
				t.Fatalf("Command() mismatch (-want +got):\n%s", diff)
			}
			args, err := p.Args()
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want.args, args); diff != "" {
				t.Fatalf("Args() mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.envs, p.Envs()); diff != "" {
				t.Fatalf("Envs() mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.lifecycle, p.Lifecycle()); diff != "" {
				t.Fatalf("Lifecycle() mismatch (-want +got):\n%s", diff)
			}

			if got := p.Image(p.Version()); got != tc.want.image {
				t.Fatalf("Image() = %s, want %s", got, tc.want.image)
			}

			if got := p.Version(); got != tc.want.version {
				t.Fatalf("Version() = %s, want %s", got, tc.want.version)
			}

			if diff := cmp.Diff(tc.want.resources, p.Resources()); diff != "" {
				t.Fatalf("Resources() mismatch (-want +got):\n%s", diff)
			}

			if diff := cmp.Diff(tc.want.shm, p.SharedMemorySize()); diff != "" {
				t.Fatalf("SharedMemorySize() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
