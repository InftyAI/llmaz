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
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"

	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
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
		creationFunc func() error
		failed       bool
	}
	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			if tc.failed {
				gomega.Expect(tc.creationFunc()).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(tc.creationFunc()).To(gomega.Succeed())
			}
		},
		ginkgo.Entry("normal BackendRuntime creation", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Obj()
				return k8sClient.Create(ctx, runtime)
			},
			failed: false,
		}),
		ginkgo.Entry("BackendRuntime creation with limits less than requests", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Limit("cpu", "1").Obj()
				return k8sClient.Create(ctx, runtime)
			},
			failed: true,
		}),
		ginkgo.Entry("BackendRuntime creation with unknown argument name", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Arg("unknown", []string{"foo", "bar"}).Obj()
				return k8sClient.Create(ctx, runtime)
			},
			failed: false,
		}),
		ginkgo.Entry("BackendRuntime creation with duplicated argument name", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Arg("default", []string{"foo", "bar"}).Obj()
				return k8sClient.Create(ctx, runtime)
			},
			failed: true,
		}),
	)
})
