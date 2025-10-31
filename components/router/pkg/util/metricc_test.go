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
	"strings"
	"testing"

	"github.com/prometheus/common/expfmt"
)

func TestParseMetricsWithNoLabel(t *testing.T) {
	metricsText := `
# HELP llamacpp:n_busy_slots_per_decode Average number of busy slots per llama_decode() call
# TYPE llamacpp:n_busy_slots_per_decode counter
llamacpp:n_busy_slots_per_decode 1
# HELP llamacpp:requests_processing Number of requests processing.
# TYPE llamacpp:requests_processing gauge
llamacpp:requests_processing 20
# HELP llamacpp:requests_deferred Number of requests deferred.
# TYPE llamacpp:requests_deferred gauge
llamacpp:requests_deferred 0
`

	parser := expfmt.TextParser{}
	mfs, err := parser.TextToMetricFamilies(strings.NewReader(metricsText))
	if err != nil {
		t.Fatalf("Failed to parse metrics: %v", err)
	}

	tests := []struct {
		metricName string
		want       float64
		err        bool
	}{
		{"llamacpp:n_busy_slots_per_decode", 1, false},
		{"llamacpp:requests_processing", 20, false},
		{"llamacpp:requests_deferred", 0, false},
		{"llamacpp:kv_cache_usage_ratio", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.metricName, func(t *testing.T) {
			got, err := ParseMetricsWithNoLabel(tt.metricName, mfs)
			if tt.err && err == nil || !tt.err && err != nil {
				t.Fatal("unexpected error")
			}

			if got != tt.want {
				t.Errorf("got = %v, want = %v", got, tt.want)
			}
		})
	}

}
