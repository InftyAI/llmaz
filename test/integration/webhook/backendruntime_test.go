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

	"github.com/inftyai/llmaz/test/util"
)

var _ = ginkgo.Describe("BackendRuntime default and validation", func() {
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
		ginkgo.Entry("BackendRuntime creation with no image", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Image("").Obj()
				return k8sClient.Create(ctx, runtime)
			},
			failed: true,
		}),
		ginkgo.Entry("BackendRuntime creation with limits less than requests", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Limit("cpu", "1").Obj()
				return k8sClient.Create(ctx, runtime)
			},
			failed: true,
		}),
		ginkgo.Entry("BackendRuntime creation with unsupported inferenceMode", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Arg("unknown", []string{"foo", "bar"}).Obj()
				return k8sClient.Create(ctx, runtime)
			},
			failed: true,
		}),
		ginkgo.Entry("BackendRuntime creation with duplicated inferenceMode", &testValidatingCase{
			creationFunc: func() error {
				runtime := util.MockASampleBackendRuntime().Obj()
				if err := k8sClient.Create(ctx, runtime); err != nil {
					return err
				}
				anotherRuntime := util.MockASampleBackendRuntime().Name("another-vllm").Obj()
				if err := k8sClient.Create(ctx, anotherRuntime); err != nil {
					return err
				}
				return nil
			},
			failed: true,
		}),
	)
})
