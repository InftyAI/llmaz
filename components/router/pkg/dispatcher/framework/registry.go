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
	"fmt"
)

type RegisterFunc = func() (Plugin, error)

// Registry is a collection of all available plugins. The framework uses a
// registry to enable and initialize configured plugins.
type Registry map[string]Plugin

// Register adds a new plugin to the registry. If a plugin with the same name
// exists, it returns an error.
func (r Registry) Register(fn RegisterFunc) error {
	plugin, err := fn()
	if err != nil {
		return err
	}

	name := plugin.Name()
	if _, ok := r[name]; ok {
		return fmt.Errorf("a plugin named %v already exists", name)
	}

	r[name] = plugin
	return nil
}

// Unregister removes an existing plugin from the registry. If no plugin with
// the provided name exists, it returns an error.
func (r Registry) Unregister(name string) error {
	if _, ok := r[name]; !ok {
		return fmt.Errorf("no plugin named %v exists", name)
	}
	delete(r, name)
	return nil
}

func (r Registry) FilterPlugins() (plugins []FilterPlugin) {
	for _, plugin := range r {
		if p, ok := plugin.(FilterPlugin); ok {
			plugins = append(plugins, p)
		}
	}
	return plugins
}

func (r Registry) ScorePlugins() (plugins []ScorePlugin) {
	for _, plugin := range r {
		if p, ok := plugin.(ScorePlugin); ok {
			plugins = append(plugins, p)
		}
	}
	return plugins
}
