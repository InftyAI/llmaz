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
	"errors"
	"fmt"
	"strings"
)

func ParseURI(uri string) (protocol string, address string, err error) {
	parsers := strings.Split(uri, "://")
	if len(parsers) != 2 {
		return "", "", errors.New("uri format error")
	}
	return strings.ToUpper(parsers[0]), parsers[1], nil
}

// ParseOSS address looks like: <bucket>.<endpoint>/<modelPath>
func ParseOSS(address string) (endpoint, bucket, modelPath string, err error) {
	splits := strings.SplitN(address, ".", 2)
	if len(splits) != 2 {
		return "", "", "", fmt.Errorf("address not right %s", address)
	}
	bucket = splits[0]

	splits = strings.SplitN(splits[1], "/", 2)
	if len(splits) != 2 {
		return "", "", "", fmt.Errorf("address not right %s", address)
	}
	endpoint, modelPath = splits[0], splits[1]
	return endpoint, bucket, modelPath, nil
}

// ParseS3 address looks like: <bucket>/<modelPath>
func ParseS3(address string) (bucket, modelPath string, err error) {
	splits := strings.SplitN(address, "/", 2)
	if len(splits) != 2 {
		return "", "", fmt.Errorf("address not right %s", address)
	}
	bucket, modelPath = splits[0], splits[1]
	return bucket, modelPath, nil
}
