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

package webhook

import (
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("playground default and validation", func() {
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
		playground func() *inferenceapi.Playground
		failed     bool
	}
	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			if tc.failed {
				gomega.Expect(k8sClient.Create(ctx, tc.playground())).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(k8sClient.Create(ctx, tc.playground())).To(gomega.Succeed())
			}
		},
		ginkgo.Entry("normal Playground creation", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).ModelClaim("llama3-8b").Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("invalid name", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground-0.5b", ns.Name).Replicas(1).ModelClaim("llama3-8b").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("no model claim declared", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("sglang backend supporeted", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).ModelClaim("llama3-8b").Backend(string(inferenceapi.SGLANG)).Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("llamacpp backend supporeted", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).ModelClaim("llama3-8b").Backend(string(inferenceapi.LLAMACPP)).Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("speculativeDecoding with SGLang is not allowed", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).MultiModelsClaim([]string{"llama3-405b", "llama3-8b"}, coreapi.SpeculativeDecoding).Backend(string(inferenceapi.SGLANG)).Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("speculativeDecoding with three models claimed", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).MultiModelsClaim([]string{"llama3-405b", "llama3-8b", "llama3-2b"}, coreapi.SpeculativeDecoding).Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("unknown backend configured", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).Backend("unknown").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("unknown inference mode", &testValidatingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).Replicas(1).MultiModelsClaim([]string{"llama3-405b", "llama3-8b"}, coreapi.InferenceMode("unknown")).Obj()
			},
			failed: true,
		}),
	)

	type testDefaultingCase struct {
		playground     func() *inferenceapi.Playground
		wantPlayground func() *inferenceapi.Playground
	}
	ginkgo.DescribeTable("test validating",
		func(tc *testDefaultingCase) {
			playground := tc.playground()
			gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
			gomega.Expect(playground).To(gomega.BeComparableTo(tc.wantPlayground(),
				cmpopts.IgnoreTypes(inferenceapi.PlaygroundStatus{}),
				cmpopts.IgnoreFields(metav1.ObjectMeta{}, "UID", "ResourceVersion", "Generation", "CreationTimestamp", "ManagedFields")))
		},
		ginkgo.Entry("defaulting label with modelClaim", &testDefaultingCase{
			playground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).ModelClaim("llama3-8b").Replicas(1).Obj()
			},
			wantPlayground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).ModelClaim("llama3-8b").Replicas(1).Label(coreapi.ModelNameLabelKey, "llama3-8b").Obj()
			},
		}),
		ginkgo.Entry("defaulting inferenceMode with multiModelsClaim", &testDefaultingCase{
			playground: func() *inferenceapi.Playground {
				playground := wrapper.MakePlayground("playground", ns.Name).Replicas(1).Obj()
				playground.Spec.MultiModelsClaim = &coreapi.MultiModelsClaim{
					ModelNames: []coreapi.ModelName{"llama3-405b", "llama3-8b"},
				}
				return playground
			},
			wantPlayground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).MultiModelsClaim([]string{"llama3-405b", "llama3-8b"}, coreapi.Standard).Replicas(1).Label(coreapi.ModelNameLabelKey, "llama3-405b").Obj()
			},
		}),
	)
})
