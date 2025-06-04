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

package dispatcher

import (
	"context"

	"github.com/inftyai/metrics-aggregator/pkg/dispatcher/framework"
	"github.com/inftyai/metrics-aggregator/pkg/store"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var _ framework.Framework = &Dispatcher{}

type Dispatcher struct {
	registry      framework.Registry
	filterPlugins []framework.FilterPlugin
	scorePlugins  []framework.ScorePlugin
}

func NewDispatcher(plugins ...framework.RegisterFunc) *Dispatcher {
	dispatcher := &Dispatcher{}
	dispatcher.RegisterPlugins(plugins)
	return dispatcher
}

// RegisterFunc is a function that registers plugins.
func (d *Dispatcher) RegisterPlugins(fns []framework.RegisterFunc) error {
	if d.registry == nil {
		d.registry = make(framework.Registry)
	}

	for _, fn := range fns {
		if err := d.registry.Register(fn); err != nil {
			return err
		}
	}

	d.filterPlugins = d.registry.FilterPlugins()
	d.scorePlugins = d.registry.ScorePlugins()

	return nil
}

func (d *Dispatcher) RunFilterPlugins(ctx context.Context, modelName string, dataStore *store.DataStore) []string {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := log.FromContext(ctx)

	candidates := dataStore.FilterIterate(ctx, func(ctx context.Context, indicator store.Indicator) bool {
		for _, plugin := range d.filterPlugins {
			status := plugin.Filter(ctx, indicator)
			if status.Code != framework.SuccessStatus {
				logger.Info("filtering out candidate", "name", indicator.Name, "status", status.Code)
				return false
			}
		}
		return true
	})

	return candidates
}

func (d *Dispatcher) RunScorePlugins(ctx context.Context, candidates []string, modelName string, dataStore *store.DataStore) string {
	if len(candidates) == 0 {
		return framework.NoneCandidate
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	logger := log.FromContext(ctx)

	candidate := dataStore.ScoreIterate(ctx, func(ctx context.Context, indicator store.Indicator) float32 {
		totalScore := float32(0)

		for _, plugin := range d.scorePlugins {
			score := plugin.Score(ctx, dataStore, indicator)
			logger.V(10).Info("scored candidate", "name", indicator.Name, "score", score, "plugin", plugin.Name())
			totalScore += standardizeScore(score * float32(plugin.Weight()))
		}

		logger.V(6).Info("total score for candidate", "name", indicator.Name, "totalScore", totalScore, "modelName", modelName)
		return totalScore
	})

	return candidate
}

// To avoid one plugin returns a score that is too low or too high.
func standardizeScore(score float32) float32 {
	if score < framework.MinScore {
		return framework.MinScore
	}
	if score > framework.MaxScore {
		return framework.MaxScore
	}
	return score
}
