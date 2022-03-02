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

package config

import (
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/pkg/config/db"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/stretchr/testify/suite"
)

type MiddlewareTestSuite struct {
	suite.Suite
}

func (suite *MiddlewareTestSuite) TestMiddleware() {
	next := func() http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, ok := config.FromContext(r.Context())
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		})
	}

	{
		config.Register(common.DBCfgManager, db.NewDBCfgManager())
		req := httptest.NewRequest("GET", "/v1/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()
		Middleware()(next()).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}
