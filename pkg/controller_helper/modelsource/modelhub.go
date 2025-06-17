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

package modelSource

import (
	"strconv"
	"strings"

	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"

	"github.com/inftyai/llmaz/pkg"
)

var _ ModelSourceProvider = &ModelHubProvider{}

type ModelHubProvider struct {
	modelName           string
	modelID             string
	modelHub            string
	fileName            *string
	modelRevision       *string
	modelAllowPatterns  []string
	modelIgnorePatterns []string
}

func (p *ModelHubProvider) ModelName() string {
	return p.modelName
}

// ModelPath Example 1:
//   - modelID: facebook/opt-125m
//     modelPath: /workspace/models/models--facebook--opt-125m
//
// Example 2:
//   - modelID: Qwen/Qwen2-0.5B-Instruct-GGUF
//     fileName: qwen2-0_5b-instruct-q5_k_m.gguf
//     modelPath: /workspace/models/qwen2-0_5b-instruct-q5_k_m.gguf
func (p *ModelHubProvider) ModelPath(skipModelLoader bool) string {
	// Skip the model loader to allow the inference engine to handle loading models directly from model hub (e.g., Hugging Face, ModelScope).
	// In this case, the model ID should be returned (e.g., facebook/opt-125m).
	if skipModelLoader {
		return p.modelID
	}

	if p.fileName != nil {
		return CONTAINER_MODEL_PATH + *p.fileName
	}
	return CONTAINER_MODEL_PATH + "models--" + strings.ReplaceAll(p.modelID, "/", "--")
}

func (p *ModelHubProvider) InjectModelLoader(template *coreapplyv1.PodTemplateSpecApplyConfiguration, index int) {
	initContainerName := MODEL_LOADER_CONTAINER_NAME
	if index != 0 {
		initContainerName += "-" + strconv.Itoa(index)
	}

	// Handle initContainer.
	initContainer := coreapplyv1.Container().
		WithName(initContainerName).
		WithImage(pkg.LOADER_IMAGE).
		WithVolumeMounts(coreapplyv1.VolumeMount().WithName(MODEL_VOLUME_NAME).WithMountPath(CONTAINER_MODEL_PATH))

	// We have exactly one container in the template.Spec.Containers.
	spreadEnvToInitContainer(template.Spec.Containers[0].Env, initContainer)

	// This is related to the model loader logics which will read the environment when loading models weights.
	initContainer.WithEnv(
		coreapplyv1.EnvVar().WithName("MODEL_SOURCE_TYPE").WithValue(MODEL_SOURCE_MODELHUB),
		coreapplyv1.EnvVar().WithName("MODEL_ID").WithValue(p.modelID),
		coreapplyv1.EnvVar().WithName("MODEL_HUB_NAME").WithValue(p.modelHub))
	if p.fileName != nil {
		initContainer.WithEnv(coreapplyv1.EnvVar().WithName("MODEL_FILENAME").WithValue(*p.fileName))
	}
	if p.modelRevision != nil {
		initContainer.WithEnv(coreapplyv1.EnvVar().WithName("REVISION").WithValue(*p.modelRevision))
	}
	if p.modelAllowPatterns != nil {
		initContainer.WithEnv(coreapplyv1.EnvVar().WithName("MODEL_ALLOW_PATTERNS").WithValue(strings.Join(p.modelAllowPatterns, ",")))
	}
	if p.modelIgnorePatterns != nil {
		initContainer.WithEnv(coreapplyv1.EnvVar().WithName("MODEL_IGNORE_PATTERNS").WithValue(strings.Join(p.modelIgnorePatterns, ",")))
	}

	// Both HUGGING_FACE_HUB_TOKEN and HF_TOKEN works.
	initContainer.WithEnv(
		coreapplyv1.EnvVar().
			WithName(HUGGING_FACE_HUB_TOKEN).
			WithValueFrom(coreapplyv1.EnvVarSource().
				WithSecretKeyRef(coreapplyv1.SecretKeySelector().
					WithName(MODELHUB_SECRET_NAME). // if secret not exists, the env is empty.
					WithKey(HUGGING_FACE_TOKEN_KEY).
					WithOptional(true))),
		coreapplyv1.EnvVar().
			WithName(HUGGING_FACE_TOKEN_KEY).
			WithValueFrom(coreapplyv1.EnvVarSource().
				WithSecretKeyRef(coreapplyv1.SecretKeySelector().
					WithName(MODELHUB_SECRET_NAME).
					WithKey(HUGGING_FACE_TOKEN_KEY).
					WithOptional(true))))

	template.Spec.WithInitContainers(initContainer)
}

func spreadEnvToInitContainer(containerEnv []coreapplyv1.EnvVarApplyConfiguration, initContainer *coreapplyv1.ContainerApplyConfiguration) {
	for i := range containerEnv {
		initContainer.WithEnv(&containerEnv[i])
	}
}

func (p *ModelHubProvider) InjectModelEnvVars(template *coreapplyv1.PodTemplateSpecApplyConfiguration) {
	for i := range template.Spec.Containers {
		if *template.Spec.Containers[i].Name == MODEL_RUNNER_CONTAINER_NAME {
			// Check if HuggingFace token environment variables already exist
			hfHubTokenExists := false
			hfTokenExists := false
			for _, env := range template.Spec.Containers[i].Env {
				if *env.Name == HUGGING_FACE_HUB_TOKEN {
					hfHubTokenExists = true
				}
				if *env.Name == HUGGING_FACE_TOKEN_KEY {
					hfTokenExists = true
				}
			}

			// Add HUGGING_FACE_HUB_TOKEN if it doesn't exist
			if !hfHubTokenExists {
				template.Spec.Containers[i].WithEnv(
					coreapplyv1.EnvVar().
						WithName(HUGGING_FACE_HUB_TOKEN).
						WithValueFrom(coreapplyv1.EnvVarSource().
							WithSecretKeyRef(coreapplyv1.SecretKeySelector().
								WithName(MODELHUB_SECRET_NAME). // if secret not exists, the env is empty.
								WithKey(HUGGING_FACE_TOKEN_KEY).
								WithOptional(true))))
			}

			// Add HF_TOKEN if it doesn't exist
			if !hfTokenExists {
				template.Spec.Containers[i].WithEnv(
					coreapplyv1.EnvVar().
						WithName(HUGGING_FACE_TOKEN_KEY).
						WithValueFrom(coreapplyv1.EnvVarSource().
							WithSecretKeyRef(coreapplyv1.SecretKeySelector().
								WithName(MODELHUB_SECRET_NAME).
								WithKey(HUGGING_FACE_TOKEN_KEY).
								WithOptional(true))))
			}
		}
	}
}
