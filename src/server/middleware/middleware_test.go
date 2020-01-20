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

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
)

type middlewareTestSuite struct {
	suite.Suite
}

func (m *middlewareTestSuite) TestWithMiddlewares() {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("key", w.Header().Get("key")+"handler")
	})
	middleware1 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("key", w.Header().Get("key")+"middleware1")
			h.ServeHTTP(w, r)
		})
	}
	middleware2 := func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("key", w.Header().Get("key")+"middleware2")
			h.ServeHTTP(w, r)
		})
	}
	record := &httptest.ResponseRecorder{}
	WithMiddlewares(handler, middleware1, middleware2).ServeHTTP(record, nil)
	m.Equal("middleware1middleware2handler", record.Header().Get("key"))
}

func (m *middlewareTestSuite) TestNew() {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	f := func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		w.Header().Set("key", "value")
		next.ServeHTTP(w, r)
	}

	req1 := httptest.NewRequest(http.MethodGet, "/req", nil)
	rec1 := httptest.NewRecorder()
	New(f)(next).ServeHTTP(rec1, req1)
	m.Equal("value", rec1.Header().Get("key"))

	req2 := httptest.NewRequest(http.MethodGet, "/req", nil)
	rec2 := httptest.NewRecorder()
	New(f, func(r *http.Request) bool { return r.URL.Path == "/req" })(next).ServeHTTP(rec2, req2)
	m.Equal("", rec2.Header().Get("key"))

	req3 := httptest.NewRequest(http.MethodGet, "/req3", nil)
	rec3 := httptest.NewRecorder()
	New(f, func(r *http.Request) bool { return r.URL.Path == "/req" })(next).ServeHTTP(rec3, req3)
	m.Equal("value", rec3.Header().Get("key"))
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &middlewareTestSuite{})
}
