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
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	helper "github.com/inftyai/llmaz/pkg/controller_helper"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/modelsource"
)

// TODO: add unit tests.
type BackendRuntimeParser struct {
	backendRuntime      *inferenceapi.BackendRuntime
	models              []*coreapi.OpenModel
	playground          *inferenceapi.Playground
	recommendConfigName string
}

func NewBackendRuntimeParser(backendRuntime *inferenceapi.BackendRuntime, models []*coreapi.OpenModel, playground *inferenceapi.Playground) *BackendRuntimeParser {
	name := helper.RecommendedConfigName(playground)
	return &BackendRuntimeParser{
		backendRuntime,
		models,
		playground,
		name,
	}
}

func (p *BackendRuntimeParser) Command() []string {
	return p.backendRuntime.Spec.Command
}

func (p *BackendRuntimeParser) Envs() []corev1.EnvVar {
	return p.backendRuntime.Spec.Envs
}

func (p *BackendRuntimeParser) Lifecycle() *corev1.Lifecycle {
	return p.backendRuntime.Spec.Lifecycle
}

func (p *BackendRuntimeParser) Args() ([]string, error) {
	mainModel := p.models[0]

	source := modelSource.NewModelSourceProvider(mainModel)
	modelInfo := map[string]string{
		"ModelPath": source.ModelPath(helper.SkipModelLoader(p.playground)),
		"ModelName": source.ModelName(),
	}

	// TODO: This is not that reliable because two models doesn't always means speculative-decoding.
	// Revisit this later.
	if len(p.models) > 1 {
		modelInfo["DraftModelPath"] = modelSource.NewModelSourceProvider(p.models[1]).ModelPath(helper.SkipModelLoader(p.playground))
	}

	for _, recommend := range p.backendRuntime.Spec.RecommendedConfigs {
		if recommend.Name == p.recommendConfigName {
			return renderFlags(recommend.Args, modelInfo)
		}
	}

	// We should not reach here.
	return nil, fmt.Errorf("failed to parse backendRuntime %s", p.backendRuntime.Name)
}

func (p *BackendRuntimeParser) Image(version string) string {
	return p.backendRuntime.Spec.Image + ":" + version
}

func (p *BackendRuntimeParser) Version() string {
	return p.backendRuntime.Spec.Version
}

func (p *BackendRuntimeParser) Resources() *inferenceapi.ResourceRequirements {
	for _, recommend := range p.backendRuntime.Spec.RecommendedConfigs {
		if recommend.Name == p.recommendConfigName {
			return recommend.Resources
		}
	}
	// We should not reach here.
	return nil
}

func (p *BackendRuntimeParser) SharedMemorySize() *resource.Quantity {
	for _, recommend := range p.backendRuntime.Spec.RecommendedConfigs {
		if recommend.Name == p.recommendConfigName {
			return recommend.SharedMemorySize
		}
	}
	return nil
}

func renderFlags(flags []string, modelInfo map[string]string) ([]string, error) {
	// Capture the word.
	re := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)

	res := []string{}

	for _, flag := range flags {
		value := flag
		matches := re.FindAllStringSubmatch(flag, -1)
		for _, match := range matches {
			if len(match) <= 1 {
				continue
			}
			key := match[1]
			replacement, exists := modelInfo[key]
			if !exists || replacement == "" {
				return nil, fmt.Errorf("missing flag or the flag has format error: %s", flag)
			}
			value = strings.Replace(value, match[0], replacement, -1)
		}

		res = append(res, value)
	}

	return res, nil
}
