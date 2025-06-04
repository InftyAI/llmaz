/*
Copyright 2025.

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

package latencyAware

import (
	"context"

	"github.com/inftyai/router/pkg/dispatcher/framework"
	"github.com/inftyai/router/pkg/store"
)

var _ framework.ScorePlugin = &LatencyAware{}

type LatencyAware struct{}

func New() (framework.Plugin, error) {
	return &LatencyAware{}, nil
}

func (a *LatencyAware) Name() string {
	return "LatencyAware"
}

func (a *LatencyAware) Weight() int {
	return 1
}

// Apply with min-max normalization, the score are calculated as follows:
// 1. Running queue size is weighted by 0.3.
// 2. Waiting queue size is weighted by 0.3.
// 3. KV cache usage is weighted by 0.4.
// The higher the score, the better the performance.
// TODO: This is not the final algorithm.
func (a *LatencyAware) Score(ctx context.Context, dataStore *store.DataStore, indicator store.Indicator) float32 {
	runningQueueSizeMinMax := dataStore.RunningQueueSize[1] - dataStore.RunningQueueSize[0]
	waitingQueueSizeMinMax := dataStore.WaitingQueueSize[1] - dataStore.WaitingQueueSize[0]
	kvCacheUsageMinMax := dataStore.KVCacheUsage[1] - dataStore.KVCacheUsage[0]

	totalScore := float32(0)

	if runningQueueSizeMinMax != 0 {
		runningQueueSizeScore := 100 * (1 - float32((indicator.RunningQueueSize-dataStore.RunningQueueSize[0])/runningQueueSizeMinMax))
		totalScore += 0.3 * runningQueueSizeScore
	}
	if waitingQueueSizeMinMax != 0 {
		waitingQueueSizeScore := 100 * (1 - float32((indicator.WaitingQueueSize-dataStore.WaitingQueueSize[0])/waitingQueueSizeMinMax))
		totalScore += 0.3 * waitingQueueSizeScore
	}
	if kvCacheUsageMinMax != 0 {
		kvCacheUsageScore := 100 * (1 - float32((indicator.KVCacheUsage-dataStore.KVCacheUsage[0])/kvCacheUsageMinMax))
		totalScore += 0.4 * kvCacheUsageScore
	}

	return totalScore
}
