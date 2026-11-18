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

package webhook

import (
	"strings"
	"testing"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

func TestPlaygroundValidateMainModelCount(t *testing.T) {
	testCases := []struct {
		name          string
		playground    *inferenceapi.Playground
		wantErr       bool
		wantErrSubstr string
	}{
		{
			name: "single main model is valid",
			playground: wrapper.MakePlayground("playground", "default").
				ModelClaims([]string{"llama3-8b"}, []string{"main"}).Obj(),
			wantErr: false,
		},
		{
			name: "more than one main model is forbidden",
			playground: wrapper.MakePlayground("playground", "default").
				ModelClaims([]string{"llama3-8b", "llama3-70b"}, []string{"main", "main"}).Obj(),
			wantErr:       true,
			wantErrSubstr: "only one main model is allowed",
		},
	}

	webhook := &PlaygroundWebhook{}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := webhook.generateValidate(tc.playground).ToAggregate()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected a validation error, but got none")
				}
				if !strings.Contains(err.Error(), tc.wantErrSubstr) {
					t.Fatalf("expected error to contain %q, but got %q", tc.wantErrSubstr, err.Error())
				}
			} else if err != nil {
				t.Fatalf("expected no validation error, but got %q", err.Error())
			}
		})
	}
}
