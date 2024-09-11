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
	"context"
	"os"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/scheme"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	lws "sigs.k8s.io/lws/api/leaderworkerset/v1"

	coreapi "github.com/inftyai/llmaz/api/core/v1alpha1"
	inferenceapi "github.com/inftyai/llmaz/api/inference/v1alpha1"
	"github.com/inftyai/llmaz/test/util"
)

const (
	timeout  = 60 * time.Second
	interval = time.Millisecond * 250
)

var cfg *rest.Config
var k8sClient client.Client
var ctx context.Context
var cancel context.CancelFunc

func TestE2E(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "E2E Suite")
}

var _ = BeforeSuite(func() {
	log.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))
	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")

	// cfg is defined in this file globally.
	cfg = config.GetConfigOrDie()
	Expect(cfg).NotTo(BeNil())

	err := coreapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = inferenceapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = admissionv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = corev1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())
	err = lws.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	readyForTesting(k8sClient)
	Expect(os.Setenv("TEST_TYPE", "E2E")).Should(Succeed())

	Expect(util.Setup(ctx, k8sClient, "../config/backends")).To(Succeed())
})

var _ = AfterSuite(func() {
	cancel()
})

func readyForTesting(client client.Client) {
	By("waiting for webhooks to ready")

	// To verify that webhooks are ready, let's create a simple model.
	model := util.MockASampleModel()

	// Once the creation succeeds, that means the webhooks are ready
	// and we can begin testing.
	Eventually(func() error {
		return client.Create(ctx, model)
	}, timeout, interval).Should(Succeed())

	// Delete this model before beginning tests.
	Expect(client.Delete(ctx, model))
	Eventually(func() error {
		return client.Get(ctx, types.NamespacedName{Name: model.Name, Namespace: model.Namespace}, &coreapi.OpenModel{})
	}).ShouldNot(Succeed())
}
