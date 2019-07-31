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

package interceptor

import (
	"net/http"
)

// Builder interceptor builder
type Builder interface {
	// Build build interceptor from http.Request returns nil if interceptor not match the request
	Build(*http.Request) Interceptor
}

// Interceptor interceptor for middleware
type Interceptor interface {
	// HandleRequest ...
	HandleRequest(*http.Request) error

	// HandleResponse won't return any error
	HandleResponse(http.ResponseWriter, *http.Request)
}
