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
	"k8s.io/apimachinery/pkg/types"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
	"github.com/inftyai/llmaz/test/util/wrapper"
)

var _ = ginkgo.Describe("BackendRuntime default and validation", func() {

	// Delete all backendRuntimes for each case.
	ginkgo.AfterEach(func() {
		var runtimes inferenceapi.BackendRuntimeList
		gomega.Expect(k8sClient.List(ctx, &runtimes)).To(gomega.Succeed())

		for _, runtime := range runtimes.Items {
			gomega.Expect(k8sClient.Delete(ctx, &runtime)).To(gomega.Succeed())
		}
	})

	type testValidatingCase struct {
		creationFunc func() *inferenceapi.BackendRuntime
		createFailed bool
		updateFunc   func(*inferenceapi.BackendRuntime) *inferenceapi.BackendRuntime
		updateFiled  bool
	}
	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			backend := tc.creationFunc()
			err := k8sClient.Create(ctx, backend)

			if tc.createFailed {
				gomega.Expect(err).To(gomega.HaveOccurred())
				return
			} else {
				gomega.Expect(err).To(gomega.Succeed())
			}

			gomega.Expect(k8sClient.Get(ctx, types.NamespacedName{Name: backend.Name, Namespace: backend.Namespace}, backend)).Should(gomega.Succeed())

			if tc.updateFunc != nil {
				err = k8sClient.Update(ctx, tc.updateFunc(backend))
				if tc.updateFiled {
					gomega.Expect(err).To(gomega.HaveOccurred())
				} else {
					gomega.Expect(err).To(gomega.Succeed())
				}
			}
		},
		ginkgo.Entry("normal BackendRuntime creation", &testValidatingCase{
			creationFunc: func() *inferenceapi.BackendRuntime {
				return util.MockASampleBackendRuntime().Obj()
			},
			createFailed: false,
		}),
		ginkgo.Entry("BackendRuntime creation with limits less than requests", &testValidatingCase{
			creationFunc: func() *inferenceapi.BackendRuntime {
				return util.MockASampleBackendRuntime().Limit("default", "cpu", "1").Obj()
			},
			createFailed: true,
		}),
		ginkgo.Entry("BackendRuntime creation with unknown argument name", &testValidatingCase{
			creationFunc: func() *inferenceapi.BackendRuntime {
				return util.MockASampleBackendRuntime().Arg("unknown", []string{"foo", "bar"}).Obj()
			},
			createFailed: false,
		}),
		ginkgo.Entry("BackendRuntime creation with no resources", &testValidatingCase{
			creationFunc: func() *inferenceapi.BackendRuntime {
				return wrapper.MakeBackendRuntime("vllm").
					Image("vllm/vllm-openai").Version(util.VllmImageVersion).
					Command([]string{"python3", "-m", "vllm.entrypoints.openai.api_server"}).
					Arg("default", []string{"--model", "{{.ModelPath}}", "--served-model-name", "{{.ModelName}}", "--host", "0.0.0.0", "--port", "8080"}).
					Obj()
			},
			createFailed: false,
		}),
	)
})
