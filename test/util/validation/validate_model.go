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

package validation

import (
	"context"
	"errors"

	"github.com/onsi/gomega"
	coreapi "inftyai.com/llmaz/api/core/v1alpha1"
	"inftyai.com/llmaz/test/util"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateModel(ctx context.Context, k8sClient client.Client, model *coreapi.OpenModel) {
	gomega.Eventually(func() error {
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: model.Name}, model); err != nil {
			return errors.New("failed to get model")
		}

		if model.Labels[coreapi.ModelFamilyNameLabelKey] != string(model.Spec.FamilyName) {
			return errors.New("family name not right")
		}

		return nil
	}, util.IntegrationTimeout, util.Interval).Should(gomega.Succeed())
}
