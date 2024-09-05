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

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
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

var _ = ginkgo.Describe("inferenceService controller test", func() {
	// Each test runs in a separate namespace.
	var ns *corev1.Namespace

	type update struct {
		updateFunc func(*inferenceapi.Service)
		checkFunc  func(context.Context, client.Client, *inferenceapi.Service)
	}

	ginkgo.BeforeEach(func() {
		// Create test namespace before each test.
		ns = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "ns-service-",
			},
		}
		gomega.Expect(k8sClient.Create(ctx, ns)).To(gomega.Succeed())
		model := util.MockASampleModel()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		modelWithURI := wrapper.MakeModel("model-with-uri").FamilyName("llama3").ModelSourceWithURI("oss://bucket.endpoint/modelPath").Obj()
		gomega.Expect(k8sClient.Create(ctx, modelWithURI)).To(gomega.Succeed())
	})
	ginkgo.AfterEach(func() {
		gomega.Expect(k8sClient.Delete(ctx, ns)).To(gomega.Succeed())
		var models coreapi.OpenModelList
		gomega.Expect(k8sClient.List(ctx, &models)).To(gomega.Succeed())
		for i := range models.Items {
			gomega.Expect(k8sClient.Delete(ctx, &models.Items[i])).To(gomega.Succeed())
		}
	})

	// TODO: Add more testCases to cover status update.

	type testValidatingCase struct {
		makeService func() *inferenceapi.Service
		updates     []*update
	}
	ginkgo.DescribeTable("test playground creation and update",
		func(tc *testValidatingCase) {
			service := tc.makeService()
			for _, update := range tc.updates {
				if update.updateFunc != nil {
					update.updateFunc(service)
				}
				newService := &inferenceapi.Service{}
				gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, newService)).To(gomega.Succeed())
				if update.checkFunc != nil {
					update.checkFunc(ctx, k8sClient, newService)
				}
			}
		},
		ginkgo.Entry("normal service create and update", &testValidatingCase{
			makeService: func() *inferenceapi.Service {
				return util.MockASampleService(ns.Name)
			},
			updates: []*update{
				{
					updateFunc: func(service *inferenceapi.Service) {
						gomega.Expect(k8sClient.Create(ctx, service)).To(gomega.Succeed())
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceProgressing, "ServiceInProgress", metav1.ConditionTrue)
					},
				},
				{
					updateFunc: func(service *inferenceapi.Service) {
						gomega.Eventually(func() error {
							updateService := &inferenceapi.Service{}
							if err := k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, updateService); err != nil {
								return err
							}
							updateService.Spec.WorkloadTemplate.Replicas = ptr.To[int32](3)
							if err := k8sClient.Update(ctx, updateService); err != nil {
								return err
							}
							return nil
						}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())

						// To make sure playground updated successfully.
						newService := inferenceapi.Service{}
						gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, &newService)).To(gomega.Succeed())
						gomega.Expect(*newService.Spec.WorkloadTemplate.Replicas).To(gomega.Equal(int32(3)))
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceProgressing, "ServiceInProgress", metav1.ConditionTrue)
					},
				},
				{
					updateFunc: func(service *inferenceapi.Service) {
						util.UpdateLwsToReady(ctx, k8sClient, service.Name, service.Namespace)
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
					},
				},
				{
					updateFunc: func(service *inferenceapi.Service) {
						util.UpdateLwsToUnReady(ctx, k8sClient, service.Name, service.Namespace)
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceProgressing, "ServiceInProgress", metav1.ConditionTrue)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceNotReady", metav1.ConditionFalse)
					},
				},
				{
					updateFunc: func(service *inferenceapi.Service) {
						util.UpdateLwsToReady(ctx, k8sClient, service.Name, service.Namespace)
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceProgressing, "ServiceInProgress", metav1.ConditionTrue)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
					},
				},
			},
		}),
		ginkgo.Entry("service created with URI configured Model", &testValidatingCase{
			makeService: func() *inferenceapi.Service {
				return wrapper.MakeService("service-llama3-8b", ns.Name).
					ModelClaims([]string{"model-with-uri"}, []string{"main"}).
					WorkerTemplate().
					Obj()
			},
			updates: []*update{
				{
					updateFunc: func(service *inferenceapi.Service) {
						gomega.Expect(k8sClient.Create(ctx, service)).To(gomega.Succeed())
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceProgressing, "ServiceInProgress", metav1.ConditionTrue)
					},
				},
				{
					updateFunc: func(service *inferenceapi.Service) {
						util.UpdateLwsToReady(ctx, k8sClient, service.Name, service.Namespace)
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
					},
				},
			},
		}),
		ginkgo.Entry("service created with speculativeDecoding mode", &testValidatingCase{
			makeService: func() *inferenceapi.Service {
				return wrapper.MakeService("service-llama3-8b", ns.Name).
					ModelClaims([]string{"llama3-8b", "model-with-uri"}, []string{"main", "draft"}).
					WorkerTemplate().
					Obj()
			},
			updates: []*update{
				{
					updateFunc: func(service *inferenceapi.Service) {
						gomega.Expect(k8sClient.Create(ctx, service)).To(gomega.Succeed())
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceProgressing, "ServiceInProgress", metav1.ConditionTrue)
					},
				},
				{
					updateFunc: func(service *inferenceapi.Service) {
						util.UpdateLwsToReady(ctx, k8sClient, service.Name, service.Namespace)
					},
					checkFunc: func(ctx context.Context, k8sClient client.Client, service *inferenceapi.Service) {
						validation.ValidateService(ctx, k8sClient, service)
						validation.ValidateServiceStatusEqualTo(ctx, k8sClient, service, inferenceapi.ServiceAvailable, "ServiceReady", metav1.ConditionTrue)
					},
				},
			},
		}),
	)
})
