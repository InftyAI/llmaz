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

package util

import (
	"fmt"

	dto "github.com/prometheus/client_model/go"
)

// ParseMetricsWithNoLabel parses the metrics from the given metric family map with no labels specified.
func ParseMetricsWithNoLabel(metricName string, mfs map[string]*dto.MetricFamily) (float64, error) {
	mf, ok := mfs[metricName]
	if !ok {
		return 0, fmt.Errorf("metric %s not found", metricName)
	}

	values := mf.GetMetric()

	if len(values) != 1 {
		return 0, fmt.Errorf("metric %s has multiple values", metricName)
	}

	switch mf.GetType() {
	case dto.MetricType_COUNTER:
		return values[0].GetCounter().GetValue(), nil
	case dto.MetricType_GAUGE:
		return values[0].GetGauge().GetValue(), nil
	default:
		return 0, fmt.Errorf("unsupported metric type %s", mf.GetType())
	}
}
