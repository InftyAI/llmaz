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

package datasource

import (
	"strings"

	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/pkg"
	corev1 "k8s.io/api/core/v1"
)

var _ DataSourceProvider = &ModelHubProvider{}

type ModelHubProvider struct {
	model *coreapi.Model
}

func (p *ModelHubProvider) ModelName() string {
	return p.model.Name
}

// Example:
// - modelID: facebook/opt-125m
// - modelPath: /workspace/models/models--facebook--opt-125m
func (p *ModelHubProvider) ModelPath() string {
	return pkg.CONTAINER_MODEL_PATH + "models--" + strings.ReplaceAll(p.model.Spec.DataSource.ModelHub.ModelID, "/", "--")
}

func (p *ModelHubProvider) InjectModelLoader(template *corev1.PodTemplateSpec) {
	template.Spec.InitContainers = append(template.Spec.InitContainers, corev1.Container{
		Name:  pkg.MODEL_LOADER_CONTAINER_NAME,
		Image: pkg.LOADER_IMAGE,
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      pkg.MODEL_VOLUME_NAME,
				MountPath: pkg.CONTAINER_MODEL_PATH,
			},
		},
	})

	// This is related to the model loader logics which will read the environment when loading models weights.
	template.Spec.InitContainers[0].Env = append(
		template.Spec.InitContainers[0].Env,
		corev1.EnvVar{Name: "MODEL_ID", Value: p.model.Spec.DataSource.ModelHub.ModelID},
		corev1.EnvVar{Name: "MODEL_HUB_NAME", Value: *p.model.Spec.DataSource.ModelHub.Name},
	)

	for _, container := range template.Spec.Containers {
		// We only consider this container.
		if container.Name == pkg.MODEL_RUNNER_CONTAINER_NAME {
			container.VolumeMounts = append(container.VolumeMounts, corev1.VolumeMount{
				Name:      pkg.MODEL_VOLUME_NAME,
				MountPath: pkg.CONTAINER_MODEL_PATH,
				ReadOnly:  true,
			})
		}
	}

	hostType := corev1.HostPathDirectoryOrCreate
	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name: pkg.MODEL_VOLUME_NAME,
		VolumeSource: corev1.VolumeSource{
			HostPath: &corev1.HostPathVolumeSource{
				Path: pkg.HOST_MODEL_PATH,
				Type: &hostType,
			},
		},
	})
}
