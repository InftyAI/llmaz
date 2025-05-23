/*
Copyright 2025 The InftyAI Team.

Licensed under the Apache License,
 Version 2.0 (the "License");
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

	"github.com/inftyai/llmaz/pkg"
)

func Test_ModelHubProvider_InjectModelLoader(t *testing.T) {
	fileName := "weights.gguf"
	revision := "v1.2"
	allowPatterns := []string{"*.gguf", "*.json"}
	ignorePatterns := []string{"*.tmp"}

	testCases := []struct {
		name              string
		provider          *ModelHubProvider
		index             int
		expectMainModel   bool
		expectEnvContains []string
	}{
		{
			name: "inject full modelhub with fileName, revision, allow/ignore",
			provider: &ModelHubProvider{
				modelName:           "llama3",
				modelID:             "meta/llama-3",
				modelHub:            "Huggingface",
				fileName:            &fileName,
				modelRevision:       &revision,
				modelAllowPatterns:  allowPatterns,
				modelIgnorePatterns: ignorePatterns,
			},
			index:           0,
			expectMainModel: true,
			expectEnvContains: []string{
				"MODEL_SOURCE_TYPE", "MODEL_ID", "MODEL_HUB_NAME", "MODEL_FILENAME",
				"REVISION", "MODEL_ALLOW_PATTERNS", "MODEL_IGNORE_PATTERNS",
				"HUGGING_FACE_HUB_TOKEN", "HF_TOKEN",
			},
		},
		{
			name: "inject with index > 0 skips volume/container mount",
			provider: &ModelHubProvider{
				modelName: "sub-model",
				modelID:   "some/model",
				modelHub:  "Huggingface",
			},
			index:           1,
			expectMainModel: false,
			expectEnvContains: []string{
				"MODEL_SOURCE_TYPE", "MODEL_ID", "MODEL_HUB_NAME",
				"HUGGING_FACE_HUB_TOKEN", "HF_TOKEN",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			template := &corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: MODEL_RUNNER_CONTAINER_NAME,
							Env: []corev1.EnvVar{
								{Name: "HTTP_PROXY", Value: "http://1.1.1.1"},
							},
						},
					},
				},
			}

			tc.provider.InjectModelLoader(template, tc.index)

			assert.Len(t, template.Spec.InitContainers, 1)
			initContainer := template.Spec.InitContainers[0]
			expectedName := MODEL_LOADER_CONTAINER_NAME
			if tc.index != 0 {
				expectedName += "-" + string(rune('0'+tc.index))
			}
			assert.Equal(t, expectedName, initContainer.Name)
			assert.Equal(t, pkg.LOADER_IMAGE, initContainer.Image)

			// Check env vars exist
			for _, key := range tc.expectEnvContains {
				found := false
				for _, env := range initContainer.Env {
					if env.Name == key {
						found = true
						break
					}
				}
				assert.True(t, found, "expected env %s not found", key)
			}

			// Main model should inject volume & container mount
			if tc.expectMainModel {
				// Volume should be present
				foundVol := false
				for _, v := range template.Spec.Volumes {
					if v.Name == MODEL_VOLUME_NAME {
						foundVol = true
						break
					}
				}
				assert.True(t, foundVol, "volume not injected")

				// Runner container mount should exist
				foundMount := false
				for _, m := range template.Spec.Containers[0].VolumeMounts {
					if m.Name == MODEL_VOLUME_NAME && m.ReadOnly && m.MountPath == CONTAINER_MODEL_PATH {
						foundMount = true
					}
				}
				assert.True(t, foundMount, "volume mount not injected to runner")
			} else {
				// No volumes or mounts should be injected
				assert.Empty(t, template.Spec.Volumes)
				assert.Empty(t, template.Spec.Containers[0].VolumeMounts)
			}

			// Should always carry over container env
			assert.Contains(t, initContainer.Env, corev1.EnvVar{Name: "HTTP_PROXY", Value: "http://1.1.1.1"})
		})
	}
}
