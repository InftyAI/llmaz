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

package helper

import (
	"fmt"

	"gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
)

type GlobalConfigs struct {
	SchedulerName      string `yaml:"scheduler-name"`
	InitContainerImage string `yaml:"init-container-image"`
}

func ParseGlobalConfigmap(cm *corev1.ConfigMap) (*GlobalConfigs, error) {
	rawConfig, ok := cm.Data["config.data"]
	if !ok {
		return nil, fmt.Errorf("config.data not found in ConfigMap")
	}

	var configs GlobalConfigs
	err := yaml.Unmarshal([]byte(rawConfig), &configs)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal config.data: %v", err)
	}

	return &configs, nil
}
