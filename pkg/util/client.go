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

package util

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	// "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	fieldManager = "llmaz"
)

func Patch(ctx context.Context, k8sClient client.Client, pointerObj interface{}) error {
	// TODO: once https://github.com/kubernetes/kubernetes/issues/67610 is fixed,
	// no need to folk the DefaultUnstructuredConverter.
	obj, err := DefaultUnstructuredConverter.ToUnstructured(pointerObj)
	if err != nil {
		return err
	}
	patch := &unstructured.Unstructured{
		Object: obj,
	}

	if err := k8sClient.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: fieldManager,
		Force:        ptr.To[bool](true),
	}); err != nil {
		return err
	}

	return nil
}
