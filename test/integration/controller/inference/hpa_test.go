/*
Copyright 2025.

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
	"context"
	"fmt"

	"github.com/google/go-cmp/cmp"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("hpa test", func() {
	// Each test runs in a separate namespace.
	var ns *corev1.Namespace
	var model *coreapi.OpenModel

	type update struct {
		updateFunc func(*inferenceapi.Playground)
		checkFunc  func(context.Context, client.Client, *inferenceapi.Playground)
	}

	ginkgo.BeforeEach(func() {
		// Create test namespace before each test.
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "ns-playground-",
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
		makePlayground func() *inferenceapi.Playground
		updates        []*update
	}
	// TODO: Add more testCases to cover updating.
	ginkgo.DescribeTable("test playground creation and update",
		func(tc *testValidatingCase) {
			playground := tc.makePlayground()
			for _, update := range tc.updates {
				if update.updateFunc != nil {
					update.updateFunc(playground)
				}
				newPlayground := &inferenceapi.Playground{}
				gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, newPlayground)).To(gomega.Succeed())
				if update.checkFunc != nil {
					update.checkFunc(ctx, k8sClient, newPlayground)
				}
			}
		},
		ginkgo.Entry("playground with scaleTrigger configured", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).ModelClaim(model.Name).Label(coreapi.ModelNameLabelKey, model.Name).
					ElasticConfig(1, 3).
					HPA(util.MockASimpleHPATrigger()).
					Obj()
			},
			updates: []*update{
				{
					updateFunc: func(playground *inferenceapi.Playground) {
						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						gomega.Eventually(func() error {
							hpa := &autoscalingv2.HorizontalPodAutoscaler{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, hpa); err != nil {
								return err
							}
							if diff := cmp.Diff(playground.Spec.ElasticConfig.ScaleTrigger.HPA.Metrics, hpa.Spec.Metrics); diff != "" {
								return fmt.Errorf("metrics not match: %s", diff)
							}
							return nil
						}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
					},
				},
			},
		}),
		ginkgo.Entry("playground with scaleTrigger configured backendRuntime", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).ModelClaim(model.Name).Label(coreapi.ModelNameLabelKey, model.Name).
					ElasticConfig(1, 3).
					BackendRuntime("fake-backend").
					Obj()
			},
			updates: []*update{
				{
					updateFunc: func(playground *inferenceapi.Playground) {
						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						gomega.Eventually(func() error {
							hpa := &autoscalingv2.HorizontalPodAutoscaler{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, hpa); err != nil {
								return err
							}
							backend := &inferenceapi.BackendRuntime{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: "fake-backend"}, backend); err != nil {
								return err
							}
							if diff := cmp.Diff(backend.Spec.ScaleTriggers[0].HPA.Metrics, hpa.Spec.Metrics); diff != "" {
								return fmt.Errorf("metrics not match: %s", diff)
							}
							return nil
						}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
					},
				},
			},
		}),
		ginkgo.Entry("playground with scaleTrigger overwrite backendRuntime's", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).ModelClaim(model.Name).Label(coreapi.ModelNameLabelKey, model.Name).
					ElasticConfig(1, 3).ScaleTriggerRef("hpa2").
					BackendRuntime("fake-backend").
					Obj()
			},
			updates: []*update{
				{
					updateFunc: func(playground *inferenceapi.Playground) {
						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						gomega.Eventually(func() error {
							hpa := &autoscalingv2.HorizontalPodAutoscaler{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, hpa); err != nil {
								return err
							}
							backend := &inferenceapi.BackendRuntime{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: "fake-backend"}, backend); err != nil {
								return err
							}
							if diff := cmp.Diff(backend.Spec.ScaleTriggers[1].HPA.Metrics, hpa.Spec.Metrics); diff != "" {
								return fmt.Errorf("metrics not match: %s", diff)
							}
							return nil
						}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
					},
				},
			},
		}),
	)
})
