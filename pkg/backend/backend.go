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

package backend

import (
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
)

const (
	DEFAULT_MODEL_PATH = "/workspace/models/"
	DEFAULT_PORT       = "8080"
)

// Backend represents the inference engine, such as vllm.
type Backend interface {
	// Name returns the inference backend name in this project.
	Name() inferenceapi.BackendName
	// Image returns the container image for the inference backend.
	Image(version string) string
	// DefaultVersion returns the default version for the inference backend.
	DefaultVersion() string
	// DefaultResources returns the default resources set for the container.
	DefaultResources() inferenceapi.ResourceRequirements
	// DefaultCommands returns the default command to start the inference backend.
	DefaultCommands() []string
}

func SwitchBackend(name inferenceapi.BackendName) Backend {
	switch name {
	case inferenceapi.VLLM:
		return &VLLM{}
	default:
		// We should not reach here because apiserver already did validation.
		return nil
	}
}
