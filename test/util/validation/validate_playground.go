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
	"slices"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	"inftyai.com/llmaz/pkg/controller_helper/backend"
	"inftyai.com/llmaz/test/util"
)

func ValidatePlayground(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
	gomega.Eventually(func() error {
		service := inferenceapi.Service{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &service); err != nil {
			return errors.New("failed to get inferenceService")
		}

		if *playground.Spec.Replicas != *service.Spec.WorkloadTemplate.Replicas {
			return fmt.Errorf("expected replicas: %d, got %d", *playground.Spec.Replicas, *service.Spec.WorkloadTemplate.Replicas)
		}

		model := coreapi.Model{}

		if playground.Spec.ModelClaim != nil {
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: string(playground.Spec.ModelClaim.ModelName), Namespace: playground.Namespace}, &model); err != nil {
				return errors.New("failed to get model")
			}

			if playground.Spec.ModelClaim.ModelName != service.Spec.MultiModelsClaims[0].ModelNames[0] {
				return fmt.Errorf("expected modelName %s, got %s", playground.Spec.ModelClaim.ModelName, service.Spec.MultiModelsClaims[0].ModelNames[0])
			}
			if diff := cmp.Diff(playground.Spec.ModelClaim.InferenceFlavors, service.Spec.MultiModelsClaims[0].InferenceFlavors); diff != "" {
				return fmt.Errorf("unexpected flavors, want %v, got %v", playground.Spec.ModelClaim.InferenceFlavors, service.Spec.MultiModelsClaims[0].InferenceFlavors)
			}
		}

		// TODO: MultiModelsClaim

		if playground.Spec.BackendConfig != nil {
			backendName := inferenceapi.DefaultBackend
			if playground.Spec.BackendConfig.Name != nil {
				backendName = *playground.Spec.BackendConfig.Name
			}
			if playground.Spec.BackendConfig.Version != nil && backend.SwitchBackend(backendName).Image(*playground.Spec.BackendConfig.Version) != service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image {
				return fmt.Errorf("expected container image %s, got %s", backend.SwitchBackend(backendName).Image(*playground.Spec.BackendConfig.Version), service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Image)
			}
			if playground.Spec.BackendConfig.Resources != nil {
				if playground.Spec.BackendConfig.Resources.Limits != nil {
					if diff := cmp.Diff(playground.Spec.BackendConfig.Resources.Limits, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits); diff != "" {
						return errors.New("unexpected resource limits")
					}
				}
				if playground.Spec.BackendConfig.Resources.Requests != nil {
					if diff := cmp.Diff(playground.Spec.BackendConfig.Resources.Requests, service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests); diff != "" {
						return errors.New("unexpected resource requests")
					}
				}
			}
			for _, arg := range playground.Spec.BackendConfig.Args {
				if !slices.Contains(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Args, arg) {
					return fmt.Errorf("didn't contain arg: %s", arg)
				}
			}
			if diff := cmp.Diff(service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Env, playground.Spec.BackendConfig.Envs); diff != "" {
				return fmt.Errorf("unexpected envs")
			}
		}
		// TODO: validate multiModelsClaims config

		return nil

	}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
}

func ValidatePlaygroundStatusEqualTo(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground, conditionType string, reason string, status metav1.ConditionStatus) {
	gomega.Eventually(func() error {
		newPlayground := inferenceapi.Playground{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &newPlayground); err != nil {
			return err
		}
		if condition := apimeta.FindStatusCondition(newPlayground.Status.Conditions, conditionType); condition == nil {
			return errors.New("condition not found")
		} else {
			if condition.Reason != reason || condition.Status != status {
				return errors.New("reason or status not right")
			}
		}
		return nil
	})
}
