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
	testing "sigs.k8s.io/lws/test/testutils"

	api "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/test/util"
)

var _ = ginkgo.Describe("playground e2e tests", func() {

	// Each test runs in a separate namespace.
	var ns *corev1.Namespace
	var model *api.Model

	ginkgo.BeforeEach(func() {
		// Create test namespace before each test.
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "test-ns-",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, ns)).To(gomega.Succeed())
		model = util.MockASampleModel()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
	})

	ginkgo.AfterEach(func() {
		gomega.Expect(testing.DeleteNamespace(ctx, k8sClient, ns)).To(gomega.Succeed())
		gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
	})

	ginkgo.It("Can deploy a normal playground", func() {
		playground := util.MockASamplePlayground(ns.Name)
		gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
		// TODO: validate the corresponding inferenceService
	})
})
