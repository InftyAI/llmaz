/*
Copyright 2024.

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
	"strings"

	corev1 "k8s.io/api/core/v1"
)

// MergeResources will merge resources in toBeMerged to toMerge.
// If resources exist in toMerge, do nothing. If not exist, will
// set the value in toBeMerged to toMerge.
func MergeResources(toMerge corev1.ResourceList, toBeMerged corev1.ResourceList) corev1.ResourceList {
	if toMerge == nil {
		toMerge = corev1.ResourceList{}
	}
	for k, v := range toBeMerged {
		if _, exist := toMerge[k]; !exist {
			toMerge[k] = v
		}
	}
	return toMerge
}

// MergeKVs will merge kvs in toBeMerged to toMerge.
// If kvs exist in toMerge, they will be replaced.
func MergeKVs(toMerge map[string]string, toBeMerged map[string]string) map[string]string {
	for k, v := range toBeMerged {
		if toMerge == nil {
			toMerge = map[string]string{}
		}
		toMerge[k] = v
	}
	return toMerge
}

// In checks if a string is in a slice.
func In(strings []string, s string) bool {
	for _, str := range strings {
		if str == s {
			return true
		}
	}
	return false
}

// MergeArgsWithCommands will merge the arguments with the commands, the arguments will be append
// sequentially and separated with spaces. If the last command has "/n", it will be trimmed.
// has "/n", it will be trimmed.
func MergeArgsWithCommands(commands []string, args []string) []string {
	fullArgs := strings.Join(args, " ")
	commands[len(commands)-1] = fmt.Sprintf("%s %s", strings.TrimSuffix(commands[len(commands)-1], "\n"), fullArgs)
	return commands
}
