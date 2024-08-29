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

package e2e

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	testing "sigs.k8s.io/lws/test/testutils"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
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

	ginkgo.It("Deploy a huggingface model with llama.cpp", func() {
		model := wrapper.MakeModel("qwen2-0-5b-gguf").FamilyName("qwen2").ModelSourceWithModelHub("Huggingface").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf").Obj()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()

		playground := wrapper.MakePlayground("qwen2-0-5b-gguf", ns.Name).ModelClaim("qwen2-0-5b-gguf").Backend("llamacpp").Replicas(3).Obj()
		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		validation.ValidatePlayground(ctx, k8sClient, playground)
		validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundAvailable, "PlaygroundReady", metav1.ConditionTrue)

		service := &inferenceapi.Service{}
		gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, service)).To(gomega.Succeed())
		validation.ValidateService(ctx, k8sClient, service)
		validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
		validation.ValidateServicePods(ctx, k8sClient, service)
	})
})
