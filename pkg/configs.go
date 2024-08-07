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

package pkg

const (
	LOADER_IMAGE = "inftyai/model-loader:v0.0.4"

	HOST_MODEL_PATH             = "/cache/models/"
	CONTAINER_MODEL_PATH        = "/workspace/models/"
	DEFAULT_BACKEND_PORT        = 8080
	MODEL_VOLUME_NAME           = "model-volume"
	MODEL_RUNNER_CONTAINER_NAME = "model-runner"
	MODEL_LOADER_CONTAINER_NAME = "model-loader"
	MODEL_SECRET_NAME           = "model-secret"

	HUGGINGFACE_TOKEN_KEY = "HF_TOKEN"

	HUGGINGFACE_HUB = "Huggingface"
	MODELSCOPE_HUB  = "ModelScope"
)
