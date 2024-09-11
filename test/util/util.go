/*
Copyright 2024 The Kubernetes Authors.
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

package util

import (
	"context"
	"fmt"
	"os"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/serializer/yaml"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"
)

func UpdateLwsToReady(ctx context.Context, k8sClient client.Client, name, namespace string) {
	gomega.Eventually(func() error {
		workload := &lws.LeaderWorkerSet{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, workload); err != nil {
			return err
		}
		condition := metav1.Condition{
			Type:    string(lws.LeaderWorkerSetAvailable),
			Status:  metav1.ConditionStatus(corev1.ConditionTrue),
			Reason:  "AllGroupsReady",
			Message: "All replicas are ready",
		}

		changed := apimeta.SetStatusCondition(&workload.Status.Conditions, condition)
		if changed {
			return k8sClient.Status().Update(ctx, workload)
		}
		return nil
	}, IntegrationTimeout, Interval).Should(gomega.Succeed())
}

func UpdateLwsToUnReady(ctx context.Context, k8sClient client.Client, name, namespace string) {
	gomega.Eventually(func() error {
		workload := &lws.LeaderWorkerSet{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, workload); err != nil {
			return err
		}
		condition := metav1.Condition{
			Type:    string(lws.LeaderWorkerSetAvailable),
			Status:  metav1.ConditionStatus(corev1.ConditionFalse),
			Reason:  "AllGroupsReady",
			Message: "All replicas are ready",
		}

		changed := apimeta.SetStatusCondition(&workload.Status.Conditions, condition)
		if changed {
			return k8sClient.Status().Update(ctx, workload)
		}
		return nil
	}, IntegrationTimeout, Interval).Should(gomega.Succeed())
}

func applyYaml(ctx context.Context, k8sClient client.Client, file string) error {
	yamlFile, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed to read YAML file: %v", err)
	}

	decode := yaml.NewDecodingSerializer(unstructured.UnstructuredJSONScheme)
	obj := &unstructured.Unstructured{}
	_, _, err = decode.Decode(yamlFile, nil, obj)
	if err != nil {
		return fmt.Errorf("failed to decode YAML into Unstructured object: %v", err)
	}

	if err = k8sClient.Create(ctx, obj); err != nil {
		return fmt.Errorf("failed to create resource: %v", err)
	}

	return nil
}

func Setup(ctx context.Context, k8sClient client.Client, path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := applyYaml(ctx, k8sClient, path+"/"+entry.Name()); err != nil {
			return err
		}
	}
	return nil
}
