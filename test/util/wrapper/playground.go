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
	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
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

func (w *PlaygroundWrapper) Label(k, v string) *PlaygroundWrapper {
	if w.Labels == nil {
		w.Labels = map[string]string{}
	}
	w.Labels[k] = v
	return w
}

func (w *PlaygroundWrapper) Replicas(replicas int32) *PlaygroundWrapper {
	w.Spec.Replicas = &replicas
	return w
}

func (w *PlaygroundWrapper) ModelClaim(modelName string, flavorNames ...string) *PlaygroundWrapper {
	names := []coreapi.FlavorName{}
	for _, name := range flavorNames {
		names = append(names, coreapi.FlavorName(name))
	}
	w.Spec.ModelClaim = &coreapi.ModelClaim{
		ModelName: coreapi.ModelName(modelName),
	}

	if len(names) > 0 {
		w.Spec.ModelClaim.InferenceFlavors = names
	}
	return w
}

func (w *PlaygroundWrapper) ModelClaims(modelNames []string, roles []string, flavorNames ...string) *PlaygroundWrapper {
	models := []coreapi.ModelRef{}
	for i, name := range modelNames {
		models = append(models, coreapi.ModelRef{Name: coreapi.ModelName(name), Role: (*coreapi.ModelRole)(&roles[i])})
	}
	w.Spec.ModelClaims = &coreapi.ModelClaims{
		Models: models,
	}

	fNames := []coreapi.FlavorName{}
	for _, name := range flavorNames {
		fNames = append(fNames, coreapi.FlavorName(name))
	}

	if len(fNames) > 0 {
		w.Spec.ModelClaims.InferenceFlavors = fNames
	}
	return w
}

func (w *PlaygroundWrapper) BackendRuntime(name string) *PlaygroundWrapper {
	if w.Spec.BackendRuntimeConfig == nil {
		w.Spec.BackendRuntimeConfig = &inferenceapi.BackendRuntimeConfig{}
	}
	backendName := inferenceapi.BackendName(name)
	w.Spec.BackendRuntimeConfig.Name = &backendName
	return w
}

func (w *PlaygroundWrapper) BackendRuntimeVersion(version string) *PlaygroundWrapper {
	if w.Spec.BackendRuntimeConfig == nil {
		w = w.BackendRuntime("vllm")
	}
	w.Spec.BackendRuntimeConfig.Version = &version
	return w
}

func (w *PlaygroundWrapper) BackendRuntimeArgs(name string, args []string) *PlaygroundWrapper {
	if w.Spec.BackendRuntimeConfig == nil {
		w = w.BackendRuntime("vllm")
	}
	if w.Spec.BackendRuntimeConfig.Args == nil {
		w.Spec.BackendRuntimeConfig.Args = &inferenceapi.BackendRuntimeArg{}
	}
	w.Spec.BackendRuntimeConfig.Args.Name = &name
	w.Spec.BackendRuntimeConfig.Args.Flags = args
	return w
}

func (w *PlaygroundWrapper) BackendRuntimeEnv(k, v string) *PlaygroundWrapper {
	if w.Spec.BackendRuntimeConfig == nil {
		w = w.BackendRuntime("vllm")
	}
	w.Spec.BackendRuntimeConfig.Envs = append(w.Spec.BackendRuntimeConfig.Envs, v1.EnvVar{
		Name:  k,
		Value: v,
	})
	return w
}

func (w *PlaygroundWrapper) BackendRuntimeRequest(r, v string) *PlaygroundWrapper {
	if w.Spec.BackendRuntimeConfig == nil {
		w = w.BackendRuntime("vllm")
	}
	if w.Spec.BackendRuntimeConfig.Resources == nil {
		w.Spec.BackendRuntimeConfig.Resources = &inferenceapi.ResourceRequirements{}
	}
	if w.Spec.BackendRuntimeConfig.Resources.Requests == nil {
		w.Spec.BackendRuntimeConfig.Resources.Requests = v1.ResourceList{}
	}
	w.Spec.BackendRuntimeConfig.Resources.Requests[v1.ResourceName(r)] = resource.MustParse(v)
	return w
}

func (w *PlaygroundWrapper) BackendRuntimeLimit(r, v string) *PlaygroundWrapper {
	if w.Spec.BackendRuntimeConfig == nil {
		w = w.BackendRuntime("vllm")
	}
	if w.Spec.BackendRuntimeConfig.Resources == nil {
		w.Spec.BackendRuntimeConfig.Resources = &inferenceapi.ResourceRequirements{}
	}
	if w.Spec.BackendRuntimeConfig.Resources.Limits == nil {
		w.Spec.BackendRuntimeConfig.Resources.Limits = v1.ResourceList{}
	}
	w.Spec.BackendRuntimeConfig.Resources.Limits[v1.ResourceName(r)] = resource.MustParse(v)
	return w
}

func (w *PlaygroundWrapper) ElasticConfig(minReplicas, maxReplicas int32) *PlaygroundWrapper {
	w.Spec.ElasticConfig = &inferenceapi.ElasticConfig{
		MaxReplicas: ptr.To[int32](maxReplicas),
		MinReplicas: ptr.To[int32](minReplicas),
	}
	return w
}

func (w *PlaygroundWrapper) HPA(config *inferenceapi.HPATrigger) *PlaygroundWrapper {
	if w.Spec.ElasticConfig == nil {
		w.Spec.ElasticConfig = &inferenceapi.ElasticConfig{}
	}
	if w.Spec.ElasticConfig.ScaleTrigger == nil {
		w.Spec.ElasticConfig.ScaleTrigger = &inferenceapi.ScaleTrigger{}
	}
	w.Spec.ElasticConfig.ScaleTrigger.HPA = config
	return w
}

func (w *PlaygroundWrapper) ScaleTriggerRef(name string) *PlaygroundWrapper {
	if w.Spec.ElasticConfig == nil {
		w.Spec.ElasticConfig = &inferenceapi.ElasticConfig{}
	}
	if w.Spec.ElasticConfig.ScaleTriggerRef == nil {
		w.Spec.ElasticConfig.ScaleTriggerRef = &inferenceapi.ScaleTriggerRef{}
	}
	w.Spec.ElasticConfig.ScaleTriggerRef.Name = name
	return w
}
