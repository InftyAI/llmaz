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

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

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

func (p *ModelHubProvider) InjectModelLoader(template *corev1.PodTemplateSpec, index int) {
	initContainerName := MODEL_LOADER_CONTAINER_NAME
	if index != 0 {
		initContainerName += "-" + strconv.Itoa(index)
	}

	// Handle initContainer.
	initContainer := &corev1.Container{
		Name:  initContainerName,
		Image: pkg.LOADER_IMAGE,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      MODEL_VOLUME_NAME,
				MountPath: CONTAINER_MODEL_PATH,
			},
		},
	}

	// We have exactly one container in the template.Spec.Containers.
	spreadEnvToInitContainer(template.Spec.Containers[0].Env, initContainer)

	// This is related to the model loader logics which will read the environment when loading models weights.
	initContainer.Env = append(
		initContainer.Env,
		corev1.EnvVar{Name: "MODEL_SOURCE_TYPE", Value: MODEL_SOURCE_MODELHUB},
		corev1.EnvVar{Name: "MODEL_ID", Value: p.modelID},
		corev1.EnvVar{Name: "MODEL_HUB_NAME", Value: p.modelHub},
	)
	if p.fileName != nil {
		initContainer.Env = append(initContainer.Env,
			corev1.EnvVar{Name: "MODEL_FILENAME", Value: *p.fileName})
	}
	if p.modelRevision != nil {
		initContainer.Env = append(initContainer.Env,
			corev1.EnvVar{Name: "REVISION", Value: *p.modelRevision},
		)
	}
	if p.modelAllowPatterns != nil {
		initContainer.Env = append(initContainer.Env,
			corev1.EnvVar{Name: "MODEL_ALLOW_PATTERNS", Value: strings.Join(p.modelAllowPatterns, ",")},
		)
	}
	if p.modelIgnorePatterns != nil {
		initContainer.Env = append(initContainer.Env,
			corev1.EnvVar{Name: "MODEL_IGNORE_PATTERNS", Value: strings.Join(p.modelIgnorePatterns, ",")},
		)
	}

	// Both HUGGING_FACE_HUB_TOKEN and HF_TOKEN works.
	initContainer.Env = append(initContainer.Env,
		corev1.EnvVar{
			Name: HUGGING_FACE_HUB_TOKEN,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: MODELHUB_SECRET_NAME, // if secret not exists, the env is empty.
					},
					Key:      HUGGING_FACE_TOKEN_KEY,
					Optional: ptr.To[bool](true),
				},
			},
		})

	initContainer.Env = append(initContainer.Env,
		corev1.EnvVar{
			Name: HUGGING_FACE_TOKEN_KEY,
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: MODELHUB_SECRET_NAME,
					},
					Key:      HUGGING_FACE_TOKEN_KEY,
					Optional: ptr.To[bool](true),
				},
			},
		})

	template.Spec.InitContainers = append(template.Spec.InitContainers, *initContainer)
}

func spreadEnvToInitContainer(containerEnv []corev1.EnvVar, initContainer *corev1.Container) {
	// Filter out credential-related environment variables to avoid duplicates
	// For example, when main model doesn't use model-loader and handle the model loading itself -- in this case, the HUGGING_FACE_HUB_TOKEN and HF_TOKEN environment variables should be set in the model-runner container.
	// While draft model uses model-loader, spreadEnvToInitContainer shouldn't copy HUGGING_FACE_HUB_TOKEN and HF_TOKEN environment variables from model-runner container to model-loader initContainer.
	excludedEnvs := map[string]struct{}{
		// HuggingFace credentials
		HUGGING_FACE_HUB_TOKEN: {},
		HUGGING_FACE_TOKEN_KEY: {},
		// AWS/S3 credentials
		AWS_ACCESS_KEY_ID:     {},
		AWS_ACCESS_KEY_SECRET: {},
		// OSS credentials
		OSS_ACCESS_KEY_ID:     {},
		OSS_ACCESS_KEY_SECRET: {},
	}

	for _, env := range containerEnv {
		if _, excluded := excludedEnvs[env.Name]; !excluded {
			initContainer.Env = append(initContainer.Env, env)
		}
	}
}

func (p *ModelHubProvider) InjectModelEnvVars(template *corev1.PodTemplateSpec) {
	for i := range template.Spec.Containers {
		if template.Spec.Containers[i].Name == MODEL_RUNNER_CONTAINER_NAME {
			// Check if HuggingFace token environment variables already exist
			hfHubTokenExists := false
			hfTokenExists := false
			for _, env := range template.Spec.Containers[i].Env {
				if env.Name == HUGGING_FACE_HUB_TOKEN {
					hfHubTokenExists = true
				}
				if env.Name == HUGGING_FACE_TOKEN_KEY {
					hfTokenExists = true
				}
			}

			// Add HUGGING_FACE_HUB_TOKEN if it doesn't exist
			if !hfHubTokenExists {
				template.Spec.Containers[i].Env = append(template.Spec.Containers[i].Env,
					corev1.EnvVar{
						Name: HUGGING_FACE_HUB_TOKEN,
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: MODELHUB_SECRET_NAME, // if secret not exists, the env is empty.
								},
								Key:      HUGGING_FACE_TOKEN_KEY,
								Optional: ptr.To[bool](true),
							},
						},
					})
			}

			// Add HF_TOKEN if it doesn't exist
			if !hfTokenExists {
				template.Spec.Containers[i].Env = append(template.Spec.Containers[i].Env,
					corev1.EnvVar{
						Name: HUGGING_FACE_TOKEN_KEY,
						ValueFrom: &corev1.EnvVarSource{
							SecretKeyRef: &corev1.SecretKeySelector{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: MODELHUB_SECRET_NAME,
								},
								Key:      HUGGING_FACE_TOKEN_KEY,
								Optional: ptr.To[bool](true),
							},
						},
					})
			}
		}
	}
}
