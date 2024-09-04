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
	"testing"

	"github.com/google/go-cmp/cmp"
	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

func Test_vllm(t *testing.T) {
	backend := VLLM{}
	models := []*coreapi.OpenModel{
		{
			ObjectMeta: v1.ObjectMeta{
				Name: "model-1",
			},
			Spec: coreapi.ModelSpec{
				Source: coreapi.ModelSource{
					ModelHub: &coreapi.ModelHub{
						Name:    ptr.To[string]("model-1"),
						ModelID: "hub/model-1",
					},
				},
			},
		},
		{
			ObjectMeta: v1.ObjectMeta{
				Name: "model-2",
			},
			Spec: coreapi.ModelSpec{
				Source: coreapi.ModelSource{
					ModelHub: &coreapi.ModelHub{
						Name:    ptr.To[string]("model-2"),
						ModelID: "hub/model-2",
					},
				},
			},
		},
	}

	testCases := []struct {
		name        string
		mode        coreapi.InferenceMode
		wantCommand []string
		wantArgs    []string
	}{
		{
			name:        "standard mode",
			mode:        coreapi.Standard,
			wantCommand: []string{"python3", "-m", "vllm.entrypoints.openai.api_server"},
			wantArgs: []string{
				"--model", "/workspace/models/models--hub--model-1",
				"--served-model-name", "model-1",
				"--host", "0.0.0.0",
				"--port", "8080",
			},
		},
		{
			name:        "speculative decoding",
			mode:        coreapi.SpeculativeDecoding,
			wantCommand: []string{"python3", "-m", "vllm.entrypoints.openai.api_server"},
			wantArgs: []string{
				"--model", "/workspace/models/models--hub--model-1",
				"--speculative_model", "/workspace/models/models--hub--model-2",
				"--served-model-name", "model-1",
				"--host", "0.0.0.0",
				"--port", "8080",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if diff := cmp.Diff(backend.Command(), tc.wantCommand); diff != "" {
				t.Fatalf("unexpected command, want %v, got %v", tc.wantCommand, backend.Command())
			}
			if diff := cmp.Diff(backend.Args(models, tc.mode), tc.wantArgs); diff != "" {
				t.Fatalf("unexpected args, want %v, got %v", tc.wantArgs, backend.Args(models, tc.mode))
			}
		})
	}
}
