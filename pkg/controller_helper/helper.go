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

package helper

import (
	"context"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These two modes are preset.
const (
	DefaultArg             string = "default"
	SpeculativeDecodingArg string = "speculative-decoding"
)

func RecommendedConfigName(playground *inferenceapi.Playground) string {
	var name string
	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.ConfigName != nil {
		name = *playground.Spec.BackendRuntimeConfig.ConfigName
	} else {
		// Auto detect the args from model roles.
		name = DetectArgFrom(playground)
	}

	return name
}

// DetectArgFrom wil auto detect the arg from model roles if not set explicitly.
func DetectArgFrom(playground *inferenceapi.Playground) string {
	if playground.Spec.ModelClaim != nil {
		return DefaultArg
	}

	if playground.Spec.ModelClaims != nil {
		for _, mr := range playground.Spec.ModelClaims.Models {
			if *mr.Role == coreapi.DraftRole {
				return SpeculativeDecodingArg
			}
		}
	}

	// We should not reach here.
	return DefaultArg
}

func FetchModelsByService(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) (models []*coreapi.OpenModel, err error) {
	return fetchModels(ctx, k8sClient, service.Spec.ModelClaims.Models)
}

func FetchModelsByPlayground(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) (models []*coreapi.OpenModel, err error) {
	mainRole := coreapi.MainRole
	mrs := []coreapi.ModelRef{}

	if playground.Spec.ModelClaim != nil {
		mrs = append(mrs, coreapi.ModelRef{Name: playground.Spec.ModelClaim.ModelName, Role: &mainRole})
	} else {
		mrs = playground.Spec.ModelClaims.Models
	}

	return fetchModels(ctx, k8sClient, mrs)
}

func fetchModels(ctx context.Context, k8sClient client.Client, mrs []coreapi.ModelRef) (models []*coreapi.OpenModel, err error) {
	for _, mr := range mrs {
		model := &coreapi.OpenModel{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: string(mr.Name)}, model); err != nil {
			return nil, err
		}
		// Make sure the main model is always the 0-index model.
		// We only have one main model right now, if this changes,
		// the logic may also change here.
		if *mr.Role == coreapi.MainRole {
			models = append([]*coreapi.OpenModel{model}, models...)
		} else {
			models = append(models, model)
		}
	}

	return models, nil
}

// FirstAssignedFlavor will return the first assigned flavor of the model.
func FirstAssignedFlavor(model *coreapi.OpenModel, playground *inferenceapi.Playground) []coreapi.Flavor {
	var flavors []coreapi.FlavorName
	if playground.Spec.ModelClaim != nil {
		flavors = playground.Spec.ModelClaim.InferenceFlavors
	} else {
		flavors = playground.Spec.ModelClaims.InferenceFlavors
	}

	if len(flavors) == 0 && (model.Spec.InferenceConfig == nil || len(model.Spec.InferenceConfig.Flavors) == 0) {
		return nil
	}

	if len(flavors) == 0 {
		return []coreapi.Flavor{model.Spec.InferenceConfig.Flavors[0]}
	}

	for _, flavor := range model.Spec.InferenceConfig.Flavors {
		if flavor.Name == flavors[0] {
			return []coreapi.Flavor{flavor}
		}
	}

	return nil
}

func SkipModelLoader(obj metav1.Object) bool {
	if annotations := obj.GetAnnotations(); annotations != nil {
		return annotations[inferenceapi.SkipModelLoaderAnnoKey] == "true"
	}
	return false
}
