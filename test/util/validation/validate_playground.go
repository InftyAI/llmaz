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

package validation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	helper "github.com/inftyai/llmaz/pkg/controller_helper"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/model_source"
	pkgutil "github.com/inftyai/llmaz/pkg/util"
	"github.com/inftyai/llmaz/test/util"
	"github.com/inftyai/llmaz/test/util/format"
)

func validateModelClaim(models []*coreapi.OpenModel, playground *inferenceapi.Playground, service inferenceapi.Service) error {
	// Make sure the first model is the main model, or the test may fail.
	if playground.Spec.ModelClaim != nil {
		if playground.Spec.ModelClaim.ModelName != service.Spec.ModelClaims.Models[0].Name {
			return fmt.Errorf("expected modelName %s, got %s", playground.Spec.ModelClaim.ModelName, service.Spec.ModelClaims.Models[0].Name)
		}
		if diff := cmp.Diff(playground.Spec.ModelClaim.InferenceFlavors, service.Spec.ModelClaims.InferenceFlavors); diff != "" {
			return fmt.Errorf("unexpected flavors, want %v, got %v", playground.Spec.ModelClaim.InferenceFlavors, service.Spec.ModelClaims.InferenceFlavors)
		}
	} else if playground.Spec.ModelClaims != nil {
		if diff := cmp.Diff(*playground.Spec.ModelClaims, service.Spec.ModelClaims); diff != "" {
			return fmt.Errorf("expected modelClaims, want %v, got %v", *playground.Spec.ModelClaims, service.Spec.ModelClaims)
		}
		if diff := cmp.Diff(playground.Spec.ModelClaims.InferenceFlavors, service.Spec.ModelClaims.InferenceFlavors); diff != "" {
			return fmt.Errorf("unexpected flavors, want %v, got %v", playground.Spec.ModelClaim.InferenceFlavors, service.Spec.ModelClaims.InferenceFlavors)
		}
	}

	if playground.Labels[coreapi.ModelNameLabelKey] != models[0].Name {
		return fmt.Errorf("unexpected Playground label value, want %v, got %v", models[0].Name, playground.Labels[coreapi.ModelNameLabelKey])
	}

	nodeSize, multiHost := helper.MultiHostInference(models[0], playground)
	if multiHost && nodeSize != *service.Spec.WorkloadTemplate.LeaderWorkerTemplate.Size {
		return fmt.Errorf("expected nodeSize %d, got %d", nodeSize, *service.Spec.WorkloadTemplate.LeaderWorkerTemplate.Size)
	}

	return nil
}

func ValidatePlayground(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
	gomega.Eventually(func() error {
		service := inferenceapi.Service{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &service); err != nil {
			return errors.New("failed to get inferenceService")
		}

		models, err := helper.FetchModelsByPlayground(ctx, k8sClient, playground)
		if err != nil {
			return err
		}

		if err := validateModelClaim(models, playground, service); err != nil {
			return err
		}

		if *playground.Spec.Replicas != *service.Spec.WorkloadTemplate.Replicas {
			return fmt.Errorf("expected replicas: %d, got %d", *playground.Spec.Replicas, *service.Spec.WorkloadTemplate.Replicas)
		}

		backendRuntimeName := inferenceapi.DefaultBackend
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Name != nil {
			backendRuntimeName = *playground.Spec.BackendRuntimeConfig.Name
		}
		backendRuntime := inferenceapi.BackendRuntime{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: string(backendRuntimeName)}, &backendRuntime); err != nil {
			return errors.New("failed to get backendRuntime")
		}

		parser := helper.NewBackendRuntimeParser(&backendRuntime)
		multiHost := service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate != nil

		if service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Name != modelSource.MODEL_RUNNER_CONTAINER_NAME {
			return fmt.Errorf("container name not right, want %s, got %s", modelSource.MODEL_RUNNER_CONTAINER_NAME, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Name)
		}
		if multiHost {
			if service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Name != modelSource.MODEL_RUNNER_CONTAINER_NAME {
				return fmt.Errorf("container name not right, want %s, got %s", modelSource.MODEL_RUNNER_CONTAINER_NAME, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Name)
			}
		}

		// compare the same part of leader and worker template, image, version, env, resources.
		if playground.Spec.BackendRuntimeConfig != nil {

			// compare image & version
			if playground.Spec.BackendRuntimeConfig.Version != nil {
				if parser.Image(*playground.Spec.BackendRuntimeConfig.Version) != service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image {
					return fmt.Errorf("expected container image %s, got %s", parser.Image(*playground.Spec.BackendRuntimeConfig.Version), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image)
				}
				if multiHost {
					if parser.Image(*playground.Spec.BackendRuntimeConfig.Version) != service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Image {
						return fmt.Errorf("expected container image %s, got %s", parser.Image(*playground.Spec.BackendRuntimeConfig.Version), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Image)
					}
				}
			} else {
				if parser.Image(parser.Version()) != service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image {
					return fmt.Errorf("expected container image %s, got %s", parser.Image(parser.Version()), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image)
				}
				if multiHost {
					if parser.Image(parser.Version()) != service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Image {
						return fmt.Errorf("expected container image %s, got %s", parser.Image(parser.Version()), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Image)
					}
				}
			}

			if playground.Spec.BackendRuntimeConfig.Envs != nil {
				if diff := cmp.Diff(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Env, playground.Spec.BackendRuntimeConfig.Envs); diff != "" {
					return fmt.Errorf("unexpected envs")
				}
				if multiHost {
					if diff := cmp.Diff(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Env, playground.Spec.BackendRuntimeConfig.Envs); diff != "" {
						return fmt.Errorf("unexpected envs")
					}
				}
			}

			if playground.Spec.BackendRuntimeConfig.Resources != nil {
				for k, v := range playground.Spec.BackendRuntimeConfig.Resources.Limits {
					if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k].Equal(v) {
						return fmt.Errorf("unexpected limits for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k])
					}
					if multiHost {
						if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Limits[k].Equal(v) {
							return fmt.Errorf("unexpected limits for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Limits[k])
						}
					}
				}
				for k, v := range playground.Spec.BackendRuntimeConfig.Resources.Requests {
					if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k].Equal(v) {
						return fmt.Errorf("unexpected requests for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k])
					}
					if multiHost {
						if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Requests[k].Equal(v) {
							return fmt.Errorf("unexpected requests for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Requests[k])
						}
					}
				}
			} else {
				// Validate default resources requirements.
				for k, v := range parser.Resources().Limits {
					if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k].Equal(v) {
						return fmt.Errorf("unexpected limit for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k])
					}
					if multiHost {
						if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Limits[k].Equal(v) {
							return fmt.Errorf("unexpected limit for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Limits[k])
						}
					}
				}
				for k, v := range parser.Resources().Requests {
					if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k].Equal(v) {
						return fmt.Errorf("unexpected limit for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k])
					}
					if multiHost {
						if !service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Requests[k].Equal(v) {
							return fmt.Errorf("unexpected limit for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Resources.Requests[k])
						}
					}
				}
			}

			// compare probes
			if backendRuntime.Spec.StartupProbe != nil {
				if multiHost {
					if diff := cmp.Diff(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].StartupProbe, *backendRuntime.Spec.StartupProbe); diff != "" {
						return fmt.Errorf("unexpected startupProbe")
					}
				} else {
					if diff := cmp.Diff(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].StartupProbe, *backendRuntime.Spec.StartupProbe); diff != "" {
						return fmt.Errorf("unexpected startupProbe")
					}
				}
			}
			if backendRuntime.Spec.LivenessProbe != nil {
				if multiHost {
					if diff := cmp.Diff(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].LivenessProbe, *backendRuntime.Spec.LivenessProbe); diff != "" {
						return fmt.Errorf("unexpected livenessProbe")
					}
				} else {
					if diff := cmp.Diff(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].LivenessProbe, *backendRuntime.Spec.LivenessProbe); diff != "" {
						return fmt.Errorf("unexpected livenessProbe")
					}
				}
			}
			if backendRuntime.Spec.ReadinessProbe != nil {
				if multiHost {
					if diff := cmp.Diff(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].ReadinessProbe, *backendRuntime.Spec.ReadinessProbe); diff != "" {
						return fmt.Errorf("unexpected readinessProbe")
					}
				} else {
					if diff := cmp.Diff(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].ReadinessProbe, *backendRuntime.Spec.ReadinessProbe); diff != "" {
						return fmt.Errorf("unexpected readinessProbe")
					}
				}
			}
		}

		// compare the different parts.

		args, err := parser.Args(playground, models, multiHost)
		if err != nil {
			return err
		}
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Args != nil {
			args = append(args, playground.Spec.BackendRuntimeConfig.Args.Flags...)
		}

		for _, arg := range args {
			if multiHost {
				if len(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Args) != 0 {
					return fmt.Errorf("args should be empty, but got: %v", service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Args)
				}
			} else {
				if !slices.Contains(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Args, arg) {
					return fmt.Errorf("didn't contain arg: %s", arg)
				}
			}
		}

		if multiHost {
			if diff := cmp.Diff(pkgutil.MergeArgsWithCommands(parser.LeaderCommands(), args), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.LeaderTemplate.Spec.Containers[0].Command); diff != "" {
				return errors.New("command not right")
			}
			if diff := cmp.Diff(parser.WorkerCommands(), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Command); diff != "" {
				return errors.New("command not right")
			}
		} else {
			if diff := cmp.Diff(parser.Commands(), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Command); diff != "" {
				return errors.New("command not right")
			}
		}
		return nil

	}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
}

func ValidatePlaygroundStatusEqualTo(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground, conditionType string, reason string, status metav1.ConditionStatus) {
	testType := os.Getenv("TEST_TYPE")
	timeout := util.IntegrationTimeout
	interval := util.Interval

	if testType == "E2E" {
		timeout = util.E2ETimeout
		interval = util.E2EInterval
	}

	gomega.Eventually(func() error {
		newPlayground := inferenceapi.Playground{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &newPlayground); err != nil {
			return err
		}
		if condition := apimeta.FindStatusCondition(newPlayground.Status.Conditions, conditionType); condition == nil {
			return fmt.Errorf("condition not found: %s", format.Object(newPlayground, 1))
		} else {
			if condition.Reason != reason || condition.Status != status {
				return fmt.Errorf("expected reason %q or status %q, but got %s", reason, status, format.Object(condition, 1))
			}
		}
		return nil
	}, timeout, interval).Should(gomega.Succeed())
}
