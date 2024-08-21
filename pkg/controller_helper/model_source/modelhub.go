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
	"strings"

	"inftyai.com/llmaz/pkg"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"
)

var _ ModelSourceProvider = &ModelHubProvider{}

type ModelHubProvider struct {
	modelName     string
	modelID       string
	modelHub      string
	fileName      *string
	modelRevision *string
}

func (p *ModelHubProvider) ModelName() string {
	return p.modelName
}

// Example 1:
//   - modelID: facebook/opt-125m
//     modelPath: /workspace/models/models--facebook--opt-125m
//
// Example 2:
//   - modelID: Qwen/Qwen2-0.5B-Instruct-GGUF
//     fileName: qwen2-0_5b-instruct-q5_k_m.gguf
//     modelPath: /workspace/models/qwen2-0_5b-instruct-q5_k_m.gguf
func (p *ModelHubProvider) ModelPath() string {
	if p.fileName != nil {
		return CONTAINER_MODEL_PATH + *p.fileName
	}
	return CONTAINER_MODEL_PATH + "models--" + strings.ReplaceAll(p.modelID, "/", "--")
}

func (p *ModelHubProvider) InjectModelLoader(template *corev1.PodTemplateSpec) {
	// Handle initContainer.
	initContainer := &corev1.Container{
		Name:  MODEL_LOADER_CONTAINER_NAME,
		Image: pkg.LOADER_IMAGE,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      MODEL_VOLUME_NAME,
				MountPath: CONTAINER_MODEL_PATH,
			},
		},
	}

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
	initContainer.Env = append(initContainer.Env,
		corev1.EnvVar{
			Name: "HUGGING_FACE_HUB_TOKEN", // vllm
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: MODELHUB_SECRET_NAME, // if secret not exists, the env is empty.
					},
					Key:      HUGGINGFACE_TOKEN_KEY,
					Optional: ptr.To[bool](true),
				},
			},
		}, corev1.EnvVar{
			Name: "HF_TOKEN", //sglang
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: MODELHUB_SECRET_NAME,
					},
					Key:      HUGGINGFACE_TOKEN_KEY,
					Optional: ptr.To[bool](true),
				},
			},
		},
	)
	template.Spec.InitContainers = append(template.Spec.InitContainers, *initContainer)

	// Handle container.

	for i := range template.Spec.Containers {
		// We only consider this container.
		if template.Spec.Containers[i].Name == MODEL_RUNNER_CONTAINER_NAME {
			template.Spec.Containers[i].VolumeMounts = append(template.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
				Name:      MODEL_VOLUME_NAME,
				MountPath: CONTAINER_MODEL_PATH,
				ReadOnly:  true,
			})
		}
	}

	// Handle spec.

	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: MODEL_VOLUME_NAME,
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	})
}
