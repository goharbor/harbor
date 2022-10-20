/*
Copyright The ORAS Authors.
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

package content

import "errors"

// Common errors
var (
	ErrNotFound           = errors.New("not_found")
	ErrNoName             = errors.New("no_name")
	ErrUnsupportedSize    = errors.New("unsupported_size")
	ErrUnsupportedVersion = errors.New("unsupported_version")
	ErrInvalidReference   = errors.New("invalid_reference")
)

// FileStore errors
var (
	ErrPathTraversalDisallowed = errors.New("path_traversal_disallowed")
	ErrOverwriteDisallowed     = errors.New("overwrite_disallowed")
)
