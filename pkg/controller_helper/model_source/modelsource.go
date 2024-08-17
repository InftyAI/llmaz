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
	corev1 "k8s.io/api/core/v1"

	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/pkg/util"
)

const (
	// model path
	CONTAINER_MODEL_PATH = "/workspace/models/"
	HOST_MODEL_PATH      = "/cache/models/"

	// container & volume configs
	DEFAULT_BACKEND_PORT        = 8080
	MODEL_VOLUME_NAME           = "model-volume"
	MODEL_RUNNER_CONTAINER_NAME = "model-runner"
	MODEL_LOADER_CONTAINER_NAME = "model-loader"

	// model source type
	MODEL_SOURCE_MODELHUB        = "modelhub"
	MODEL_SOURCE_MODEL_OBJ_STORE = "objstore"

	// secrets
	MODELHUB_SECRET_NAME  = "modelhub-secret"
	HUGGINGFACE_TOKEN_KEY = "HF_TOKEN"

	OSS_ACCESS_SECRET_NAME = "oss-access-secret"
	OSS_ACCESS_KEY_ID      = "OSS_ACCESS_KEY_ID"
	OSS_ACCESS_KEY_SECRET  = "OSS_ACCESS_KEY_SECRET"
)

type ModelSourceProvider interface {
	ModelName() string
	ModelPath() string
	InjectModelLoader(*corev1.PodTemplateSpec)
}

func NewModelSourceProvider(model *coreapi.OpenModel) ModelSourceProvider {
	if model.Spec.Source.ModelHub != nil {
		return &ModelHubProvider{
			modelName:     model.Name,
			modelID:       model.Spec.Source.ModelHub.ModelID,
			modelHub:      *model.Spec.Source.ModelHub.Name,
			fileName:      model.Spec.Source.ModelHub.Filename,
			modelRevision: model.Spec.Source.ModelHub.Revision,
		}
	}

	if model.Spec.Source.URI != nil {
		// We'll validate the format in the webhook, so generally no error should happen here.
		protocol, address, _ := util.ParseURI(string(*model.Spec.Source.URI))
		provider := &URIProvider{modelName: model.Name, protocol: protocol}

		switch protocol {
		case OSS:
			provider.endpoint, provider.bucket, provider.modelPath, _ = util.ParseOSS(address)
		default:
			// This should be validated at webhooks.
			panic("protocol not supported")
		}

		return provider
	}
	// Should not reach here, it will be validated at webhook in prior.
	return nil
}
