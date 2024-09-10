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
	type testValidatingCase struct {
		backendRuntime func() *inferenceapi.BackendRuntime
		failed         bool
	}
	ginkgo.DescribeTable("test validating",
		func(tc *testValidatingCase) {
			if tc.failed {
				gomega.Expect(k8sClient.Create(ctx, tc.backendRuntime())).To(gomega.HaveOccurred())
			} else {
				gomega.Expect(k8sClient.Create(ctx, tc.backendRuntime())).To(gomega.Succeed())
			}
		},
		ginkgo.Entry("normal BackendRuntime creation", &testValidatingCase{
			backendRuntime: func() *inferenceapi.BackendRuntime {
				return util.MockASampleBackendRuntime().Obj()
			},
			failed: false,
		}),
		ginkgo.Entry("BackendRuntime creation with no image", &testValidatingCase{
			backendRuntime: func() *inferenceapi.BackendRuntime {
				return util.MockASampleBackendRuntime().Image("").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("BackendRuntime creation with limits less than requests", &testValidatingCase{
			backendRuntime: func() *inferenceapi.BackendRuntime {
				return util.MockASampleBackendRuntime().Limit("cpu", "1").Obj()
			},
			failed: true,
		}),
		ginkgo.Entry("BackendRuntime creation with unsupported inferenceOption", &testValidatingCase{
			backendRuntime: func() *inferenceapi.BackendRuntime {
				return util.MockASampleBackendRuntime().Arg("unknown", []string{"foo", "bar"}).Obj()
			},
			failed: true,
		}),
	)
})
