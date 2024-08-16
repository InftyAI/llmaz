/*
Copyright 2024 The Kubernetes Authors.
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

package util

import (
	api "inftyai.com/llmaz/api/core/v1alpha1"
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	"inftyai.com/llmaz/test/util/wrapper"
)

const (
	sampleModelName = "llama3-8b"
)

func MockASampleModel() *api.OpenModel {
	return wrapper.MakeModel(sampleModelName).FamilyName("llama3").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("meta-llama/Meta-Llama-3-8B").Obj()
}

func MockASamplePlayground(ns string) *inferenceapi.Playground {
	return wrapper.MakePlayground("playground-llama3-8b", ns).ModelClaim(sampleModelName).Obj()
}

func MockASampleService(ns string) *inferenceapi.Service {
	return wrapper.MakeService("service-llama3-8b", ns).
		ModelsClaim([]string{"llama3-8b"}, []string{}, nil).
		WorkerTemplate().
		Obj()
}
