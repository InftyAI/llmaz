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

package framework

import (
	"context"

	"github.com/inftyai/router/pkg/store"
)

const (
	MaxScore = 100
	MinScore = 0

	SuccessStatus       StatusCode = "Success"
	UnschedulableStatus StatusCode = "Unschedulable"
)

const (
	NoneCandidate string = ""
)

type Status struct {
	Code StatusCode
}

type StatusCode string

// Framework represents the algo about how to pick the candidates among all the peers.
type Framework interface {
	// RegisterPlugins will register the plugins to run.
	RegisterPlugins([]RegisterFunc) error
	// RunFilterPlugins will filter out unsatisfied peers.
	RunFilterPlugins(ctx context.Context, modelName string, store *store.DataStore) []string
	// RunScorePlugins will calculate the scores of all the peers.
	RunScorePlugins(ctx context.Context, candidates []string, modelName string, store *store.DataStore) string
}

// Plugin is the parent type for all the framework plugins.
// the same time.
type Plugin interface {
	Name() string
}

type FilterPlugin interface {
	Plugin
	// Filter helps to filter out unrelated peers.
	Filter(context.Context, store.Indicator) Status
}

type ScorePlugin interface {
	Plugin
	Score(context.Context, *store.DataStore, store.Indicator) float32
	// TODO: Weight should be configurable via yaml files.
	Weight() int
}
