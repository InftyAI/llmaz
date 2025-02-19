/*
Copyright 2024 The InftyAI Team.

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
	"testing"

	"github.com/google/go-cmp/cmp"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMergeResources(t *testing.T) {
	testCases := []struct {
		name          string
		toMerge       corev1.ResourceList
		toBeMerged    corev1.ResourceList
		wantResources corev1.ResourceList
	}{
		{
			name: "toBeMerged and toMerge has same CPU",
			toMerge: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("1"),
				corev1.ResourceMemory: resource.MustParse("2Gi"),
			},
			toBeMerged: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("2"),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
			},
			wantResources: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("1"),
				corev1.ResourceMemory:           resource.MustParse("2Gi"),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
			},
		},
		{
			name:    "toMerge is nil",
			toMerge: nil,
			toBeMerged: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("2"),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
			},
			wantResources: corev1.ResourceList{
				corev1.ResourceCPU:              resource.MustParse("2"),
				corev1.ResourceEphemeralStorage: resource.MustParse("100Gi"),
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeResources(tc.toMerge, tc.toBeMerged)
			if diff := cmp.Diff(got, tc.wantResources); diff != "" {
				t.Fatalf("unexpected resources: %s", diff)
			}
		})
	}
}

func TestMergeKVs(t *testing.T) {
	testCases := []struct {
		name       string
		toMerge    map[string]string
		toBeMerged map[string]string
		want       map[string]string
	}{
		{
			name:       "toBeMerged and toMerge has same key",
			toMerge:    map[string]string{"foo": "bar"},
			toBeMerged: map[string]string{"foo": "buz"},
			want:       map[string]string{"foo": "buz"},
		},
		{
			name:       "toMerge has exclusive keys",
			toMerge:    map[string]string{"foo": "bar"},
			toBeMerged: map[string]string{"fuz": "buz"},
			want:       map[string]string{"foo": "bar", "fuz": "buz"},
		},
		{
			name:       "toMerge is nil",
			toMerge:    nil,
			toBeMerged: map[string]string{"fuz": "buz"},
			want:       map[string]string{"fuz": "buz"},
		},
		{
			name:       "toBeMerge is nil",
			toMerge:    map[string]string{"fuz": "buz"},
			toBeMerged: nil,
			want:       map[string]string{"fuz": "buz"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := MergeKVs(tc.toMerge, tc.toBeMerged)
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Fatalf("unexpected kvs: %s", diff)
			}
		})
	}
}

func TestIn(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		search   string
		expected bool
	}{
		{
			name:     "search for 'name1' in a list containing it",
			input:    []string{"name1", "name2", "name3"},
			search:   "name1",
			expected: true,
		},
		{
			name:     "search for 'name1' in a list without it",
			input:    []string{"name2", "name3", "name4"},
			search:   "name1",
			expected: false,
		},
	}

	for _, test := range tests {
		result := In(test.input, test.search)
		if result != test.expected {
			t.Fatalf("Test '%s' failed: For input %v and search %s, expected %t but got %t", test.name, test.input, test.search, test.expected, result)
		}
	}
}

func TestMergeArgsWithCommands(t *testing.T) {
	testCases := []struct {
		name         string
		commands     []string
		args         []string
		wantCommands []string
	}{
		{
			name:         "commands with no line break",
			commands:     []string{"run server"},
			args:         []string{"--host", "localhost"},
			wantCommands: []string{"run server --host localhost"},
		},
		{
			name:         "commands with line break",
			commands:     []string{"go", "run server\n"},
			args:         []string{"--port", "8080"},
			wantCommands: []string{"go", "run server --port 8080"},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			got := MergeArgsWithCommands(test.commands, test.args)
			if diff := cmp.Diff(got, test.wantCommands); diff != "" {
				t.Fatalf("unexpected commands: %s", diff)
			}
		})
	}
}
