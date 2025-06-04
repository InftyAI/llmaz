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

package backend

import (
	dto "github.com/prometheus/client_model/go"

	"github.com/inftyai/metrics-aggregator/pkg/store"
	"github.com/inftyai/metrics-aggregator/pkg/util"
)

var _ Backend = &LlamaCpp{}

type LlamaCpp struct{}

func (l *LlamaCpp) ParseMetrics(name string, metrics map[string]*dto.MetricFamily) (store.Indicator, error) {
	res := make(store.MetricValues)

	for k, v := range l.metricsMap() {
		value, err := util.ParseMetricsWithNoLabel(v, metrics)
		if err != nil {
			return store.Indicator{}, err
		}
		res[k] = value
	}
	return store.MapToInstanceMetrics(name, res), nil
}

func (l *LlamaCpp) metricsMap() map[store.MetricType]string {
	return map[store.MetricType]string{
		store.RunningQueueSize: "llamacpp:requests_processing",
		store.WaitingQueueSize: "llamacpp:requests_deferred",
		store.KVCacheUsage:     "llamacpp:kv_cache_usage_ratio",
	}
}
