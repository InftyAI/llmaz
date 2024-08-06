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
	core "inftyai.com/llmaz/api/core/v1alpha1"
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
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
	names := []core.FlavorName{}
	for _, name := range flavorNames {
		names = append(names, core.FlavorName(name))
	}

	w.Spec.ModelClaim = &core.ModelClaim{
		ModelName:        core.ModelName(modelName),
		InferenceFlavors: names,
	}
	return w
}

func (w *PlaygroundWrapper) Backend(name string) *PlaygroundWrapper {
	if w.Spec.BackendConfig == nil {
		w.Spec.BackendConfig = &inferenceapi.BackendConfig{}
	}
	backendName := inferenceapi.BackendName(name)
	w.Spec.BackendConfig.Name = &backendName
	return w
}

func (w *PlaygroundWrapper) BackendVersion(version string) *PlaygroundWrapper {
	if w.Spec.BackendConfig == nil {
		w = w.Backend("vllm")
	}
	w.Spec.BackendConfig.Version = &version
	return w
}

func (w *PlaygroundWrapper) BackendArgs(args []string) *PlaygroundWrapper {
	if w.Spec.BackendConfig == nil {
		w = w.Backend("vllm")
	}
	w.Spec.BackendConfig.Args = args
	return w
}

func (w *PlaygroundWrapper) BackendEnv(k, v string) *PlaygroundWrapper {
	if w.Spec.BackendConfig == nil {
		w = w.Backend("vllm")
	}
	w.Spec.BackendConfig.Envs = append(w.Spec.BackendConfig.Envs, v1.EnvVar{
		Name:  k,
		Value: v,
	})
	return w
}

func (w *PlaygroundWrapper) BackendRequest(r, v string) *PlaygroundWrapper {
	if w.Spec.BackendConfig == nil {
		w = w.Backend("vllm")
	}
	w.Spec.BackendConfig.Resources = &inferenceapi.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceName(r): resource.MustParse(v),
		},
	}
	return w
}

func (w *PlaygroundWrapper) BackendLimit(r, v string) *PlaygroundWrapper {
	if w.Spec.BackendConfig == nil {
		w = w.Backend("vllm")
	}
	w.Spec.BackendConfig.Resources = &inferenceapi.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceName(r): resource.MustParse(v),
		},
	}
	return w
}
