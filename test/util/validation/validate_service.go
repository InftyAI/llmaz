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
	"io"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apimeta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"sigs.k8s.io/controller-runtime/pkg/client"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/pkg"
	helper "github.com/inftyai/llmaz/pkg/controller_helper"
	modelSource "github.com/inftyai/llmaz/pkg/controller_helper/modelsource"
	pkgUtil "github.com/inftyai/llmaz/pkg/util"
	"github.com/inftyai/llmaz/test/util"
)

func ValidateService(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
	gomega.Eventually(func() error {
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, service); err != nil {
			return errors.New("failed to get service")
		}

		workload := lws.LeaderWorkerSet{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, &workload); err != nil {
			return errors.New("failed to get lws")
		}
		if *service.Spec.Replicas != *workload.Spec.Replicas {
			return fmt.Errorf("unexpected replicas %d, got %d", *service.Spec.Replicas, *workload.Spec.Replicas)
		}

		models := []*coreapi.OpenModel{}
		for _, mr := range service.Spec.ModelClaims.Models {
			model := &coreapi.OpenModel{}
			if err := k8sClient.Get(ctx, types.NamespacedName{Name: string(mr.Name)}, model); err != nil {
				return errors.New("failed to get model")
			}
			models = append(models, model)
		}

		for index, model := range models {
			if helper.SkipModelLoader(service) {
				if service.Spec.WorkloadTemplate.LeaderTemplate != nil {
					if err := ValidateSkipModelLoader(model, index, *workload.Spec.LeaderWorkerTemplate.LeaderTemplate, service); err != nil {
						return err
					}
				}
				if err := ValidateSkipModelLoader(model, index, workload.Spec.LeaderWorkerTemplate.WorkerTemplate, service); err != nil {
					return err
				}
			} else {
				// Validate injecting modelLoaders
				if service.Spec.WorkloadTemplate.LeaderTemplate != nil {
					if err := ValidateModelLoader(model, index, *workload.Spec.LeaderWorkerTemplate.LeaderTemplate, service); err != nil {
						return err
					}
				}
				if err := ValidateModelLoader(model, index, workload.Spec.LeaderWorkerTemplate.WorkerTemplate, service); err != nil {
					return err
				}
			}
		}

		mainModel := models[0]
		if workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Labels[coreapi.ModelNameLabelKey] != mainModel.Name {
			return fmt.Errorf("unexpected model name %s in template, want %s", workload.Labels[coreapi.ModelNameLabelKey], mainModel.Name)
		}
		if workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Labels[coreapi.ModelFamilyNameLabelKey] != string(mainModel.Spec.FamilyName) {
			return fmt.Errorf("unexpected model family name %s in template, want %s", workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Labels[coreapi.ModelFamilyNameLabelKey], mainModel.Spec.FamilyName)
		}

		// Validate injecting flavors.
		if mainModel.Spec.InferenceConfig != nil && len(mainModel.Spec.InferenceConfig.Flavors) != 0 {
			if err := ValidateModelFlavor(service, mainModel, &workload); err != nil {
				return err
			}
		}

		if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.Name + "-lb", Namespace: service.Namespace}, &corev1.Service{}); err != nil {
			return err
		}

		if err := ValidateConfigmap(ctx, k8sClient, service); err != nil {
			return err
		}

		return nil
	}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
}

func ValidateModelLoader(model *coreapi.OpenModel, index int, template corev1.PodTemplateSpec, service *inferenceapi.Service) error {
	if model.Spec.Source.URI != nil {
		protocol, _, _ := pkgUtil.ParseURI(string(*model.Spec.Source.URI))
		if protocol == modelSource.Ollama {
			return nil
		}
	}
	if model.Spec.Source.ModelHub != nil || model.Spec.Source.URI != nil {
		if len(template.Spec.InitContainers) == 0 {
			return errors.New("no initContainer configured")
		}

		initContainer := template.Spec.InitContainers[index]

		containerName := modelSource.MODEL_LOADER_CONTAINER_NAME
		if index != 0 {
			containerName += "-" + strconv.Itoa(index)
		}
		if initContainer.Name != containerName {
			return fmt.Errorf("unexpected initContainer name, want %s, got %s", modelSource.MODEL_LOADER_CONTAINER_NAME, initContainer.Name)
		}
		if initContainer.Image != pkg.LOADER_IMAGE {
			return fmt.Errorf("unexpected initContainer image, want %s, got %s", pkg.LOADER_IMAGE, initContainer.Image)
		}

		var envStrings []string

		if model.Spec.Source.ModelHub != nil {
			envStrings = []string{"MODEL_SOURCE_TYPE", "MODEL_ID", "MODEL_HUB_NAME", modelSource.HUGGING_FACE_TOKEN_KEY, modelSource.HUGGING_FACE_HUB_TOKEN}
			if model.Spec.Source.ModelHub.Revision != nil {
				envStrings = append(envStrings, "REVISION")
			}
			if model.Spec.Source.ModelHub.AllowPatterns != nil {
				envStrings = append(envStrings, "MODEL_ALLOW_PATTERNS")
			}
			if model.Spec.Source.ModelHub.IgnorePatterns != nil {
				envStrings = append(envStrings, "MODEL_IGNORE_PATTERNS")
			}
		}
		if model.Spec.Source.URI != nil {
			envStrings = []string{"MODEL_SOURCE_TYPE", "PROVIDER", "ENDPOINT", "BUCKET", "MODEL_PATH", modelSource.OSS_ACCESS_KEY_ID, modelSource.OSS_ACCESS_KEY_SECRET}
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

		container := service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Containers[0]
		for _, v := range container.VolumeMounts {
			if v.Name == modelSource.MODEL_VOLUME_NAME && v.MountPath != modelSource.CONTAINER_MODEL_PATH {
				return fmt.Errorf("unexpected mount path, want %s, got %s", modelSource.CONTAINER_MODEL_PATH, v.MountPath)
			}
		}

		for _, v := range service.Spec.WorkloadTemplate.WorkerTemplate.Spec.Volumes {
			if v.Name == modelSource.MODEL_VOLUME_NAME {
				if v.HostPath == nil || v.HostPath.Path != modelSource.HOST_CLUSTER_MODEL_PATH || *v.HostPath.Type != corev1.HostPathDirectoryOrCreate {
					return errors.New("when using modelHub modelSource, the hostPath shouldn't be nil")
				}
			}
		}
	}
	return nil
}

func ValidateModelFlavor(service *inferenceapi.Service, model *coreapi.OpenModel, workload *lws.LeaderWorkerSet) error {
	flavorName := model.Spec.InferenceConfig.Flavors[0].Name
	if len(service.Spec.ModelClaims.InferenceFlavors) > 0 {
		flavorName = service.Spec.ModelClaims.InferenceFlavors[0]
	}

	for _, flavor := range model.Spec.InferenceConfig.Flavors {
		if flavor.Name == flavorName {
			limits := flavor.Limits
			container := workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0]
			for k, v := range limits {
				if !container.Resources.Requests[k].Equal(v) {
					return fmt.Errorf("unexpected request value %v, got %v", v, workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Requests[k])
				}
				if !container.Resources.Limits[k].Equal(v) {
					return fmt.Errorf("unexpected limit value %v, got %v", v, workload.Spec.LeaderWorkerTemplate.WorkerTemplate.Spec.Containers[0].Resources.Limits[k])
				}
			}
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
			if condition.Reason != reason {
				return fmt.Errorf("reason not right, want %s, got %s", reason, condition.Reason)
			}
			if condition.Status != status {
				return fmt.Errorf("status not right, want %s, got %s", status, string(condition.Status))
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
		if len(pods.Items) != int(*service.Spec.Replicas)*int(*service.Spec.WorkloadTemplate.Size) {
			return fmt.Errorf("pods number not right, want: %d, got: %d", int(*service.Spec.Replicas)*int(*service.Spec.WorkloadTemplate.Size), len(pods.Items))
		}
		return nil
	}).Should(gomega.Succeed())
}

// ValidateSkipModelLoader validates the model-loader initContainer is not injected into the template
// and checks if the model-runner contains model credentials environment variables
func ValidateSkipModelLoader(model *coreapi.OpenModel, index int, template corev1.PodTemplateSpec, service *inferenceapi.Service) error {
	if model.Spec.Source.URI != nil {
		protocol, _, _ := pkgUtil.ParseURI(string(*model.Spec.Source.URI))
		if protocol == modelSource.Ollama {
			return nil
		}
	}

	if model.Spec.Source.ModelHub != nil || model.Spec.Source.URI != nil {
		// Check if the template does not contain the model-loader initContainer
		containerName := modelSource.MODEL_LOADER_CONTAINER_NAME
		if index != 0 {
			containerName += "-" + strconv.Itoa(index)
		}

		for _, container := range template.Spec.InitContainers {
			if container.Name == containerName {
				return fmt.Errorf("template has model-loader initContainer: %s", container.Name)
			}
		}

		for _, container := range template.Spec.Containers {
			if container.Name == modelSource.MODEL_RUNNER_CONTAINER_NAME {
				// Check if the model-runner container contains model credentials environment variables
				var envStrings []string
				if model.Spec.Source.ModelHub != nil {
					envStrings = append(envStrings, modelSource.HUGGING_FACE_TOKEN_KEY, modelSource.HUGGING_FACE_HUB_TOKEN)
				} else if model.Spec.Source.URI != nil {
					protocol, _, _ := pkgUtil.ParseURI(string(*model.Spec.Source.URI))
					if protocol == modelSource.S3 || protocol == modelSource.GCS {
						envStrings = append(envStrings, modelSource.AWS_ACCESS_KEY_ID, modelSource.AWS_ACCESS_KEY_SECRET)
					} else if protocol == modelSource.OSS {
						envStrings = append(envStrings, modelSource.OSS_ACCESS_KEY_ID, modelSource.OSS_ACCESS_KEY_SECRET)
					}
				}

				for _, str := range envStrings {
					envExist := false
					for _, env := range container.Env {
						if env.Name == str {
							envExist = true
							break
						}
					}
					if !envExist {
						return fmt.Errorf("env %s doesn't exist", str)
					}
				}

				// The model-runner container should not mount the model-volume if the model-loader initContainer is not injected
				for _, v := range container.VolumeMounts {
					if v.Name == modelSource.MODEL_VOLUME_NAME {
						return fmt.Errorf("model-runner container has volume mount %s", v.Name)
					}
				}
			}
		}

		for _, v := range template.Spec.Volumes {
			if v.Name == modelSource.MODEL_VOLUME_NAME {
				return errors.New("when skip the model-loader initContainer, the model-volume should not be created")
			}
		}
	}

	return nil
}

type CheckServiceAvailableFunc func() error

func ValidateServiceAvaliable(ctx context.Context, k8sClient client.Client, cfg *rest.Config, service *inferenceapi.Service, check CheckServiceAvailableFunc) error {
	pods := corev1.PodList{}
	podSelector := client.MatchingLabels(map[string]string{
		lws.SetNameLabelKey: service.Name,
	})
	if err := k8sClient.List(ctx, &pods, podSelector, client.InNamespace(service.Namespace)); err != nil {
		return err
	}
	if len(pods.Items) != int(*service.Spec.Replicas)*int(*service.Spec.WorkloadTemplate.Size) {
		return fmt.Errorf("pods number not right, want: %d, got: %d", int(*service.Spec.Replicas)*int(*service.Spec.WorkloadTemplate.Size), len(pods.Items))
	}

	var targetPod *corev1.Pod
	for i := range pods.Items {
		if pods.Items[i].Status.Phase == corev1.PodRunning {
			targetPod = &pods.Items[i]
			break
		}
	}

	if targetPod == nil {
		return fmt.Errorf("no running pods found for service %s", service.Name)
	}

	portForwardK8sClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return fmt.Errorf("init port forward client failed: %w", err)
	}

	targetPort := targetPod.Spec.Containers[0].Ports[0].ContainerPort
	stopChan, readyChan := make(chan struct{}, 1), make(chan struct{}, 1)
	req := portForwardK8sClient.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(service.Namespace).
		Name(targetPod.Name).
		SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(cfg)
	if err != nil {
		return fmt.Errorf("creating round tripper failed: %v", err)
	}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", req.URL())
	// create port forwarder
	fw, err := portforward.New(dialer, []string{fmt.Sprintf("%d:%d", modelSource.DEFAULT_BACKEND_PORT, targetPort)}, stopChan, readyChan, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("creating port forwarder failed: %v", err)
	}
	// stop port forward when done
	defer fw.Close()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		fmt.Println("Received termination signal, shutting down port forward...")
		close(stopChan)
	}()

	// wait for port forward to be ready
	go func() {
		if err = fw.ForwardPorts(); err != nil {
			fmt.Printf("Error forwarding ports: %v\n", err)
			close(stopChan)
		}
	}()
	<-readyChan
	gomega.Eventually(check()).Should(gomega.Succeed())
	return nil
}

func CheckServiceAvaliable() error {
	url := fmt.Sprintf("http://localhost:%d/completions", modelSource.DEFAULT_BACKEND_PORT)
	reqBody := `{
		"prompt": "<|im_start|>user\nWhat is the capital city of China?\n<|im_end|>\n<|im_start|>assistant\n",
		"stream": false
	  }`

	req, err := http.NewRequest("POST", url, strings.NewReader(reqBody))
	if err != nil {
		return err
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error HTTP status code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	if !strings.Contains(strings.ToLower(string(body)), "beijing") {
		return fmt.Errorf("error response body: %s", string(body))
	}
	return nil
}

func ValidateConfigmap(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) error {
	cm := corev1.ConfigMap{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: "llmaz-global-config", Namespace: "llmaz-system"}, &cm); err != nil {
		return err
	}

	data, err := helper.ParseGlobalConfigmap(&cm)
	if err != nil {
		return fmt.Errorf("failed to parse global configmap: %v", err)
	}

	if service.Spec.WorkloadTemplate.LeaderTemplate != nil {
		if service.Spec.WorkloadTemplate.LeaderTemplate.Spec.SchedulerName != data.SchedulerName {
			return fmt.Errorf("unexpected scheduler name %s, want %s", service.Spec.WorkloadTemplate.LeaderTemplate.Spec.SchedulerName, data.SchedulerName)
		}
	}

	if service.Spec.WorkloadTemplate.WorkerTemplate.Spec.SchedulerName != data.SchedulerName {
		return fmt.Errorf("unexpected scheduler name %s, want %s", service.Spec.WorkloadTemplate.WorkerTemplate.Spec.SchedulerName, data.SchedulerName)
	}

	return nil
}
