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

package e2e

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	testing "sigs.k8s.io/lws/test/testutils"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
	"github.com/inftyai/llmaz/test/util/validation"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("playground e2e tests", func() {

	// Each test runs in a separate namespace.
	var ns *corev1.Namespace

	ginkgo.BeforeEach(func() {
		// Create test namespace before each test.
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, ns)).To(gomega.Succeed())
	})

	ginkgo.AfterEach(func() {
		gomega.Expect(testing.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
	})

	ginkgo.It("Deploy a ollama model with ollama", func() {
		backendRuntime := wrapper.MakeBackendRuntime("llmaz-ollama").
			Image("ollama/ollama").Version("latest").
			Command([]string{"sh", "-c"}).
			Arg("default", []string{"ollama serve & while true;do output=$(ollama list 2>&1);if ! echo $output | grep -q 'could not connect to ollama app' && echo $output | grep -q 'NAME';then echo 'ollama is running';break; else echo 'Waiting for the ollama to be running...';sleep 1;fi;done;ollama run {{.ModelName}};while true;do sleep 60;done"}).
			Request("default", "cpu", "1").Request("default", "memory", "2Gi").Limit("default", "cpu", "2").Limit("default", "memory", "4Gi").Obj()
		gomega.Expect(k8sClient.Create(ctx, backendRuntime)).To(gomega.Succeed())

		model := wrapper.MakeModel("qwen2-0--5b").FamilyName("qwen2").ModelSourceWithURI("ollama://qwen2:0.5b").Obj()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()
		playground := wrapper.MakePlayground("qwen2-0--5b", ns.Name).ModelClaim("qwen2-0--5b").BackendRuntime("llmaz-ollama").BackendRuntimeEnv("OLLAMA_HOST", "0.0.0.0:8080").Replicas(1).Obj()
		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)
	})
	ginkgo.It("Deploy a huggingface model with llama.cpp", func() {
		model := wrapper.MakeModel("qwen2-0-5b-gguf").FamilyName("qwen2").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", nil, nil).Obj()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()

		playground := wrapper.MakePlayground("qwen2-0-5b-gguf", ns.Name).ModelClaim("qwen2-0-5b-gguf").BackendRuntime("llamacpp").Replicas(2).Obj()
		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)
		gomega.Expect(validation.ValidateServiceAvaliable(ctx, k8sClient, cfg, service, validation.CheckServiceAvaliable)).To(gomega.Succeed())
	})
	ginkgo.It("Deploy a huggingface model with customized backendRuntime", func() {
		backendRuntime := wrapper.MakeBackendRuntime("llmaz-llamacpp").
			Image("ghcr.io/ggerganov/llama.cpp").Version("server").
			Command([]string{"./llama-server"}).
			Arg("default", []string{"-m", "{{.ModelPath}}", "--host", "0.0.0.0", "--port", "8080"}).
			Request("default", "cpu", "2").Request("default", "memory", "4Gi").Limit("default", "cpu", "4").Limit("default", "memory", "4Gi").Obj()
		gomega.Expect(k8sClient.Create(ctx, backendRuntime)).To(gomega.Succeed())

		model := wrapper.MakeModel("qwen2-0-5b-gguf").FamilyName("qwen2").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", nil, nil).Obj()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()

		playground := wrapper.MakePlayground("qwen2-0-5b-gguf", ns.Name).ModelClaim("qwen2-0-5b-gguf").BackendRuntime("llmaz-llamacpp").Replicas(1).Obj()
		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)
		gomega.Expect(validation.ValidateServiceAvaliable(ctx, k8sClient, cfg, service, validation.CheckServiceAvaliable)).To(gomega.Succeed())
	})
	ginkgo.It("Deploy a huggingface model with llama.cpp, HPA enabled", func() {
		model := wrapper.MakeModel("qwen2-0-5b-gguf").FamilyName("qwen2").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", nil, nil).Obj()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()

		playground := wrapper.MakePlayground("qwen2-0-5b-gguf", ns.Name).ModelClaim("qwen2-0-5b-gguf").
			BackendRuntime("llamacpp").ElasticConfig(1, 10).HPA(util.MockASimpleHPATrigger()).
			Replicas(1).Obj()
		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)

		hpa := &autoscalingv2.HorizontalPodAutoscaler{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, hpa)).To(gomega.Succeed())
	})
	ginkgo.It("SpeculativeDecoding with llama.cpp", func() {
		targetModel := wrapper.MakeModel("llama2-7b-q8-gguf").FamilyName("llama2").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("TheBloke/Llama-2-7B-GGUF", "llama-2-7b.Q8_0.gguf", "", nil, nil).Obj()
		gomega.Expect(k8sClient.Create(ctx, targetModel)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, targetModel)).To(gomega.Succeed())
		}()
		draftModel := wrapper.MakeModel("llama2-7b-q2-k-gguf").FamilyName("llama2").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("TheBloke/Llama-2-7B-GGUF", "llama-2-7b.Q2_K.gguf", "", nil, nil).Obj()
		gomega.Expect(k8sClient.Create(ctx, draftModel)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, draftModel)).To(gomega.Succeed())
		}()

		playground := wrapper.MakePlayground("llamacpp-speculator", ns.Name).
			ModelClaims([]string{"llama2-7b-q8-gguf", "llama2-7b-q2-k-gguf"}, []string{"main", "draft"}).
			BackendRuntime("llamacpp").Replicas(1).Obj()
		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)
	})

	ginkgo.It("Deploy huggingface model with llmaz.io/skip-model-loader annotation", func() {
		model := wrapper.MakeModel("deepseek-r1-distill-qwen-1-5b").FamilyName("deepseek").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("deepseek-ai/DeepSeek-R1-Distill-Qwen-1.5B", "", "", nil, nil).Obj()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()

		playground := wrapper.MakePlayground("deepseek-r1-distill-qwen-1-5b", ns.Name).
			ModelClaim("deepseek-r1-distill-qwen-1-5b").
			BackendRuntime("vllm").Replicas(1).Obj()

		if playground.Annotations == nil {
			playground.Annotations = make(map[string]string)
		}
		playground.Annotations["llmaz.io/skip-model-loader"] = "true"

		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)
		gomega.Expect(validation.ValidateServiceAvaliable(ctx, k8sClient, cfg, service, validation.CheckServiceAvaliable)).To(gomega.Succeed())
	})

	ginkgo.It("Deploy S3 model with llmaz.io/skip-model-loader annotation", func() {
		model := wrapper.MakeModel("deepseek-r1-distill-qwen-1-5b").FamilyName("deepseek").ModelSourceWithURI("s3://test-bucket/DeepSeek-R1-Distill-Qwen-1.5B").Obj()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()

		playground := wrapper.MakePlayground("deepseek-r1-distill-qwen-1-5b", ns.Name).
			ModelClaim("deepseek-r1-distill-qwen-1-5b").
			BackendRuntime("vllm").Replicas(1).Obj()

		if playground.Annotations == nil {
			playground.Annotations = make(map[string]string)
		}
		playground.Annotations["llmaz.io/skip-model-loader"] = "true"

		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)
		gomega.Expect(validation.ValidateServiceAvaliable(ctx, k8sClient, cfg, service, validation.CheckServiceAvaliable)).To(gomega.Succeed())
	})
})
