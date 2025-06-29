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
)

var _ ModelSourceProvider = &URIProvider{}

const (
	GCS      = "GCS"
	OSS      = "OSS"
	S3       = "S3"
	Ollama   = "OLLAMA"
	HostPath = "HOST"
)

type URIProvider struct {
	modelName string
	protocol  string
	bucket    string
	endpoint  string
	modelPath string
	uri       string
}

func (p *URIProvider) ModelName() string {
	if p.protocol == Ollama {
		// model path stores the ollama model name,
		// the model name is the name of model CRD.
		return p.modelPath
	}
	return p.modelName
}

// Example 1:
//   - uri: bucket.endpoint/modelPath/opt-125m
//     modelPath: /workspace/models/models--opt-125m
//
// Example 2:
//   - uri: bucket.endpoint/modelPath/model.gguf
//     modelPath: /workspace/models/model.gguf
func (p *URIProvider) ModelPath(skipModelLoader bool) string {
	if p.protocol == HostPath {
		return p.modelPath
	}

	// Skip the model loader to allow the inference engine to handle loading models directly from remote storage (e.g., S3, OSS).
	// In this case, the remote model path should be returned (e.g., s3://bucket/modelPath).
	if skipModelLoader {
		return p.uri
	}

	// protocol is oss.
	splits := strings.Split(p.modelPath, "/")

	if strings.Contains(p.modelPath, ".gguf") {
		return CONTAINER_MODEL_PATH + splits[len(splits)-1]
	}
	return CONTAINER_MODEL_PATH + "models--" + splits[len(splits)-1]
}

func (p *URIProvider) InjectModelLoader(template *coreapplyv1.PodTemplateSpecApplyConfiguration, index int, initContainerImage string) {
	// We don't have additional operations for Ollama, just load in runtime.
	if p.protocol == Ollama {
		return
	}

	if p.protocol == HostPath {
		template.Spec.WithVolumes(
			coreapplyv1.Volume().
				WithName(MODEL_VOLUME_NAME).
				WithHostPath(coreapplyv1.HostPathVolumeSource().
					WithPath(p.modelPath)),
		)

		for i, container := range template.Spec.Containers {
			// We only consider this container.
			if *container.Name == MODEL_RUNNER_CONTAINER_NAME {
				template.Spec.Containers[i].WithVolumeMounts(coreapplyv1.VolumeMount().
					WithName(MODEL_VOLUME_NAME).
					WithMountPath(p.modelPath).
					WithReadOnly(true),
				)
			}
		}
		return
	}

	// Other protocols.
	initContainerName := MODEL_LOADER_CONTAINER_NAME
	if index != 0 {
		initContainerName += "-" + strconv.Itoa(index)
	}

	// Handle initContainer.
	initContainer := coreapplyv1.Container().
		WithName(initContainerName).
		WithImage(initContainerImage).
		WithVolumeMounts(
			coreapplyv1.VolumeMount().
				WithName(MODEL_VOLUME_NAME).
				WithMountPath(CONTAINER_MODEL_PATH),
		)

	// We have exactly one container in the template.Spec.Containers.
	spreadEnvToInitContainer(template.Spec.Containers[0].Env, initContainer)

	switch p.protocol {
	case OSS:
		initContainer.WithEnv(
			coreapplyv1.EnvVar().WithName("MODEL_SOURCE_TYPE").WithValue(MODEL_SOURCE_MODEL_OBJ_STORE),
			coreapplyv1.EnvVar().WithName("PROVIDER").WithValue(OSS),
			coreapplyv1.EnvVar().WithName("ENDPOINT").WithValue(p.endpoint),
			coreapplyv1.EnvVar().WithName("BUCKET").WithValue(p.bucket),
			coreapplyv1.EnvVar().WithName("MODEL_PATH").WithValue(p.modelPath),
			coreapplyv1.EnvVar().WithName(OSS_ACCESS_KEY_ID).WithValueFrom(coreapplyv1.EnvVarSource().WithSecretKeyRef(coreapplyv1.SecretKeySelector().WithName(OSS_ACCESS_SECRET_NAME).WithKey(OSS_ACCESS_KEY_ID).WithOptional(true))),
			coreapplyv1.EnvVar().WithName(OSS_ACCESS_KEY_SECRET).WithValueFrom(coreapplyv1.EnvVarSource().WithSecretKeyRef(coreapplyv1.SecretKeySelector().WithName(OSS_ACCESS_SECRET_NAME).WithKey(OSS_ACCESS_KEY_SECRET).WithOptional(true))),
		)
	}

	template.Spec.WithInitContainers(initContainer)
}

func (p *URIProvider) InjectModelEnvVars(template *coreapplyv1.PodTemplateSpecApplyConfiguration) {
	switch p.protocol {
	case S3, GCS:
		for i := range template.Spec.Containers {
			if *template.Spec.Containers[i].Name == MODEL_RUNNER_CONTAINER_NAME {
				// Check if AWS credentials already exist
				awsKeyIDExists := false
				awsKeySecretExists := false
				for _, env := range template.Spec.Containers[i].Env {
					if *env.Name == AWS_ACCESS_KEY_ID {
						awsKeyIDExists = true
					}
					if *env.Name == AWS_ACCESS_KEY_SECRET {
						awsKeySecretExists = true
					}
				}

				// Add AWS_ACCESS_KEY_ID if it doesn't exist
				if !awsKeyIDExists {
					template.Spec.Containers[i].WithEnv(coreapplyv1.EnvVar().WithName(AWS_ACCESS_KEY_ID).WithValueFrom(coreapplyv1.EnvVarSource().WithSecretKeyRef(coreapplyv1.SecretKeySelector().WithName(AWS_ACCESS_SECRET_NAME).WithKey(AWS_ACCESS_KEY_ID).WithOptional(true))))
				}

				// Add AWS_ACCESS_KEY_SECRET if it doesn't exist
				if !awsKeySecretExists {
					template.Spec.Containers[i].WithEnv(coreapplyv1.EnvVar().WithName(AWS_ACCESS_KEY_SECRET).WithValueFrom(coreapplyv1.EnvVarSource().WithSecretKeyRef(coreapplyv1.SecretKeySelector().WithName(AWS_ACCESS_SECRET_NAME).WithKey(AWS_ACCESS_KEY_SECRET).WithOptional(true))))
				}
			}
		}
	case OSS:
		for i := range template.Spec.Containers {
			if *template.Spec.Containers[i].Name == MODEL_RUNNER_CONTAINER_NAME {
				// Check if OSS credentials already exist
				ossKeyIDExists := false
				ossKeySecretExists := false
				for _, env := range template.Spec.Containers[i].Env {
					if *env.Name == OSS_ACCESS_KEY_ID {
						ossKeyIDExists = true
					}
					if *env.Name == OSS_ACCESS_KEY_SECRET {
						ossKeySecretExists = true
					}
				}

				// Add OSS_ACCESS_KEY_ID if it doesn't exist
				if !ossKeyIDExists {
					template.Spec.Containers[i].WithEnv(coreapplyv1.EnvVar().WithName(OSS_ACCESS_KEY_ID).WithValueFrom(coreapplyv1.EnvVarSource().WithSecretKeyRef(coreapplyv1.SecretKeySelector().WithName(OSS_ACCESS_SECRET_NAME).WithKey(OSS_ACCESS_KEY_ID).WithOptional(true))))
				}

				// Add OSS_ACCESS_KEY_SECRET if it doesn't exist
				if !ossKeySecretExists {
					template.Spec.Containers[i].WithEnv(coreapplyv1.EnvVar().WithName(OSS_ACCESS_KEY_SECRET).WithValueFrom(coreapplyv1.EnvVarSource().WithSecretKeyRef(coreapplyv1.SecretKeySelector().WithName(OSS_ACCESS_SECRET_NAME).WithKey(OSS_ACCESS_KEY_SECRET).WithOptional(true))))
				}
			}
		}
	}
}
