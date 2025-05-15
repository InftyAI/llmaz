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

package indicator

import "github.com/inftyai/metrics-aggregator/pkg/util"

var _ Indicator = &LlamaCpp{}

type LlamaCpp struct {
}

func (l *LlamaCpp) Endpoint() string {
	return "/metrics"
}

func (l *LlamaCpp) MetricsMap() map[IndicatorType]string {
	return map[IndicatorType]string{
		RunningQueueSize: "llamacpp:requests_processing",
		WaitingQueueSize: "llamacpp:requests_deferred",
	}
}

// TODO: add tests, mock the API request here.
func (l *LlamaCpp) QueryMetrics(uri string) (MetricValues, error) {
	if uri[len(uri)-1] == '/' {
		uri = uri[:len(uri)-1]
	}
	url := uri + l.Endpoint()

	metrics, err := util.RequestMetric(url)
	if err != nil {
		return nil, err
	}

	res := make(MetricValues)

	for k, v := range l.MetricsMap() {
		value, err := util.ParseMetricsWithNoLabel(v, metrics)
		if err != nil {
			return nil, err
		}
		res[k] = value
	}
	return res, nil
}
