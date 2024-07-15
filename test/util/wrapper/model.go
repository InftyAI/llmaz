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

	api "inftyai.com/llmaz/api/v1alpha1"
)

type ModelWrapper struct {
	api.Model
}

func MakeModel(name string) *ModelWrapper {
	return &ModelWrapper{
		api.Model{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}
}

func (w *ModelWrapper) Obj() *api.Model {
	return &w.Model
}

func (w *ModelWrapper) FamilyName(name string) *ModelWrapper {
	w.Spec.FamilyName = api.ModelName(name)
	return w
}

func (w *ModelWrapper) DataSourceWithModel(modelID, modelHub string) *ModelWrapper {
	if modelID != "" {
		w.Spec.DataSource.ModelID = &modelID
	}
	if modelHub != "" {
		w.Spec.DataSource.ModelHub = &modelHub
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
	api.Flavor
}

func (w *FlavorWrapper) Obj() *api.Flavor {
	return &w.Flavor
}

func (w *FlavorWrapper) SetName(name string) *api.Flavor {
	w.Name = api.FlavorName(name)
	return &w.Flavor
}

func (w *FlavorWrapper) SetRequest(r, v string) *api.Flavor {
	w.Requests[v1.ResourceName(r)] = resource.MustParse(v)
	return &w.Flavor
}

func (w *FlavorWrapper) SetNodeSelector(selector v1.NodeSelector) *api.Flavor {
	w.NodeSelector = append(w.NodeSelector, selector)
	return &w.Flavor
}

func (w *FlavorWrapper) SetParams(k, v string) *api.Flavor {
	w.Params[k] = v
	return &w.Flavor
}
