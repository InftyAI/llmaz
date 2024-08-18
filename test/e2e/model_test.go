/*
Copyright 2024 The Kubernetes Authors.
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

package e2e

import (
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"inftyai.com/llmaz/test/util"
	"inftyai.com/llmaz/test/util/validation"
)

var _ = ginkgo.Describe("model e2e tests", func() {
	ginkgo.It("Can deploy a normal model", func() {
		model := util.MockASampleModel()
		gomega.Expect(k8sClient.Create(ctx, model)).To(gomega.Succeed())
		defer func() {
			gomega.Expect(k8sClient.Delete(ctx, model)).To(gomega.Succeed())
		}()

		validation.ValidateModel(ctx, k8sClient, model)
	})
})
