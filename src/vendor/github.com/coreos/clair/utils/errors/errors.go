// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package errors defines error types that are used in several modules
package errors

import "errors"

var (
	// ErrFilesystem occurs when a filesystem interaction fails.
	ErrFilesystem = errors.New("something went wrong when interacting with the fs")

	// ErrCouldNotDownload occurs when a download fails.
	ErrCouldNotDownload = errors.New("could not download requested resource")

	// ErrNotFound occurs when a resource could not be found.
	ErrNotFound = errors.New("the resource cannot be found")

	// ErrCouldNotParse is returned when a fetcher fails to parse the update data.
	ErrCouldNotParse = errors.New("updater/fetchers: could not parse")
)

// ErrBadRequest occurs when a method has been passed an inappropriate argument.
type ErrBadRequest struct {
	s string
}

// NewBadRequestError instantiates a ErrBadRequest with the specified message.
func NewBadRequestError(message string) error {
	return &ErrBadRequest{s: message}
}

func (e *ErrBadRequest) Error() string {
	return e.s
}
