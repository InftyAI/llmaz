/*
Copyright 2025 The InftyAI Team.

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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGlobalConfigs_validate(t *testing.T) {
	tests := []struct {
		name        string
		config      *GlobalConfigs
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &GlobalConfigs{
				SchedulerName:      "custom-scheduler",
				InitContainerImage: "inftyai/model-loader:v0.0.10",
			},
			expectError: false,
		},
		{
			name: "empty init container image",
			config: &GlobalConfigs{
				SchedulerName:      "custom-scheduler",
				InitContainerImage: "",
			},
			expectError: true,
			errorMsg:    "init-container-image is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()

			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
