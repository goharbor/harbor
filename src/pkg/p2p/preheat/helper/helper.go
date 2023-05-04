// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helper

import (
	"strings"
)

// ImageRepository represents the image repository name
// e.g: library/ubuntu:latest
type ImageRepository string

// Valid checks if the repository name is valid
func (ir ImageRepository) Valid() bool {
	if len(ir) == 0 {
		return false
	}

	trimName := strings.TrimSpace(string(ir))
	segments := strings.SplitN(trimName, "/", 2)
	if len(segments) != 2 {
		return false
	}

	nameAndTag := segments[1]
	subSegments := strings.SplitN(nameAndTag, ":", 2)
	return len(subSegments) == 2
}

// Name returns the name of the image repository
func (ir ImageRepository) Name() string {
	// No check here, should call Valid() before calling name
	segments := strings.SplitN(string(ir), ":", 2)
	if len(segments) == 0 {
		return ""
	}

	return segments[0]
}

// Tag returns the tag of the image repository
func (ir ImageRepository) Tag() string {
	// No check here, should call Valid() before calling name
	segments := strings.SplitN(string(ir), ":", 2)
	if len(segments) < 2 {
		return ""
	}

	return segments[1]
}
