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
	"errors"
	"fmt"
	"strings"

	dto "github.com/prometheus/client_model/go"

	"github.com/inftyai/metrics-aggregator/pkg/store"
	"github.com/inftyai/metrics-aggregator/pkg/util"
)

const (
	// We assumed that all the backends are using the same endpoint for metrics.
	METRICS_ENDPOINT = "/metrics"
)

type Backend interface {
	// ParseMetrics parses the metrics from the given metric family map.
	ParseMetrics(name string, metrics map[string]*dto.MetricFamily) (store.Indicator, error)

	// metricsMap returns the map of the metricType to the real metric name.
	// The whole list of metricTypes are defined in the store package.
	// TODO: in the future, we should make this configurable, so people can quick add a new backend
	// rather than modifying the code here.
	metricsMap() map[store.MetricType]string
}

func QueryMetrics(name string, endpoint string) (store.Indicator, error) {
	if endpoint[len(endpoint)-1] == '/' {
		endpoint = endpoint[:len(endpoint)-1]
	}
	url := endpoint + METRICS_ENDPOINT

	mfs, err := util.RequestMetrics(url)
	if err != nil {
		return store.Indicator{}, err
	}

	if len(mfs) == 0 {
		return store.Indicator{}, fmt.Errorf("no metrics found at %s", url)
	}

	backend, err := detectBackend(mfs)
	if err != nil {
		return store.Indicator{}, err
	}

	return backend.ParseMetrics(name, mfs)
}

func detectBackend(mfs map[string]*dto.MetricFamily) (Backend, error) {
	for name := range mfs {
		if strings.HasPrefix(name, "llamacpp") {
			return &LlamaCpp{}, nil
		}

		// detect the first item is enough.
		break
	}
	// TODO: add more backends here.
	return nil, errors.New("unsupported backend")
}
