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

package sizequota

import (
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewPatchBlobInterceptor(t *testing.T) {
	bi := NewPatchBlobInterceptor()
	assert.NotNil(t, bi)
}

func TestPatchBlobHandleRequest(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	bi := NewPatchBlobInterceptor()
	assert.Nil(t, bi.HandleRequest(req))
}

func TestPatchBlobHandleResponse(t *testing.T) {
	req, _ := http.NewRequest("PUT", "http://127.0.0.1:5000/v2/library/ubuntu/manifests/14.04", nil)
	rw := httptest.NewRecorder()
	customResW := util.CustomResponseWriter{ResponseWriter: rw}
	customResW.WriteHeader(400)
	NewPatchBlobInterceptor().HandleResponse(customResW, req)
}
