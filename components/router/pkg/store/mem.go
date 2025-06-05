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
	"fmt"
	"sync"
)

var _ Store = &MemoryStore{}

type MemoryStore struct {
	mu   sync.RWMutex
	data map[string]*DataStore // Key: modelName, Value: *podWrapperStore
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		data: make(map[string]*DataStore),
	}
}

func (m *MemoryStore) Insert(ctx context.Context, podWrapperName string, modelName string, metrics Indicator) error {
	m.mu.Lock()
	store := m.data[modelName]

	if store == nil {
		store = &DataStore{
			data: make(map[string]Indicator),
		}
		m.data[modelName] = store
	}
	m.mu.Unlock()

	store.mu.Lock()
	// always refresh the metrics.
	store.data[podWrapperName] = metrics

	// Insert the min/max value of all kinds of metrics, they will be used in Score plugins.
	store.RunningQueueSize[0] = min(store.RunningQueueSize[0], metrics.RunningQueueSize)
	store.RunningQueueSize[1] = max(store.RunningQueueSize[1], metrics.RunningQueueSize)
	store.WaitingQueueSize[0] = min(store.WaitingQueueSize[0], metrics.WaitingQueueSize)
	store.WaitingQueueSize[1] = max(store.WaitingQueueSize[1], metrics.WaitingQueueSize)
	store.KVCacheUsage[0] = min(store.KVCacheUsage[0], metrics.KVCacheUsage)
	store.KVCacheUsage[1] = max(store.KVCacheUsage[1], metrics.KVCacheUsage)

	store.mu.Unlock()

	return nil
}

func (m *MemoryStore) Remove(ctx context.Context, podWrapperName string, modelName string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	store := m.data[modelName]

	if store == nil {
		return nil
	}

	store.mu.Lock()
	delete(store.data, podWrapperName)
	store.mu.Unlock()

	if len(store.data) == 0 {
		delete(m.data, modelName)
	}
	return nil
}

func (m *MemoryStore) Len() int32 {
	count := 0
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, store := range m.data {
		store.mu.RLock()
		count += len(store.data)
		store.mu.RUnlock()
	}

	return int32(count)
}

func (m *MemoryStore) Get(ctx context.Context, podWrapperName string, modelName string) (Indicator, error) {
	m.mu.RLock()
	store := m.data[modelName]
	if store == nil {
		m.mu.RUnlock()
		return Indicator{}, fmt.Errorf("model %s not found", modelName)
	}
	m.mu.RUnlock()

	store.mu.RLock()
	storeMetrics, exists := store.data[podWrapperName]
	if !exists {
		store.mu.RUnlock()
		return Indicator{}, fmt.Errorf("pod wrapper %s not found", podWrapperName)
	}
	store.mu.RUnlock()
	return storeMetrics, nil
}

func (m *MemoryStore) GetDataStore(ctx context.Context, modelName string) (*DataStore, error) {
	m.mu.RLock()
	store, exists := m.data[modelName]
	m.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("podWrapperStore with name %s not found", modelName)
	}

	return store, nil
}
