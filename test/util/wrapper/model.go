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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
)

type ModelWrapper struct {
	coreapi.OpenModel
}

func MakeModel(name string) *ModelWrapper {
	return &ModelWrapper{
		coreapi.OpenModel{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
		},
	}
}

func (w *ModelWrapper) Obj() *coreapi.OpenModel {
	return &w.OpenModel
}

func (w *ModelWrapper) FamilyName(name string) *ModelWrapper {
	w.Spec.FamilyName = coreapi.ModelName(name)
	return w
}

func (w *ModelWrapper) ModelSourceWithModelID(modelID string, filename string, revision string, allowPatterns, ignorePatterns []string) *ModelWrapper {
	if modelID != "" {
		if w.Spec.Source.ModelHub == nil {
			w.Spec.Source.ModelHub = &coreapi.ModelHub{}
		}
		w.Spec.Source.ModelHub.ModelID = modelID

		if filename != "" {
			w.Spec.Source.ModelHub.Filename = &filename
		}

		if revision != "" {
			w.Spec.Source.ModelHub.Revision = &revision
		}

		if allowPatterns != nil {
			w.Spec.Source.ModelHub.AllowPatterns = allowPatterns
		}

		if ignorePatterns != nil {
			w.Spec.Source.ModelHub.IgnorePatterns = ignorePatterns
		}
	}
	return w
}

func (w *ModelWrapper) ModelSourceWithModelHub(modelHub string) *ModelWrapper {
	if modelHub != "" {
		if w.Spec.Source.ModelHub == nil {
			w.Spec.Source.ModelHub = &coreapi.ModelHub{}
		}
		w.Spec.Source.ModelHub.Name = &modelHub
	}
	return w
}

func (w *ModelWrapper) ModelSourceWithURI(uri string) *ModelWrapper {
	value := coreapi.URIProtocol(uri)
	if uri != "" {
		w.Spec.Source.URI = &value
	}
	return w
}

func (w *ModelWrapper) InferenceFlavors(flavors ...coreapi.Flavor) *ModelWrapper {
	if w.Spec.InferenceConfig == nil {
		w.Spec.InferenceConfig = &coreapi.InferenceConfig{}
	}
	w.Spec.InferenceConfig.Flavors = flavors
	return w
}

func (w *ModelWrapper) Label(k, v string) *ModelWrapper {
	if w.Labels == nil {
		w.Labels = map[string]string{}
	}
	w.Labels[k] = v
	return w
}

func (w *ModelWrapper) OwnedBy(ownedBy string) *ModelWrapper {
	w.Spec.OwnedBy = &ownedBy
	return w
}

func (w *ModelWrapper) CreatedAt(createdAt metav1.Time) *ModelWrapper {
	w.Spec.CreatedAt = &createdAt
	return w
}

func MakeFlavor(name string) *FlavorWrapper {
	return &FlavorWrapper{
		coreapi.Flavor{
			Name: coreapi.FlavorName(name),
		},
	}
}

type FlavorWrapper struct {
	coreapi.Flavor
}

func (w *FlavorWrapper) Obj() *coreapi.Flavor {
	return &w.Flavor
}

func (w *FlavorWrapper) SetRequest(r, v string) *FlavorWrapper {
	if w.Limits == nil {
		w.Limits = map[v1.ResourceName]resource.Quantity{}
	}
	w.Limits[v1.ResourceName(r)] = resource.MustParse(v)
	return w
}

func (w *FlavorWrapper) SetNodeSelector(k, v string) *FlavorWrapper {
	if w.NodeSelector == nil {
		w.NodeSelector = map[string]string{}
	}
	w.NodeSelector[k] = v
	return w
}

func (w *FlavorWrapper) SetParams(k, v string) *FlavorWrapper {
	if w.Params == nil {
		w.Params = map[string]string{}
	}
	w.Params[k] = v
	return w
}
