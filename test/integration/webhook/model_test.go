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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	core "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("model default and validation", func() {

	// Delete all the Models for each case.
	ginkgo.AfterEach(func() {
		var models core.ModelList
		gomega.Expect(k8sClient.List(ctx, &models)).To(gomega.Succeed())

		for _, model := range models.Items {
			gomega.Expect(k8sClient.Delete(ctx, &model)).To(gomega.Succeed())
		}
	})

	type testDefaultingCase struct {
		model     func() *core.Model
		wantModel func() *core.Model
	}
	ginkgo.DescribeTable("Defaulting test",
		func(tc *testDefaultingCase) {
			model := tc.model()
			gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
			gomega.Expect(model).To(gomega.BeComparableTo(tc.wantModel(),
				cmpopts.IgnoreTypes(core.ModelStatus{}),
				cmpopts.IgnoreFields(metav1.ObjectMeta{}, "UID", "ResourceVersion", "Generation", "CreationTimestamp", "ManagedFields")))
		},
		ginkgo.Entry("apply model family name", &testDefaultingCase{
			model: func() *core.Model {
				return wrapper.MakeModel("llama3-8b").DataSourceWithModelID("meta-llama/Meta-Llama-3-8B").FamilyName("llama3").Obj()
			},
			wantModel: func() *core.Model {
				return wrapper.MakeModel("llama3-8b").DataSourceWithModelID("meta-llama/Meta-Llama-3-8B").DataSourceWithModelHub("Huggingface").FamilyName("llama3").Label(core.ModelFamilyNameLabelKey, "llama3").Obj()
			},
		}),
	)

	type testValidatingCase struct {
		model  func() *core.Model
		failed bool
	}
	// TODO: add more testCases to cover update.
	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			if tc.failed {
				gomega.Expect(k8sClient.Create(ctx, tc.model())).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(k8sClient.Create(ctx, tc.model())).To(gomega.Succeed())
			}
		},
		ginkgo.Entry("normal model creation", &testValidatingCase{
			model: func() *core.Model {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").DataSourceWithModelID("meta-llama/Meta-Llama-3-8B").Obj()
			},
			failed: false,
		}),
		// ginkgo.Entry("model creation with URI configured", &testValidatingCase{
		// 	model: func() *core.Model {
		// 		return wrapper.MakeModel("llama3-8b").FamilyName("llama3").DataSourceWithURI("image://meta-llama-3-8B").Obj()
		// 	},
		// 	failed: false,
		// }),
		ginkgo.Entry("no data source configured", &testValidatingCase{
			model: func() *core.Model {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").Obj()
			},
			failed: true,
		}),
	)
})
