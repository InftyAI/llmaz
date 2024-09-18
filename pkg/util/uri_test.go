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
	"reflect"
	"testing"
)

func TestParseOSS(t *testing.T) {
	testCases := []struct {
		name          string
		address       string
		wantEndpoint  string
		wantBucket    string
		wantModelPath string
		failed        bool
	}{
		{
			name:          "normal address",
			address:       "bucket.endpoint/model/to/path",
			wantEndpoint:  "endpoint",
			wantBucket:    "bucket",
			wantModelPath: "model/to/path",
			failed:        false,
		},
		{
			name:    "no buckets",
			address: "endpoint/model/to/path",
			failed:  true,
		},
		{
			name:    "no buckets",
			address: "endpoint/model/to/path",
			failed:  true,
		},
		{
			name:    "no path",
			address: "bucket.endpoint",
			failed:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotEndpoint, gotBucket, gotModelPath, err := ParseOSS(tc.address)
			if tc.failed && err == nil {
				t.Fatal("test should fail")
			}
			if tc.wantEndpoint != gotEndpoint || tc.wantBucket != gotBucket || tc.wantModelPath != gotModelPath {
				t.Fatal("unexpected result")
			}
		})
	}
}

func TestParseURI(t *testing.T) {
	tests := []struct {
		name          string
		uri           string
		expectedProto string
		expectedAddr  string
		expectedErr   error
	}{
		{
			name:          "valid uri with http",
			uri:           "http://example.com",
			expectedProto: "HTTP",
			expectedAddr:  "example.com",
			expectedErr:   nil,
		},
		{
			name:          "invalid URI",
			uri:           "invalid_uri",
			expectedProto: "",
			expectedAddr:  "",
			expectedErr:   errors.New("uri format error"),
		},
		{
			name:          "uri with incorrect format",
			uri:           "missing-protocol",
			expectedProto: "",
			expectedAddr:  "",
			expectedErr:   errors.New("uri format error"),
		},
	}

	for _, test := range tests {
		proto, addr, err := ParseURI(test.uri)

		if proto != test.expectedProto {
			t.Errorf("Test '%s' failed: Expected protocol %s, but got %s", test.name, test.expectedProto, proto)
		}

		if addr != test.expectedAddr {
			t.Errorf("Test '%s' failed: Expected address %s, but got %s", test.name, test.expectedAddr, addr)
		}

		if !reflect.DeepEqual(err, test.expectedErr) {
			t.Errorf("Test '%s' failed: Expected error %v, but got %v", test.name, test.expectedErr, err)
		}
	}
}
