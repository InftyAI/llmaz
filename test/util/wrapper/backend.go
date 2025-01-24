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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
)

type BackendRuntimeWrapper struct {
	inferenceapi.BackendRuntime
}

func MakeBackendRuntime(name string) *BackendRuntimeWrapper {
	return &BackendRuntimeWrapper{
		inferenceapi.BackendRuntime{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}
}

func (w *BackendRuntimeWrapper) Obj() *inferenceapi.BackendRuntime {
	return &w.BackendRuntime
}

func (w *BackendRuntimeWrapper) Name(name string) *BackendRuntimeWrapper {
	w.ObjectMeta.Name = name
	return w
}

func (w *BackendRuntimeWrapper) Image(image string) *BackendRuntimeWrapper {
	w.Spec.Image = image
	return w
}

func (w *BackendRuntimeWrapper) Version(version string) *BackendRuntimeWrapper {
	w.Spec.Version = version
	return w
}

func (w *BackendRuntimeWrapper) Command(commands []string) *BackendRuntimeWrapper {
	w.Spec.Commands = commands
	return w
}

func (w *BackendRuntimeWrapper) Arg(name string, flags []string) *BackendRuntimeWrapper {
	w.Spec.Args = append(w.Spec.Args, inferenceapi.BackendRuntimeArg{
		Name:  ptr.To[string](name),
		Flags: flags,
	})
	return w
}

func (w *BackendRuntimeWrapper) Request(r, v string) *BackendRuntimeWrapper {
	if w.Spec.Resources.Requests == nil {
		w.Spec.Resources.Requests = corev1.ResourceList{}
	}
	w.Spec.Resources.Requests[corev1.ResourceName(r)] = resource.MustParse(v)
	return w
}

func (w *BackendRuntimeWrapper) Limit(r, v string) *BackendRuntimeWrapper {
	if w.Spec.Resources.Limits == nil {
		w.Spec.Resources.Limits = corev1.ResourceList{}
	}
	w.Spec.Resources.Limits[corev1.ResourceName(r)] = resource.MustParse(v)
	return w
}

func (w *BackendRuntimeWrapper) Probe(name string, probe *corev1.Probe) *BackendRuntimeWrapper {
	if name == "liveness" {
		w.Spec.LivenessProbe = probe
	}
	if name == "readiness" {
		w.Spec.ReadinessProbe = probe
	}
	if name == "startup" {
		w.Spec.LivenessProbe = probe
	}
	return w
}
