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

// import (
// 	"github.com/onsi/ginkgo/v2"
// 	"github.com/onsi/gomega"
// 	corev1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// var _ = ginkgo.Describe("model default and validation", func() {
// 	// Each test runs in a separate namespace.
// 	var ns *corev1.Namespace

// 	ginkgo.BeforeEach(func() {
// 		// Create test namespace before each test.
// 		ns = &corev1.Namespace{
// 			ObjectMeta: metav1.ObjectMeta{
// 				GenerateName: "test-ns-",
// 			},
// 		}
// 		gomega.Expect(k8sClient.Create(ctx, ns)).To(gomega.Succeed())
// 	})
// 	ginkgo.AfterEach(func() {
// 		gomega.Expect(k8sClient.Delete(ctx, ns)).To(gomega.Succeed())
// 	})

// 	ginkgo.DescribeTable("test defaulting",
// 		func() {

// 		},
// 	)
// })
