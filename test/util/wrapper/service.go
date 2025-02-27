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

package wrapper

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
)

type ServiceWrapper struct {
	inferenceapi.Service
}

func MakeService(name string, ns string) *ServiceWrapper {
	return &ServiceWrapper{
		inferenceapi.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: ns,
			},
		},
	}
}

func (w *ServiceWrapper) Obj() *inferenceapi.Service {
	return &w.Service
}

func (w *ServiceWrapper) ModelClaims(modelNames []string, roles []string, flavorNames ...string) *ServiceWrapper {
	models := []coreapi.ModelRef{}
	for i, name := range modelNames {
		models = append(models, coreapi.ModelRef{Name: coreapi.ModelName(name), Role: (*coreapi.ModelRole)(&roles[i])})
	}
	w.Spec.ModelClaims = coreapi.ModelClaims{
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

func (w *ServiceWrapper) WorkerTemplate() *ServiceWrapper {
	w.Spec.WorkloadTemplate.WorkerTemplate = corev1.PodTemplateSpec{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "model-runner",
					Image: "vllm:test",
				},
			},
		},
	}
	return w
}

func (w *ServiceWrapper) ContainerName(name string) *ServiceWrapper {
	w.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Name = name
	return w
}

func (w *ServiceWrapper) InitContainerName(name string) *ServiceWrapper {
	w.Spec.WorkloadTemplate.WorkerTemplate.Spec.InitContainers[0].Name = name
	return w
}

func (w *ServiceWrapper) RolloutStrategy(typ string, maxUnavailable int, maxSurge int) *ServiceWrapper {
	if w.Spec.RolloutStrategy == nil {
		w.Spec.RolloutStrategy = &lws.RolloutStrategy{}
	}
	w.Spec.RolloutStrategy.Type = lws.RolloutStrategyType(typ)
	w.Spec.RolloutStrategy.RollingUpdateConfiguration = &lws.RollingUpdateConfiguration{
		MaxUnavailable: intstr.FromInt(maxUnavailable),
		MaxSurge:       intstr.FromInt(maxSurge),
	}
	return w
}

func (w *ServiceWrapper) Size(size int32) *ServiceWrapper {
	w.Spec.WorkloadTemplate.Size = ptr.To[int32](size)
	return w
}

func (w *ServiceWrapper) Replicas(replicas int32) *ServiceWrapper {
	w.Spec.Replicas = ptr.To[int32](replicas)
	return w
}

func (w *ServiceWrapper) RestartPolicy(policy string) *ServiceWrapper {
	w.Spec.WorkloadTemplate.RestartPolicy = lws.RestartPolicyType(policy)
	return w
}
