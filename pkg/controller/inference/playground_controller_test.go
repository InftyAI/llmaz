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

package inference

import (
	"testing"

	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/test/util"
	"inftyai.com/llmaz/test/util/wrapper"
)

func TestModelIdentifiers(t *testing.T) {
	testCases := []struct {
		name          string
		model         *coreapi.Model
		wantModelID   string
		wantModelName string
	}{
		{
			name:          "meta-llama/meta-llama-3-8b",
			model:         util.MockASampleModel(),
			wantModelID:   "meta-llama/meta-llama-3-8b",
			wantModelName: "meta-llama--meta-llama-3-8b",
		},
		{
			name:          "meta-llama/meta-llama-3-8b/test",
			model:         wrapper.MakeModel("test").FamilyName("llama3").DataSourceWithModelID("meta-llama/meta-llama-3-8b/test").Obj(),
			wantModelID:   "meta-llama/meta-llama-3-8b/test",
			wantModelName: "meta-llama--meta-llama-3-8b--test",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotModelID, gotModelName := modelIdentifiers(tc.model)
			if gotModelID != tc.wantModelID {
				t.Fatalf("unexpected modelID, want %s, got %s", tc.wantModelID, gotModelID)
			}
			if gotModelName != tc.wantModelName {
				t.Fatalf("unexpected modelName, want %s, got %s", tc.wantModelName, gotModelName)
			}
		})
	}
}
