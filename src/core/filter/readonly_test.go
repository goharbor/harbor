// Copyright 2018 Project Harbor Authors
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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/stretchr/testify/assert"
)

func TestReadonlyFilter(t *testing.T) {

	var defaultConfig = map[string]interface{}{
		common.ReadOnly: true,
	}
	config.Upload(defaultConfig)

	assert := assert.New(t)
	req1, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/ubuntu", nil)
	rec := httptest.NewRecorder()
	filter(req1, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req2, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/hello-world", nil)
	rec = httptest.NewRecorder()
	filter(req2, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req3, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags/14.04", nil)
	rec = httptest.NewRecorder()
	filter(req3, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req4, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags/latest", nil)
	rec = httptest.NewRecorder()
	filter(req4, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req5, _ := http.NewRequest("DELETE", "http://127.0.0.1:5000/api/repositories/library/vmware/hello-world", nil)
	rec = httptest.NewRecorder()
	filter(req5, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)

	req6, _ := http.NewRequest("POST", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags", nil)
	rec = httptest.NewRecorder()
	filter(req6, rec)
	assert.Equal(http.StatusServiceUnavailable, rec.Code)
}

func TestMatchRetag(t *testing.T) {
	req1, _ := http.NewRequest("POST", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags", nil)
	assert.True(t, matchRetag(req1))

	req2, _ := http.NewRequest("POST", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags/v1.0", nil)
	assert.False(t, matchRetag(req2))

	req3, _ := http.NewRequest("GET", "http://127.0.0.1:5000/api/repositories/library/hello-world/tags", nil)
	assert.False(t, matchRetag(req3))

	req4, _ := http.NewRequest("POST", "http://127.0.0.1:5000/api/repositories/library/hello-world", nil)
	assert.False(t, matchRetag(req4))
}
