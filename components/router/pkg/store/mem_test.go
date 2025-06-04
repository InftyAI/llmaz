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

package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMemoryStore(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()

	// Test Insert
	err := store.Insert(ctx, "pod0-0", "model0", Indicator{RunningQueueSize: 100, WaitingQueueSize: 200})
	assert.NoError(t, err)

	if store.Len() != 1 {
		t.Fatalf("Expected length to be 1, got %d", store.Len())
	}

	_, err = store.Get(ctx, "pod0-0", "model0")
	assert.NoError(t, err)

	err = store.Insert(ctx, "pod1-0", "model1", Indicator{RunningQueueSize: 100, WaitingQueueSize: 200})
	assert.NoError(t, err)

	if store.Len() != 2 {
		t.Fatalf("Expected length to be 2, got %d", store.Len())
	}

	// Test Remove
	err = store.Remove(ctx, "pod0-0", "model0")
	assert.NoError(t, err)

	err = store.Remove(ctx, "pod1-0", "model1")
	assert.NoError(t, err)

	// Test Length
	if store.Len() != 0 {
		t.Fatalf("Expected length to be 0, got %d", store.Len())
	}

	_, err = store.Get(ctx, "pod0-0", "model0")
	assert.Error(t, err)
}
