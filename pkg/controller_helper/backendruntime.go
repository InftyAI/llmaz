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
	"fmt"
	"regexp"
	"strings"

	corev1 "k8s.io/api/core/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/model_source"
)

// TODO: add unit tests.
type BackendRuntimeParser struct {
	backendRuntime *inferenceapi.BackendRuntime
}

func NewBackendRuntimeParser(backendRuntime *inferenceapi.BackendRuntime) *BackendRuntimeParser {
	return &BackendRuntimeParser{backendRuntime}
}

func (p *BackendRuntimeParser) Commands() []string {
	return p.backendRuntime.Spec.Commands
}

func (p *BackendRuntimeParser) Envs() []corev1.EnvVar {
	return p.backendRuntime.Spec.Envs
}

func (p *BackendRuntimeParser) Args(playground *inferenceapi.Playground, models []*coreapi.OpenModel) ([]string, error) {
	var argName string
	if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.ArgName != nil {
		argName = *playground.Spec.BackendRuntimeConfig.ArgName
	} else {
		// Auto detect the args from model roles.
		argName = DetectArgFrom(playground)
	}

	source := modelSource.NewModelSourceProvider(models[0])
	modelInfo := map[string]string{
		"ModelPath": source.ModelPath(),
		"ModelName": source.ModelName(),
	}

	// TODO: This is not that reliable because two models doesn't always means speculative-decoding.
	// Revisit this later.
	if len(models) > 1 {
		modelInfo["DraftModelPath"] = modelSource.NewModelSourceProvider(models[1]).ModelPath()
	}

	for _, arg := range p.backendRuntime.Spec.Args {
		if arg.Name == argName {
			return renderFlags(arg.Flags, modelInfo)
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

func (p *BackendRuntimeParser) Resources() inferenceapi.ResourceRequirements {
	return p.backendRuntime.Spec.Resources
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
			if !exists {
				return nil, fmt.Errorf("missing flag or the flag has format error: %s", flag)
			}
			value = strings.Replace(value, match[0], replacement, -1)
		}

		res = append(res, value)
	}

	return res, nil
}
