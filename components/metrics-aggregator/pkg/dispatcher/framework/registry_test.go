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

package framework

import (
	"context"
	"testing"

	"github.com/inftyai/metrics-aggregator/pkg/store"
	"github.com/stretchr/testify/assert"
)

type MockPlugin struct {
	name string
}

func (m *MockPlugin) Name() string {
	return m.name
}

func (m *MockPlugin) Filter(context.Context, store.Indicator) Status {
	return Status{Code: SuccessStatus}
}

func TestRegistry(t *testing.T) {
	registry := make(Registry)

	// Test Register
	plugin1 := &MockPlugin{name: "plugin1"}
	err := registry.Register(func() (Plugin, error) {
		return plugin1, nil
	})
	assert.NoError(t, err)
	assert.Contains(t, registry, "plugin1")

	// Test Register with duplicate name
	err = registry.Register(func() (Plugin, error) {
		return &MockPlugin{name: "plugin1"}, nil
	})
	assert.Error(t, err)

	assert.Equal(t, len(registry.FilterPlugins()), 1)
	assert.Equal(t, len(registry.ScorePlugins()), 0)

	// Test Unregister
	err = registry.Unregister("plugin1")
	assert.NoError(t, err)
	assert.NotContains(t, registry, "plugin1")

	// Test Unregister non-existing plugin
	err = registry.Unregister("non-existing")
	assert.Error(t, err)
}
