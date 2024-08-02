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
	"errors"
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

func ParseURI(uri string) (format string, url string, err error) {
	parsers := strings.Split(uri, "://")
	if len(parsers) != 2 {
		return "", "", errors.New("uri format error")
	}
	return parsers[0], parsers[1], nil
}
