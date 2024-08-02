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
	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/pkg"
	corev1 "k8s.io/api/core/v1"
)

var _ DataSourceProvider = &URIProvider{}

type URIProvider struct {
	model *coreapi.Model
}

func (p *URIProvider) ModelName() string {
	return p.model.Name
}

// Example:
// - URI: s3://a/b/c
// - modelPath: /workspace/models/
// c is the folder name of the model.
func (p *URIProvider) ModelPath() string {
	return pkg.CONTAINER_MODEL_PATH
}

func (p *URIProvider) InjectModelLoader(template *corev1.PodTemplateSpec) {
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

	// We'll validate the format in the webhook, go generally no error should happen here.
	// _, url, _ := util.ParseURI(*p.model.Spec.DataSource.URI)

	template.Spec.Volumes = append(template.Spec.Volumes, corev1.Volume{
		Name:         pkg.MODEL_VOLUME_NAME,
		VolumeSource: corev1.VolumeSource{
			// Image: &corev1.ImageVolumeSource{
			// 	Reference:  url,
			// 	PullPolicy: corev1.PullIfNotPresent,
			// },
		},
	})
}
