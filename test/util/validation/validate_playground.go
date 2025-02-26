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
	backendruntime "github.com/inftyai/llmaz/pkg/controller_helper/backendruntime"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/modelsource"
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

		if *playground.Spec.Replicas != *service.Spec.Replicas {
			return fmt.Errorf("expected replicas: %d, got %d", *playground.Spec.Replicas, *service.Spec.Replicas)
		}

		backendRuntimeName := inferenceapi.DefaultBackend
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.BackendName != nil {
			backendRuntimeName = *playground.Spec.BackendRuntimeConfig.BackendName
		}
		backendRuntime := inferenceapi.BackendRuntime{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: string(backendRuntimeName)}, &backendRuntime); err != nil {
			return errors.New("failed to get backendRuntime")
		}

		parser := backendruntime.NewBackendRuntimeParser(&backendRuntime, models, playground)

		if service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Name != modelSource.MODEL_RUNNER_CONTAINER_NAME {
			return fmt.Errorf("container name not right, want %s, got %s", modelSource.MODEL_RUNNER_CONTAINER_NAME, service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Name)
		}

		// compare fields both backendRuntime and playground can configure.

		sharedMemorySize := parser.SharedMemorySize()
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.SharedMemorySize != nil {
			sharedMemorySize = playground.Spec.BackendRuntimeConfig.SharedMemorySize
		}
		if sharedMemorySize != nil {
			if *sharedMemorySize != *service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Volumes[0].EmptyDir.SizeLimit {
				return fmt.Errorf("expected SharedMemorySize %s, got %s", sharedMemorySize.String(), service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Volumes[0].EmptyDir.SizeLimit.String())
			}
		}

		resources := parser.Resources()
		if resources == nil {
			resources = &inferenceapi.ResourceRequirements{}
		}
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Resources != nil {
			resources = playground.Spec.BackendRuntimeConfig.Resources
		}
		for k, v := range resources.Limits {
			if !service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k].Equal(v) {
				return fmt.Errorf("unexpected limits for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k])
			}
		}
		for k, v := range resources.Requests {
			if !service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k].Equal(v) {
				return fmt.Errorf("unexpected requests for %s, want %v, got %v", k, v, service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k])
			}
		}

		version := parser.Version()
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Version != nil {
			version = *playground.Spec.BackendRuntimeConfig.Version
		}
		if parser.Image(version) != service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Image {
			return fmt.Errorf("expected container image %s, got %s", parser.Image(version), service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Image)
		}

		envs := parser.Envs()
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Envs != nil {
			envs = playground.Spec.BackendRuntimeConfig.Envs
		}
		if diff := cmp.Diff(envs, service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Env); diff != "" {
			return fmt.Errorf("unexpected envs")
		}

		args, err := parser.Args()
		if err != nil {
			return err
		}
		if playground.Spec.BackendRuntimeConfig != nil && playground.Spec.BackendRuntimeConfig.Args != nil {
			args = append(args, playground.Spec.BackendRuntimeConfig.Args...)
		}

		for _, arg := range args {
			if !slices.Contains(service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Args, arg) {
				return fmt.Errorf("didn't contain arg: %s", arg)
			}
		}

		// compare commands
		if diff := cmp.Diff(parser.Commands(), service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].Command); diff != "" {
			return errors.New("command not right")
		}

		// compare fields only can be configured in backend.

		if backendRuntime.Spec.StartupProbe != nil {
			if diff := cmp.Diff(*service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].StartupProbe, *backendRuntime.Spec.StartupProbe); diff != "" {
				return fmt.Errorf("unexpected startupProbe")
			}
		}
		if backendRuntime.Spec.LivenessProbe != nil {
			if diff := cmp.Diff(*service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].LivenessProbe, *backendRuntime.Spec.LivenessProbe); diff != "" {
				return fmt.Errorf("unexpected livenessProbe")
			}
		}
		if backendRuntime.Spec.ReadinessProbe != nil {
			if diff := cmp.Diff(*service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0].ReadinessProbe, *backendRuntime.Spec.ReadinessProbe); diff != "" {
				return fmt.Errorf("unexpected readinessProbe")
			}
		}

		return nil

	}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
}

// Verify the condition field of status.
func ValidatePlaygroundConditionEqualTo(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground, conditionType string, reason string, status metav1.ConditionStatus) {
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

// Verify the whole fields of status.
func ValidatePlaygroundStatusEqualTo(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground, conditionType string, reason string, status metav1.ConditionStatus) {
	testType := os.Getenv("TEST_TYPE")
	timeout := util.IntegrationTimeout
	interval := util.Interval

	if testType == "E2E" {
		timeout = util.E2ETimeout
		interval = util.E2EInterval
	}

	ValidatePlaygroundConditionEqualTo(ctx, k8sClient, playground, conditionType, reason, status)

	gomega.Eventually(func() error {
		newPlayground := inferenceapi.Playground{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &newPlayground); err != nil {
			return err
		}

		service := inferenceapi.Service{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &service); err != nil {
			return errors.New("failed to get inferenceService")
		}

		if newPlayground.Status.Selector != service.Status.Selector {
			return fmt.Errorf("expected selector %s, got %s", service.Status.Selector, newPlayground.Status.Selector)
		}
		if newPlayground.Status.Replicas != service.Status.Replicas {
			return fmt.Errorf("expected replicas %d, got %d", service.Status.Replicas, newPlayground.Status.Replicas)
		}

		return nil
	}, timeout, interval).Should(gomega.Succeed())
}
