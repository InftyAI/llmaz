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

type MetricType string

// A collection of the metrics, indicatorType as the key.
type MetricValues map[MetricType]float64

const (
	// RunningQueueSize represents the number of requests currently running on GPU.
	RunningQueueSize MetricType = "running_queue_size"
	// WaitingQueueSize represents the number of requests waiting to be processed.
	WaitingQueueSize MetricType = "waiting_queue_size"
	// KVCacheUsage represents the kvcache usage, 1 means 100 percent usage.
	KVCacheUsage MetricType = "kv_cache_usage"
)

type Indicator struct {
	Name             string
	RunningQueueSize float64
	WaitingQueueSize float64
	KVCacheUsage     float64
}

func MapToInstanceMetrics(name string, m map[MetricType]float64) Indicator {
	return Indicator{
		Name:             name,
		RunningQueueSize: m[RunningQueueSize],
		WaitingQueueSize: m[WaitingQueueSize],
		KVCacheUsage:     m[KVCacheUsage],
	}
}
