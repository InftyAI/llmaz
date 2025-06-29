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
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"
)

func Test_ModelHubProvider_InjectModelLoader(t *testing.T) {
	fileName := "weights.gguf"
	revision := "v1.2"
	allowPatterns := []string{"*.gguf", "*.json"}
	ignorePatterns := []string{"*.tmp"}

	tests := []struct {
		name               string
		provider           *ModelHubProvider
		index              int
		expectMainModel    bool
		initContainerImage string
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
			index:              0,
			expectMainModel:    true,
			initContainerImage: "model-loader:latest",
		},
		{
			name: "inject with index > 0 skips volume/container mount",
			provider: &ModelHubProvider{
				modelName: "sub-model",
				modelID:   "some/model",
				modelHub:  "Huggingface",
			},
			index:              1,
			expectMainModel:    false,
			initContainerImage: "model-loader:latest",
		},
		{
			name: "inject with custom initContainerImage",
			provider: &ModelHubProvider{
				modelName: "llama3",
				modelID:   "meta/llama-3",
				modelHub:  "Huggingface",
			},
			index:              0,
			expectMainModel:    true,
			initContainerImage: "custom-model-loader:latest",
		},
	}

	envSortOpt := cmpopts.SortSlices(func(a, b corev1.EnvVar) bool {
		return a.Name < b.Name
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			template := coreapplyv1.PodTemplateSpec().
				WithSpec(coreapplyv1.PodSpec().
					WithContainers(coreapplyv1.Container().
						WithName(MODEL_RUNNER_CONTAINER_NAME).
						WithEnv(coreapplyv1.EnvVar().WithName("HTTP_PROXY").WithValue("http://1.1.1.1")),
					),
				)

			tt.provider.InjectModelLoader(template, tt.index, tt.initContainerImage)

			assert.Len(t, template.Spec.InitContainers, 1)
			initContainer := template.Spec.InitContainers[0]

			expectedName := MODEL_LOADER_CONTAINER_NAME
			if tt.index != 0 {
				expectedName += "-" + strconv.Itoa(tt.index)
			}
			expectedImage := tt.initContainerImage
			assert.Equal(t, expectedName, *initContainer.Name)
			assert.Equal(t, expectedImage, *initContainer.Image)

			wantEnv := buildExpectedEnv(tt.provider)
			if diff := cmp.Diff(wantEnv, initContainer.Env, envSortOpt); diff != "" {
				t.Errorf("InitContainer.Env mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func buildExpectedEnv(p *ModelHubProvider) []coreapplyv1.EnvVarApplyConfiguration {
	envs := make([]coreapplyv1.EnvVarApplyConfiguration, 0, 10)

	envs = append(envs, *coreapplyv1.EnvVar().WithName("HTTP_PROXY").WithValue("http://1.1.1.1"))

	envs = append(envs,
		*coreapplyv1.EnvVar().WithName("MODEL_SOURCE_TYPE").WithValue(MODEL_SOURCE_MODELHUB),
		*coreapplyv1.EnvVar().WithName("MODEL_ID").WithValue(p.modelID),
		*coreapplyv1.EnvVar().WithName("MODEL_HUB_NAME").WithValue(p.modelHub),
	)

	if p.fileName != nil {
		envs = append(envs, *coreapplyv1.EnvVar().WithName("MODEL_FILENAME").WithValue(*p.fileName))
	}
	if p.modelRevision != nil {
		envs = append(envs, *coreapplyv1.EnvVar().WithName("REVISION").WithValue(*p.modelRevision))
	}
	if p.modelAllowPatterns != nil {
		envs = append(envs, *coreapplyv1.EnvVar().WithName("MODEL_ALLOW_PATTERNS").WithValue(strings.Join(p.modelAllowPatterns, ",")))
	}
	if p.modelIgnorePatterns != nil {
		envs = append(envs, *coreapplyv1.EnvVar().WithName("MODEL_IGNORE_PATTERNS").WithValue(strings.Join(p.modelIgnorePatterns, ",")))
	}

	for _, tokenName := range []string{"HUGGING_FACE_HUB_TOKEN", "HF_TOKEN"} {
		envs = append(envs, *coreapplyv1.EnvVar().WithName(tokenName).WithValueFrom(coreapplyv1.EnvVarSource().WithSecretKeyRef(coreapplyv1.SecretKeySelector().WithName(MODELHUB_SECRET_NAME).WithKey(HUGGING_FACE_TOKEN_KEY).WithOptional(true))))
	}

	return envs
}
