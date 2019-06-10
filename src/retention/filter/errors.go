// Copyright 2019 Project Harbor Authors
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

package filter

import "fmt"

// ErrMissingMetadata builds an error that indicates a required metadata key is missing from the filter metadata
func ErrMissingMetadata(key string) error {
	return fmt.Errorf("filter: metadata: missing required key %s", key)
}

// ErrWrongMetadataType builds an error that indicates a metadata value is of the wrong type
func ErrWrongMetadataType(key, t string) error {
	return fmt.Errorf("filter: metadata: %s is wrong type (not a %s)", key, t)
}

// ErrInvalidMetadata builds a generic error that indicates a problem with the filter metadata
func ErrInvalidMetadata(key, err string) error {
	return fmt.Errorf("filter: metadata: %s: %s", key, err)
}
