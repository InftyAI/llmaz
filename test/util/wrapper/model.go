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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	core "inftyai.com/llmaz/api/core/v1alpha1"
)

type ModelWrapper struct {
	core.Model
}

func MakeModel(name string) *ModelWrapper {
	return &ModelWrapper{
		core.Model{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}
}

func (w *ModelWrapper) Obj() *core.Model {
	return &w.Model
}

func (w *ModelWrapper) FamilyName(name string) *ModelWrapper {
	w.Spec.FamilyName = core.ModelName(name)
	return w
}

func (w *ModelWrapper) DataSourceWithModelID(modelID string) *ModelWrapper {
	if modelID != "" {
		if w.Spec.DataSource.ModelHub == nil {
			w.Spec.DataSource.ModelHub = &core.ModelHub{}
		}
		w.Spec.DataSource.ModelHub.ModelID = modelID
	}
	return w
}

func (w *ModelWrapper) DataSourceWithModelHub(modelHub string) *ModelWrapper {
	if modelHub != "" {
		if w.Spec.DataSource.ModelHub == nil {
			w.Spec.DataSource.ModelHub = &core.ModelHub{}
		}
		w.Spec.DataSource.ModelHub.Name = &modelHub
	}
	return w
}

func (w *ModelWrapper) DataSourceWithURI(uri string) *ModelWrapper {
	if uri != "" {
		w.Spec.DataSource.URI = &uri
	}
	return w
}

func (w *ModelWrapper) InferenceFlavors() *ModelWrapper {
	return w
}

func (w *ModelWrapper) Label(k, v string) *ModelWrapper {
	if w.Labels == nil {
		w.Labels = map[string]string{}
	}
	w.Labels[k] = v
	return w
}

type FlavorWrapper struct {
	core.Flavor
}

func (w *FlavorWrapper) Obj() *core.Flavor {
	return &w.Flavor
}

func (w *FlavorWrapper) SetName(name string) *core.Flavor {
	w.Name = core.FlavorName(name)
	return &w.Flavor
}

func (w *FlavorWrapper) SetRequest(r, v string) *core.Flavor {
	w.Requests[v1.ResourceName(r)] = resource.MustParse(v)
	return &w.Flavor
}

func (w *FlavorWrapper) SetNodeSelector(k, v string) *core.Flavor {
	if w.NodeSelector == nil {
		w.NodeSelector = map[string]string{}
	}
	w.NodeSelector[k] = v
	return &w.Flavor
}

func (w *FlavorWrapper) SetParams(k, v string) *core.Flavor {
	w.Params[k] = v
	return &w.Flavor
}
