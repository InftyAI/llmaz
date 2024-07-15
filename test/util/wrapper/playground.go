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

package wrapper

import (
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	api "inftyai.com/llmaz/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type PlaygroundWrapper struct {
	inferenceapi.Playground
}

func MakePlayground(name string, ns string) *PlaygroundWrapper {
	return &PlaygroundWrapper{
		inferenceapi.Playground{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
		},
	}
}

func (w *PlaygroundWrapper) Obj() *inferenceapi.Playground {
	return &w.Playground
}

func (w *PlaygroundWrapper) Replicas(replicas int32) *PlaygroundWrapper {
	w.Spec.Replicas = &replicas
	return w
}

func (w *PlaygroundWrapper) ModelClaim(modelName string, flavorNames ...string) *PlaygroundWrapper {
	var names []api.FlavorName
	for _, name := range flavorNames {
		names = append(names, api.FlavorName(name))
	}

	w.Spec.ModelClaim = &api.ModelClaim{
		ModelName:        api.ModelName(modelName),
		InferenceFlavors: names,
	}
	return w
}
