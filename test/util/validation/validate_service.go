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

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	"inftyai.com/llmaz/pkg"
	modelSource "inftyai.com/llmaz/pkg/controller_helper/model_source"
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

		// TODO: multiModelsClaim
		// TODO: multi-host

		modelName := string(service.Spec.MultiModelsClaims[0].ModelNames[0])
		model := coreapi.OpenModel{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: modelName}, &model); err != nil {
			return errors.New("failed to get model")
		}

		if workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Labels[coreapi.ModelNameLabelKey] != model.Name {
			return fmt.Errorf("unexpected model name %s in template, want %s", workload.Labels[coreapi.ModelNameLabelKey], model.Name)
		}
		if workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Labels[coreapi.ModelFamilyNameLabelKey] != string(model.Spec.FamilyName) {
			return fmt.Errorf("unexpected model family name %s in template, want %s", workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Labels[coreapi.ModelFamilyNameLabelKey], model.Spec.FamilyName)
		}

		// Validate injecting modelLoaders
		if err := ValidateModelLoader(&model, &workload, service); err != nil {
			return err
		}

		// Validate injecting flavors.
		if len(model.Spec.InferenceFlavors) != 0 {
			if err := ValidateModelFlavor(&model, &workload); err != nil {
				return err
			}
		}
		return nil

	}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
}

func ValidateModelLoader(model *coreapi.OpenModel, workload *lws.LeaderWorkerSet, service *inferenceapi.Service) error {
	if model.Spec.Source.ModelHub != nil || model.Spec.Source.URI != nil {
		if len(workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.InitContainers) == 0 {
			return errors.New("no initContainer configured")
		}

		initContainer := workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.InitContainers[0]

		if initContainer.Name != modelSource.MODEL_LOADER_CONTAINER_NAME {
			return fmt.Errorf("unexpected initContainer name, want %s, got %s", modelSource.MODEL_LOADER_CONTAINER_NAME, initContainer.Name)
		}
		if initContainer.Image != pkg.LOADER_IMAGE {
			return fmt.Errorf("unexpected initContainer image, want %s, got %s", pkg.LOADER_IMAGE, initContainer.Image)
		}

		var envStrings []string

		if model.Spec.Source.ModelHub != nil {
			envStrings = []string{"MODEL_SOURCE_TYPE", "MODEL_ID", "MODEL_HUB_NAME", "HF_TOKEN", "HUGGING_FACE_HUB_TOKEN"}
			if model.Spec.Source.ModelHub.Revision != nil {
				envStrings = append(envStrings, "REVISION")
			}
		}
		if model.Spec.Source.URI != nil {
			envStrings = []string{"MODEL_SOURCE_TYPE", "PROVIDER", "ENDPOINT", "BUCKET", "MODEL_PATH", "OSS_ACCESS_KEY_ID", "OSS_ACCESS_KEY_SECRET"}
		}

		for _, str := range envStrings {
			envExist := false
			for _, env := range initContainer.Env {
				if env.Name == str {
					envExist = true
					break
				}
			}
			if !envExist {
				return fmt.Errorf("env %s doesn't exist", str)
			}
		}
		for _, v := range initContainer.VolumeMounts {
			if v.Name == modelSource.MODEL_VOLUME_NAME && v.MountPath != modelSource.CONTAINER_MODEL_PATH {
				return fmt.Errorf("unexpected mount path, want %s, got %s", modelSource.CONTAINER_MODEL_PATH, v.MountPath)
			}
		}

		container := service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0]
		for _, v := range container.VolumeMounts {
			if v.Name == modelSource.MODEL_VOLUME_NAME && v.MountPath != modelSource.CONTAINER_MODEL_PATH {
				return fmt.Errorf("unexpected mount path, want %s, got %s", modelSource.CONTAINER_MODEL_PATH, v.MountPath)
			}
		}

		for _, v := range service.Spec.WorkloadTemplate.LeaderWorkerTemplate.WorkerTemplate.Spec.Volumes {
			if v.Name == modelSource.MODEL_VOLUME_NAME {
				if v.HostPath == nil || v.HostPath.Path != modelSource.HOST_CLUSTER_MODEL_PATH || *v.HostPath.Type != corev1.HostPathDirectoryOrCreate {
					return errors.New("when using modelHub modelSource, the hostPath shouldn't be nil")
				}
			}
		}
	}
	return nil
}

func ValidateModelFlavor(model *coreapi.OpenModel, workload *lws.LeaderWorkerSet) error {
	// TODO: Use the 0-index flavor for validation right now.
	flavor := model.Spec.InferenceFlavors[0]

	requests := flavor.Requests
	container := workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0]
	for k, v := range requests {
		if !container.Resources.Requests[k].Equal(v) {
			return fmt.Errorf("unexpected request value %v, got %v", v, workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k])
		}
		if !container.Resources.Limits[k].Equal(v) {
			return fmt.Errorf("unexpected limit value %v, got %v", v, workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k])
		}
	}

	if len(flavor.NodeSelector) != 0 {
		terms := workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms
		requirements := []corev1.NodeSelectorRequirement{}
		for k, v := range flavor.NodeSelector {
			requirements = append(requirements, corev1.NodeSelectorRequirement{
				Key:      k,
				Values:   []string{v},
				Operator: corev1.NodeSelectorOpIn,
			})
		}
		if diff := cmp.Diff(terms, []corev1.NodeSelectorTerm{
			{MatchExpressions: requirements},
		}); diff != "" {
			return errors.New("unexpected nodeSelectors")
		}
	}
	return nil
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
	}).Should(gomega.Succeed())
}

// This can only be used in e2e test because of integration test has no lws controllers, no pods will be created.
func ValidateServicePods(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
	gomega.Eventually(func() error {
		pods := corev1.PodList{}
		podSelector := client.MatchingLabels(map[string]string{
			lws.SetNameLabelKey: service.Name,
		})
		if err := k8sClient.List(ctx, &pods, podSelector, client.InNamespace(service.Namespace)); err != nil {
			return err
		}
		if len(pods.Items) != int(*service.Spec.WorkloadTemplate.Replicas)*int(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.Size) {
			return fmt.Errorf("pods number not right, want: %d, got: %d", int(*service.Spec.WorkloadTemplate.Replicas)*int(*service.Spec.WorkloadTemplate.LeaderWorkerTemplate.Size), len(pods.Items))
		}
		return nil
	}).Should(gomega.Succeed())
}
