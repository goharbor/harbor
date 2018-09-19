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

package filter

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMediaTypeFilter(t *testing.T) {
	assert := assert.New(t)
	getReq := httptest.NewRequest(http.MethodGet, "/the/path", nil)
	rec := httptest.NewRecorder()
	filterContentType(getReq, rec, "application/json")
	assert.Equal(http.StatusOK, rec.Code)

	postReq := httptest.NewRequest(http.MethodPost, "/the/path", nil)
	postReq.Header.Set("Content-Type", "text/html")
	rec2 := httptest.NewRecorder()
	filterContentType(postReq, rec2, "application/json")
	assert.Equal(http.StatusUnsupportedMediaType, rec2.Code)
	postReq2 := httptest.NewRequest(http.MethodPost, "/the/path", nil)
	postReq2.Header.Set("Content-Type", "application/json; charset=utf-8")
	rec3 := httptest.NewRecorder()
	filterContentType(postReq2, rec3, "application/json")
	assert.Equal(http.StatusOK, rec3.Code)

}
