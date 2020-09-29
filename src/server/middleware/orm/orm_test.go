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

package orm

import (
	"net/http"
	"net/http/httptest"
	"testing"

	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/stretchr/testify/assert"
)

type mockOrmer struct {
	o.Ormer
}

func TestOrm(t *testing.T) {
	assert := assert.New(t)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := orm.FromContext(r.Context())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	})

	req1 := httptest.NewRequest(http.MethodGet, "/req1", nil)
	rec1 := httptest.NewRecorder()
	next.ServeHTTP(rec1, req1)
	assert.Equal(http.StatusInternalServerError, rec1.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/req2", nil)
	rec2 := httptest.NewRecorder()

	MiddlewareWithConfig(Config{Creator: func() o.Ormer { return &mockOrmer{} }})(next).ServeHTTP(rec2, req2)
	assert.Equal(http.StatusOK, rec2.Code)
}
