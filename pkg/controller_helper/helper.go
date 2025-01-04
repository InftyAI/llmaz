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

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// These two modes are preset.
const (
	DefaultArg             string = "default"
	SpeculativeDecodingArg string = "speculative-decoding"
)

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
	mrs := []coreapi.ModelRefer{}

	if playground.Spec.ModelClaim != nil {
		mrs = append(mrs, coreapi.ModelRefer{Name: playground.Spec.ModelClaim.ModelName, Role: &mainRole})
	} else {
		mrs = playground.Spec.ModelClaims.Models
	}

	return fetchModels(ctx, k8sClient, mrs)
}

func fetchModels(ctx context.Context, k8sClient client.Client, mrs []coreapi.ModelRefer) (models []*coreapi.OpenModel, err error) {
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
