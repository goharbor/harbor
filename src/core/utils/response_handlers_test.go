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
package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusRespHandler(t *testing.T) {
	assert := assert.New(t)
	h := NewStatusRespHandler(http.StatusCreated)
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusCreated)
	recorder.WriteString("test passed")
	resp1 := recorder.Result()
	err := h.Handle(resp1)
	assert.Nil(err)
	recorder2 := httptest.NewRecorder()
	recorder2.WriteHeader(http.StatusForbidden)
	recorder2.WriteString("test forbidden")
	resp2 := recorder2.Result()
	err = h.Handle(resp2)
	assert.NotNil(err)
	assert.Contains(err.Error(), "forbidden")
}
