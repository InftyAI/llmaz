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

package helper

import (
	"context"
	"strconv"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These two modes are preset.
const (
	DefaultArg             string = "default"
	SpeculativeDecodingArg string = "speculative-decoding"
	ModelParallelismArg    string = "model-parallelism"
)

// DetectArgFrom wil auto detect the arg from model roles if not set explicitly.
func DetectArgFrom(playground *inferenceapi.Playground, isMultiNodesInference bool) string {
	if isMultiNodesInference {
		return ModelParallelismArg
	}

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

// FirstAssignedFlavor will return the first assigned flavor of the model, always the 0-index flavor.
func FirstAssignedFlavor(model *coreapi.OpenModel, playground *inferenceapi.Playground) []coreapi.Flavor {
	var flavors []coreapi.FlavorName
	if playground.Spec.ModelClaim != nil {
		flavors = playground.Spec.ModelClaim.InferenceFlavorClaims
	} else {
		flavors = playground.Spec.ModelClaims.InferenceFlavorClaims
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

// MultiHostInference returns two values, the first one is the TP size,
// the second one is whether this is a multi-host inference.
func MultiHostInference(model *coreapi.OpenModel, playground *inferenceapi.Playground) (int32, bool) {
	flavors := FirstAssignedFlavor(model, playground)
	if len(flavors) > 0 && flavors[0].Params["PP"] != "" {
		size, err := strconv.Atoi(flavors[0].Params["PP"])
		if err != nil {
			return 0, false
		}
		return int32(size), true
	}
	return 0, false
}
