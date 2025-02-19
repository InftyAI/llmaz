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

package webhook

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("service default and validation", func() {
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
		gomega.Expect(k8sClient.Delete(ctx, ns)).To(gomega.Succeed())
	})

	type testValidatingCase struct {
		service func() *inferenceapi.Service
		failed  bool
	}

	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			if tc.failed {
				gomega.Expect(k8sClient.Create(ctx, tc.service())).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(k8sClient.Create(ctx, tc.service())).To(gomega.Succeed())
			}
		},
		ginkgo.Entry("normal Service creation", &testValidatingCase{
			service: func() *inferenceapi.Service {
				return util.MockASampleService(ns.Name)
			},
			failed: false,
		}),
		ginkgo.Entry("invalid name", &testValidatingCase{
			service: func() *inferenceapi.Service {
				return wrapper.MakeService("service-0.5b", ns.Name).WorkerTemplate().Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("model-runner container doesn't exist", &testValidatingCase{
			service: func() *inferenceapi.Service {
				return wrapper.MakeService("service-llama3-8b", ns.Name).
					ModelClaims([]string{"llama3-8b"}, []string{"main"}).
					WorkerTemplate().
					ContainerName("model-runner-fake").
					Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("speculative-decoding with three models", &testValidatingCase{
			service: func() *inferenceapi.Service {
				return wrapper.MakeService("service-llama3-8b", ns.Name).
					ModelClaims([]string{"llama3-405b", "llama3-8b", "llama3-2b"}, []string{"main", "draft", "draft"}).
					WorkerTemplate().
					Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("modelClaims with nil role", &testValidatingCase{
			service: func() *inferenceapi.Service {
				service := wrapper.MakeService("service-llama3-8b", ns.Name).
					ModelClaims([]string{"llama3-405b", "llama3-8b"}, []string{"main", "draft"}).
					WorkerTemplate().
					Obj()
				// Set the role to nil
				service.Spec.ModelClaims.Models[0].Role = nil
				return service
			},
			failed: false,
		}),
		ginkgo.Entry("no main model", &testValidatingCase{
			service: func() *inferenceapi.Service {
				return wrapper.MakeService("service-llama3-8b", ns.Name).
					ModelClaims([]string{"llama3-8b"}, []string{"draft"}).
					WorkerTemplate().
					Obj()
			},
			failed: true,
		}),
	)
})
