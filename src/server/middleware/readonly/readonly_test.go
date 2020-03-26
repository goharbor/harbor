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

package readonly

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadOnly(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	readOnly := func(readOnly bool) Config {
		return Config{
			ReadOnly: func(*http.Request) bool {
				return readOnly
			},
		}
	}

	req1 := httptest.NewRequest(http.MethodDelete, "/resource", nil)
	rec1 := httptest.NewRecorder()
	MiddlewareWithConfig(readOnly(true))(next).ServeHTTP(rec1, req1)
	assert.Equal(http.StatusForbidden, rec1.Code)

	req2 := httptest.NewRequest(http.MethodDelete, "/resource", nil)
	rec2 := httptest.NewRecorder()
	MiddlewareWithConfig(readOnly(false))(next).ServeHTTP(rec2, req2)
	assert.Equal(http.StatusOK, rec2.Code)

	// safe method
	req3 := httptest.NewRequest(http.MethodGet, "/resource", nil)
	rec3 := httptest.NewRecorder()
	MiddlewareWithConfig(readOnly(true))(next).ServeHTTP(rec3, req3)
	assert.Equal(http.StatusOK, rec3.Code)
}
