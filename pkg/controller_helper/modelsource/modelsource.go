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
	coreapplyv1 "k8s.io/client-go/applyconfigurations/core/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	"github.com/inftyai/llmaz/pkg/util"
)

const (
	// model path
	CONTAINER_MODEL_PATH = "/workspace/models/"
	HOST_MODEL_BASE_PATH = "/mnt/models/"
	// TODO: we may need /mnt/models/namespace1/ path in the future for isolates.
	HOST_CLUSTER_MODEL_PATH = HOST_MODEL_BASE_PATH + "cluster/"

	// container & volume configs
	DEFAULT_BACKEND_PORT        = 8080
	MODEL_VOLUME_NAME           = "model-volume"
	MODEL_RUNNER_CONTAINER_NAME = "model-runner"
	MODEL_LOADER_CONTAINER_NAME = "model-loader"

	// model source type
	MODEL_SOURCE_MODELHUB        = "modelhub"
	MODEL_SOURCE_MODEL_OBJ_STORE = "objstore"

	// secrets
	MODELHUB_SECRET_NAME   = "modelhub-secret"
	HUGGING_FACE_TOKEN_KEY = "HF_TOKEN"
	HUGGING_FACE_HUB_TOKEN = "HUGGING_FACE_HUB_TOKEN"

	OSS_ACCESS_SECRET_NAME = "oss-access-secret"
	OSS_ACCESS_KEY_ID      = "OSS_ACCESS_KEY_ID"
	OSS_ACCESS_KEY_SECRET  = "OSS_ACCESS_KEY_SECRET"

	AWS_ACCESS_SECRET_NAME = "aws-access-secret"
	AWS_ACCESS_KEY_ID      = "AWS_ACCESS_KEY_ID"
	AWS_ACCESS_KEY_SECRET  = "AWS_SECRET_ACCESS_KEY"
)

type ModelSourceProvider interface {
	ModelName() string
	ModelPath(skipModelLoader bool) string
	// InjectModelLoader will inject the model loader to the spec,
	// index refers to the suffix of the initContainer name, like model-loader, model-loader-1.
	InjectModelLoader(spec *coreapplyv1.PodTemplateSpecApplyConfiguration, index int)
	// InjectModelEnvVars will inject the model credentials env to the model-runner container.
	// This is used when the model-loader initContainer is not injected, and the model loading is handled by the model-runner container.
	InjectModelEnvVars(spec *coreapplyv1.PodTemplateSpecApplyConfiguration)
}

func NewModelSourceProvider(model *coreapi.OpenModel) ModelSourceProvider {
	if model.Spec.Source.ModelHub != nil {
		return &ModelHubProvider{
			modelName:           model.Name,
			modelID:             model.Spec.Source.ModelHub.ModelID,
			modelHub:            *model.Spec.Source.ModelHub.Name,
			fileName:            model.Spec.Source.ModelHub.Filename,
			modelRevision:       model.Spec.Source.ModelHub.Revision,
			modelAllowPatterns:  model.Spec.Source.ModelHub.AllowPatterns,
			modelIgnorePatterns: model.Spec.Source.ModelHub.IgnorePatterns,
		}
	}

	if model.Spec.Source.URI != nil {
		// We'll validate the format in the webhook, so generally no error should happen here.
		protocol, value, _ := util.ParseURI(string(*model.Spec.Source.URI))
		provider := &URIProvider{modelName: model.Name, protocol: protocol, uri: string(*model.Spec.Source.URI)}

		switch protocol {
		case OSS:
			provider.endpoint, provider.bucket, provider.modelPath, _ = util.ParseOSS(value)
		case S3, GCS:
			provider.bucket, provider.modelPath, _ = util.ParseS3(value)
		case HostPath:
			provider.modelPath = value
		case Ollama:
			provider.modelPath = value
		default:
			// This should be validated at webhooks.
			panic("protocol not supported")
		}

		return provider
	}
	// Should not reach here, it will be validated at webhook in prior.
	return nil
}

// InjectModelVolume mounts the model-volume to the pod template
// The logic for mounting model-volume to model-runner container is identical in both ModelHubProvider and URIProvider,
// so this function can be reused and only needs to be configured once
func InjectModelVolume(template *coreapplyv1.PodTemplateSpecApplyConfiguration) {
	// Handle container.
	for i, container := range template.Spec.Containers {
		// We only consider this container.
		if *container.Name == MODEL_RUNNER_CONTAINER_NAME {
			template.Spec.Containers[i].WithVolumeMounts(coreapplyv1.VolumeMount().WithName(MODEL_VOLUME_NAME).WithMountPath(CONTAINER_MODEL_PATH).WithReadOnly(true))
		}
	}

	// Handle spec.
	template.Spec.WithVolumes(coreapplyv1.Volume().WithName(MODEL_VOLUME_NAME).WithEmptyDir(coreapplyv1.EmptyDirVolumeSource()))

	// TODO: support OCI image volume
	// template.Spec.WithVolumes(coreapplyv1.Volume().WithName(MODEL_VOLUME_NAME).WithImage(coreapplyv1.ImageVolumeSource().WithReference(url).WithPullPolicy(corev1.PullIfNotPresent)))
}
