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
