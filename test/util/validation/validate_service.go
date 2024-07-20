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

	"github.com/onsi/gomega"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	"inftyai.com/llmaz/test/util"
)

func ValidateService(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
	gomega.Eventually(func() error {
		workload := lws.LeaderWorkerSet{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, &workload); err != nil {
			return errors.New("failed to get lws")
		}
		if *service.Spec.WorkloadTemplate.Replicas != *workload.Spec.Replicas {
			return fmt.Errorf("unexpected replicas %d, got %d", *service.Spec.WorkloadTemplate.Replicas, *workload.Spec.Replicas)
		}
		return nil

	}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
}

func ValidateServiceStatusEqualTo(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service, conditionType string, reason string, status metav1.ConditionStatus) {
	gomega.Eventually(func() error {
		newService := inferenceapi.Service{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, &newService); err != nil {
			return err
		}
		if condition := apimeta.FindStatusCondition(newService.Status.Conditions, conditionType); condition == nil {
			return errors.New("condition not found")
		} else {
			if condition.Reason != reason || condition.Status != status {
				return errors.New("reason or status not right")
			}
		}
		return nil
	})
}
