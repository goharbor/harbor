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

package log

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
	suite.Suite
}

func (suite *MiddlewareTestSuite) TestMiddleware() {
	next := func(fields log.Fields) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.G(r.Context()).WithFields(fields).Info("this is message")

			w.WriteHeader(http.StatusOK)
		})
	}

	{
		req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		Middleware()(next(nil)).ServeHTTP(rr, req)
	}

	{
		req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil)
		req.Header.Set("X-Request-ID", "fd6139e6-9092-4181-9220-42d3d48bf658")
		rr := httptest.NewRecorder()

		Middleware()(next(nil)).ServeHTTP(rr, req)
	}

	{
		req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil)
		req.Header.Set("X-Request-ID", "fd6139e6-9092-4181-9220-42d3d48bf658")
		rr := httptest.NewRecorder()

		Middleware()(next(log.Fields{"method": req.Method})).ServeHTTP(rr, req)
	}
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}
