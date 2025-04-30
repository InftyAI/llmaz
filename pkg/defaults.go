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

package pkg

import "os"

var (
	LOADER_IMAGE = os.Getenv("MODEL_LOADER_IMAGE")
)

func init() {
	// If the environment variable is not set,
	// assign the default image tag.
	if LOADER_IMAGE == "" {
		LOADER_IMAGE = "inftyai/model-loader:v0.0.10"
	}
}
