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

package error

import (
	"errors"
)

// ErrDupProject is the error returned when creating a duplicate project
var ErrDupProject = errors.New("duplicate project")

const (
	// ReasonNotFound indicates resource not found
	ReasonNotFound = "NotFound"
)

// KnownError represents known type errors
type KnownError struct {
	// Reason is reason of the error, such as NotFound
	Reason string
	// Message is the message of the error
	Message string
}

// Error returns the error message
func (e KnownError) Error() string {
	return e.Message
}

// Is checks whether a error is a given type error
func Is(err error, reason string) bool {
	if e, ok := err.(KnownError); ok && e.Reason == reason {
		return true
	}
	return false
}
