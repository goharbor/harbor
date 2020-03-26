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

package requestid

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestID(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodGet, "/req1", nil)
	rec1 := httptest.NewRecorder()
	next.ServeHTTP(rec1, req1)
	assert.Equal("", rec1.Header().Get(HeaderXRequestID))

	req2 := httptest.NewRequest(http.MethodGet, "/req2", nil)
	rec2 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec2, req2)
	assert.NotEqual("", rec2.Header().Get(HeaderXRequestID))

	req3 := httptest.NewRequest(http.MethodGet, "/req3", nil)
	req3.Header.Add(HeaderXRequestID, "852803be-e5fe-499b-bbea-c9e5b5f43916")
	rec3 := httptest.NewRecorder()
	Middleware()(next).ServeHTTP(rec3, req3)
	assert.Equal("852803be-e5fe-499b-bbea-c9e5b5f43916", rec3.Header().Get(HeaderXRequestID))
}
