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

	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
)

func TestSwitchBackend(t *testing.T) {
	testCases := []struct {
		name                string
		backendName         inferenceapi.BackendName
		expectedBackendName inferenceapi.BackendName
		shouldErr           bool
	}{
		{
			name:                "vllm should support",
			backendName:         "vllm",
			expectedBackendName: inferenceapi.VLLM,
			shouldErr:           false,
		},
		{
			name:                "sglang should support",
			backendName:         "sglang",
			expectedBackendName: inferenceapi.SGLANG,
			shouldErr:           false,
		},
		{
			name:                "llamacpp should support",
			backendName:         "llamacpp",
			expectedBackendName: inferenceapi.LLAMACPP,
			shouldErr:           false,
		},
		{
			name:                "tgi should not support",
			backendName:         "tgi",
			expectedBackendName: inferenceapi.BackendName(""),
			shouldErr:           true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			backend := SwitchBackend(tc.backendName)

			if !tc.shouldErr && backend == nil {
				t.Fatal("unexpected error")
			}

			if !tc.shouldErr && backend.Name() != tc.expectedBackendName {
				t.Fatalf("unexpected backend, want %s, got %s", tc.expectedBackendName, backend.Name())
			}
		})
	}
}
