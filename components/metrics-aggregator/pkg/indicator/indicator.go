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

type IndicatorType string

const (
	RunningQueueSize IndicatorType = "running_queue_size"
	WaitingQueueSize IndicatorType = "waiting_queue_size"
)

// A collection of the metrics, indicatorType as the key.
type MetricValues map[IndicatorType]float64

type Indicator interface {
	// Endpoint represents the endpoint to get the metrics, e.g. /metrics.
	Endpoint() string
	// MetricsMap returns the map of the indicatorType to the real metric name.
	MetricsMap() map[IndicatorType]string
	// QueryMetrics returns the metrics from the url.
	QueryMetrics(uri string) (MetricValues, error)
}
