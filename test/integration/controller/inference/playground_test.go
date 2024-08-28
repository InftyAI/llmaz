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
	"context"
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
	"github.com/inftyai/llmaz/test/util/validation"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("playground controller test", func() {
	// Each test runs in a separate namespace.
	var ns *corev1.Namespace
	var model *coreapi.OpenModel

	type update struct {
		playgroundUpdateFn func(*inferenceapi.Playground)
		checkPlayground    func(context.Context, client.Client, *inferenceapi.Playground)
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
				if update.playgroundUpdateFn != nil {
					update.playgroundUpdateFn(playground)
				}
				newPlayground := &inferenceapi.Playground{}
				gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, newPlayground)).To(gomega.Succeed())
				if update.checkPlayground != nil {
					update.checkPlayground(ctx, k8sClient, newPlayground)
				}
			}
		},
		ginkgo.Entry("normal Playground create and update", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return util.MockASamplePlayground(ns.Name)
			},
			updates: []*update{
				{
					playgroundUpdateFn: func(playground *inferenceapi.Playground) {
						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						validation.ValidatePlayground(ctx, k8sClient, playground)
						validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundProgressing, "Pending", metav1.ConditionTrue)
					},
				},
				{
					playgroundUpdateFn: func(playground *inferenceapi.Playground) {
						gomega.Eventually(func() error {
							updatePlayground := &inferenceapi.Playground{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, updatePlayground); err != nil {
								return err
							}
							updatePlayground.Spec.Replicas = ptr.To[int32](3)
							if err := k8sClient.Update(ctx, updatePlayground); err != nil {
								return err
							}
							return nil
						}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())

						// To make sure playground updated successfully.
						newPlayground := inferenceapi.Playground{}
						gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &newPlayground)).To(gomega.Succeed())
						gomega.Expect(*newPlayground.Spec.Replicas).To(gomega.Equal(int32(3)))
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						validation.ValidatePlayground(ctx, k8sClient, playground)
						validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundProgressing, "Pending", metav1.ConditionTrue)
					},
				},
			},
		}),
		ginkgo.Entry("advance configured Playground with sglang", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).ModelClaim(model.Name).Label(coreapi.ModelNameLabelKey, model.Name).
					Backend("sglang").BackendVersion("main").BackendArgs([]string{"--foo", "bar"}).BackendEnv("FOO", "BAR").BackendRequest("cpu", "1").BackendLimit("cpu", "10").
					Obj()
			},
			updates: []*update{
				{
					playgroundUpdateFn: func(playground *inferenceapi.Playground) {
						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						validation.ValidatePlayground(ctx, k8sClient, playground)
						validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundProgressing, "Pending", metav1.ConditionTrue)
					},
				},
			},
		}),
		ginkgo.Entry("advance configured Playground with llamacpp", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return wrapper.MakePlayground("playground", ns.Name).ModelClaim(model.Name).Label(coreapi.ModelNameLabelKey, model.Name).
					Backend("llamacpp").BackendVersion("main").BackendArgs([]string{"--foo", "bar"}).BackendEnv("FOO", "BAR").BackendRequest("cpu", "1").BackendLimit("cpu", "10").
					Obj()
			},
			updates: []*update{
				{
					playgroundUpdateFn: func(playground *inferenceapi.Playground) {
						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						validation.ValidatePlayground(ctx, k8sClient, playground)
						validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundProgressing, "Pending", metav1.ConditionTrue)
					},
				},
			},
		}),
		ginkgo.Entry("playground is created when service exists with the same name", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return util.MockASamplePlayground(ns.Name)
			},
			updates: []*update{
				{
					playgroundUpdateFn: func(playground *inferenceapi.Playground) {
						// Create a service with the same name as the playground.
						service := wrapper.MakeService(playground.Name, playground.Namespace).
							ModelsClaim([]string{"llama3-8b"}, []string{}, nil).
							WorkerTemplate().
							Obj()
						gomega.Expect(k8sClient.Create(ctx, service)).To(gomega.Succeed())
						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundProgressing, "AbortProcessing", metav1.ConditionFalse)
					},
				},
				{
					// Delete the service, playground should be updated to Pending.
					playgroundUpdateFn: func(playground *inferenceapi.Playground) {
						service := wrapper.MakeService(playground.Name, playground.Namespace).
							ModelsClaim([]string{"llama3-8b"}, []string{}, nil).
							WorkerTemplate().
							Obj()
						gomega.Expect(k8sClient.Delete(ctx, service)).To(gomega.Succeed())
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						validation.ValidatePlayground(ctx, k8sClient, playground)
						validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundProgressing, "Pending", metav1.ConditionTrue)
					},
				},
			},
		}),
		ginkgo.Entry("create the model after playground is created", &testValidatingCase{
			makePlayground: func() *inferenceapi.Playground {
				return util.MockASamplePlayground(ns.Name)
			},
			updates: []*update{
				{
					playgroundUpdateFn: func(playground *inferenceapi.Playground) {
						// Delete the pre-provision model before creating playground.
						gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
						// To make sure model not exists.
						gomega.Eventually(func() error {
							oldModel := &coreapi.OpenModel{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: model.Name, Namespace: model.Namespace}, oldModel); err != nil {
								if apierrors.IsNotFound(err) {
									return nil
								}
								return err
							}
							return fmt.Errorf("model %s/%s still exists", model.Namespace, model.Name)
						}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())

						gomega.Expect(k8sClient.Create(ctx, playground)).To(gomega.Succeed())
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						gomega.Consistently(func() error {
							updatePlayground := inferenceapi.Playground{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: playground.Name, Namespace: playground.Namespace}, &updatePlayground); err != nil {
								return err
							}
							if len(updatePlayground.Status.Conditions) != 0 {
								return fmt.Errorf("playground status conditions should be empty, got %v", updatePlayground.Status.Conditions)
							}
							return nil
						}, 3*util.Interval, util.Interval).Should(gomega.Succeed())
					},
				},
				{
					// create the model after playground is created.
					playgroundUpdateFn: func(_ *inferenceapi.Playground) {
						model = util.MockASampleModel()
						gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
					},
					checkPlayground: func(ctx context.Context, k8sClient client.Client, playground *inferenceapi.Playground) {
						validation.ValidatePlayground(ctx, k8sClient, playground)
						validation.ValidatePlaygroundStatusEqualTo(ctx, k8sClient, playground, inferenceapi.PlaygroundProgressing, "Pending", metav1.ConditionTrue)
					},
				},
			},
		}),
	)
})
