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

package inference

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	api "inftyai.com/llmaz/api/core/v1alpha1"
	inferenceapi "inftyai.com/llmaz/api/inference/v1alpha1"
	"inftyai.com/llmaz/test/util"
	"inftyai.com/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("playground controller test", func() {
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
		gomega.Expect(k8sClient.Delete(ctx, ns)).To(gomega.Succeed())
		gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
	})

	type testValidatingCase struct {
		playground func() *inferenceapi.Playground
		failed     bool
	}
	// TODO: Add more testCases to cover updating.
	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			if tc.failed {
				gomega.Expect(k8sClient.Create(ctx, tc.playground())).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(k8sClient.Create(ctx, tc.playground())).To(gomega.Succeed())
			}
		},
		ginkgo.Entry("normal playground creation", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("test-playground", ns.Name).Replicas(1).ModelClaim("llama3-8b").Obj()
			},
			failed: false,
		}),
	)
})
