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

type Store interface {
	Get(ctx context.Context, identifier string, modelName string) (Indicator, error)
	Insert(ctx context.Context, identifier string, modelName string, metrics Indicator) error
	Remove(ctx context.Context, identifier string, modelName string) error
	GetDataStore(ctx context.Context, modelName string) (*DataStore, error)

	// Should only used for testing.
	Len() int32
}

type DataStore struct {
	mu   sync.RWMutex
	data map[string]Indicator // Key: name, Value: Indicator

	// Keep track of the min/max values for each metric, 0-index is min and 1-index is max.
	// They will be used in score plugins.
	RunningQueueSize [2]float64
	WaitingQueueSize [2]float64
	KVCacheUsage     [2]float64
}

func (d *DataStore) Get(ctx context.Context, name string) (Indicator, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	metrics, exists := d.data[name]
	if !exists {
		return Indicator{}, fmt.Errorf("metrics for datastore %s not found", name)
	}
	return metrics, nil
}

// TODO: we should not iterate all the dataStore which may lead to performance issue.
func (d *DataStore) FilterIterate(ctx context.Context, fn func(context.Context, Indicator) bool) (names []string) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for name, indicator := range d.data {
		if fn(ctx, indicator) {
			names = append(names, name)
		}
	}
	return

}

func (d *DataStore) ScoreIterate(ctx context.Context, fn func(context.Context, Indicator) float32) string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	var highestScore float32
	var candidate string

	for name, indicator := range d.data {
		score := fn(ctx, indicator)
		// Iterate the d.data is already in random order, so we can just pick the first one with the highest score.
		if score > highestScore {
			highestScore = score
			candidate = name
		}
	}
	return candidate
}
