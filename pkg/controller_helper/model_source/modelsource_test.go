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

package modelSource

import (
	"testing"

	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/test/util"
)

func Test_ModelSourceProvider(t *testing.T) {
	testCases := []struct {
		name          string
		model         *coreapi.OpenModel
		wantModelName string
		wantModelPath string
	}{
		{
			name:          "model with model hub configured",
			model:         util.MockASampleModel(),
			wantModelName: "llama3-8b",
			wantModelPath: "/workspace/models/models--meta-llama--Meta-Llama-3-8B",
		},
		// {
		// 	name:          "model with URI configured",
		// 	model:         wrapper.MakeModel("test-7b").FamilyName("test").DataSourceWithURI("s3://a/b/c").Obj(),
		// 	wantModelName: "test-7b",
		// 	wantModelPath: "/workspace/models/",
		// },
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
