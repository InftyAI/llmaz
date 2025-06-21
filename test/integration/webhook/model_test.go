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
	"time"

	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

var (
	testTime = metav1.Date(2025, 6, 20, 10, 0, 0, 0, time.UTC)
)

var _ = ginkgo.Describe("model default and validation", func() {

	// Delete all the Models for each case.
	ginkgo.AfterEach(func() {
		var models coreapi.OpenModelList
		gomega.Expect(k8sClient.List(ctx, &models)).To(gomega.Succeed())

		for _, model := range models.Items {
			gomega.Expect(k8sClient.Delete(ctx, &model)).To(gomega.Succeed())
		}
	})

	type testDefaultingCase struct {
		model     func() *coreapi.OpenModel
		wantModel func() *coreapi.OpenModel
	}
	ginkgo.DescribeTable("Defaulting test",
		func(tc *testDefaultingCase) {
			model := tc.model()
			gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
			gomega.Expect(model).To(gomega.BeComparableTo(tc.wantModel(),
				cmpopts.IgnoreTypes(coreapi.ModelStatus{}),
				cmpopts.IgnoreFields(coreapi.ModelSpec{}, "CreatedAt"),
				cmpopts.IgnoreFields(metav1.ObjectMeta{}, "UID", "ResourceVersion", "Generation", "CreationTimestamp", "ManagedFields")))
		},
		ginkgo.Entry("apply model family name", &testDefaultingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("meta-llama/Meta-Llama-3-8B", "", "", nil, nil).FamilyName("llama3").Obj()
			},
			wantModel: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("meta-llama/Meta-Llama-3-8B", "", "main", nil, nil).ModelSourceWithModelHub("Huggingface").FamilyName("llama3").Label(coreapi.ModelFamilyNameLabelKey, "llama3").OwnedBy(coreapi.DefaultOwnedBy).Obj()
			},
		}),
		ginkgo.Entry("apply modelscope model hub name", &testDefaultingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithModelHub("ModelScope").ModelSourceWithModelID("LLM-Research/Meta-Llama-3-8B", "", "", nil, nil).Obj()
			},
			wantModel: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("LLM-Research/Meta-Llama-3-8B", "", "main", nil, nil).ModelSourceWithModelHub("ModelScope").FamilyName("llama3").Label(coreapi.ModelFamilyNameLabelKey, "llama3").OwnedBy(coreapi.DefaultOwnedBy).Obj()
			},
		}),
		ginkgo.Entry("custom ownedBy should not be overwritten", &testDefaultingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithModelHub("ModelScope").ModelSourceWithModelID("LLM-Research/Meta-Llama-3-8B", "", "", nil, nil).OwnedBy("custom-owner").Obj()
			},
			wantModel: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("LLM-Research/Meta-Llama-3-8B", "", "main", nil, nil).ModelSourceWithModelHub("ModelScope").FamilyName("llama3").Label(coreapi.ModelFamilyNameLabelKey, "llama3").OwnedBy("custom-owner").Obj()
			},
		}),
		ginkgo.Entry("set custom createdAt", &testDefaultingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithModelHub("ModelScope").ModelSourceWithModelID("LLM-Research/Meta-Llama-3-8B", "", "", nil, nil).CreatedAt(testTime).Obj()
			},
			wantModel: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("LLM-Research/Meta-Llama-3-8B", "", "main", nil, nil).ModelSourceWithModelHub("ModelScope").FamilyName("llama3").Label(coreapi.ModelFamilyNameLabelKey, "llama3").OwnedBy(coreapi.DefaultOwnedBy).CreatedAt(testTime).Obj()
			},
		}),
	)

	type testValidatingCase struct {
		model  func() *coreapi.OpenModel
		failed bool
	}
	// TODO: add more testCases to cover update.
	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			if tc.failed {
				gomega.Expect(k8sClient.Create(ctx, tc.model())).Should(gomega.HaveOccurred())
			} else {
				gomega.Expect(k8sClient.Create(ctx, tc.model())).To(gomega.Succeed())
			}
		},
		ginkgo.Entry("default normal huggingface model creation", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithModelID("meta-llama/Meta-Llama-3-8B", "", "", nil, nil).Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("normal modelScope model creation", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithModelHub("ModelScope").ModelSourceWithModelID("LLM-Research/Meta-Llama-3-8B", "", "", nil, nil).Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("invalid model name", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("qwen-2-0.5b").FamilyName("qwen2").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct", "", "", nil, nil).Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("model creation with URI configured", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithURI("oss://bucket.endpoint/models/meta-llama-3-8B").Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("model creation with host protocol", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithURI("host:///models/meta-llama-3-8B").Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("model creation with protocol unknown URI", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithURI("unknown://bucket.endpoint/models/meta-llama-3-8B").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("model creation with no bucket URI", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").ModelSourceWithURI("oss://endpoint/models/meta-llama-3-8B").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("unknown modelHub", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelHub("unknown").FamilyName("llama3").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("no data source configured", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").FamilyName("llama3").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("set filename when modelHub is Huggingface", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", nil, nil).FamilyName("llama3").Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("set filename and allowPatterns when modelHub is Huggingface", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", []string{"*"}, nil).FamilyName("llama3").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("set filename and ignorePatterns when modelHub is Huggingface", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", nil, []string{"*"}).FamilyName("llama3").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("set allowPatterns and ignorePatterns when modelHub is Huggingface", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "", "", []string{"*"}, []string{"*.gguf"}).FamilyName("llama3").Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("set filename when modelHub is ModelScope", &testValidatingCase{
			model: func() *coreapi.OpenModel {
				return wrapper.MakeModel("llama3-8b").ModelSourceWithModelHub("ModelScope").ModelSourceWithModelID("Qwen/Qwen2-0.5B-Instruct-GGUF", "qwen2-0_5b-instruct-q5_k_m.gguf", "", nil, nil).FamilyName("llama3").Obj()
			},
			failed: true,
		}),
	)
})
