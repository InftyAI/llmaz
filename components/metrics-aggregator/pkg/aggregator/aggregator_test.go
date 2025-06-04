/*
Copyright 2025 The InftyAI Team.

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

package aggregator

import (
	"context"
	"testing"
	"time"

	"github.com/inftyai/metrics-aggregator/pkg/store"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDefaultKeyFunc(t *testing.T) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
	}
	key := DefaultKeyFunc(pod)
	if key != "test-namespace/test-pod" {
		t.Fatal("key is not correct")
	}
}

func TestAggregator(t *testing.T) {
	store := store.NewMemoryStore()
	agg := NewAggregator(context.Background(), 500*time.Millisecond, store)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-namespace",
		},
	}

	agg.AddPod(pod)

	if agg.Len() != 1 {
		t.Fatal("pod count is not correct")
	}

	agg.AddPod(pod)

	if agg.Len() != 1 {
		t.Fatal("pod count is not correct")
	}

	if _, ok := agg.GetPod("test-namespace/test-pod"); !ok {
		t.Fatal("pod not found")
	}

	if _, ok := agg.GetPod("test-namespace/test-pod2"); ok {
		t.Fatal("pod should not be found")
	}

	agg.DeletePod("test-namespace/test-pod")

	if agg.Len() != 0 {
		t.Fatal("pod count is not correct")
	}

	if _, ok := agg.GetPod("test-namespace/test-pod"); ok {
		t.Fatal("pod should not be found")
	}

	agg.DeletePod("test-namespace/test-pod")
	if agg.Len() != 0 {
		t.Fatal("pod count is not correct")
	}
}
