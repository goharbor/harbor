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

package quota

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/stretchr/testify/suite"
)

type PutManifestMiddlewareTestSuite struct {
	RequestMiddlewareTestSuite
}

func (suite *PutManifestMiddlewareTestSuite) TestMiddleware() {
	original := parseManifestDigestAndSize
	defer func() {
		parseManifestDigestAndSize = original
	}()

	parseManifestDigestAndSize = func(r *http.Request) (string, int64, error) {
		return "digest", 100, nil
	}

	mock.OnAnything(suite.quotaController, "IsEnabled").Return(true, nil)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	{
		mock.OnAnything(suite.blobController, "Exist").Return(true, nil).Once()

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

	{
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()
		mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
			resources := args.Get(3).(types.ResourceList)
			suite.Len(resources, 2)
			suite.Equal(resources[types.ResourceStorage], int64(100))
			suite.Equal(resources[types.ResourceCount], int64(1))

			f := args.Get(4).(func() error)
			f()
		})

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}
}

func TestPutManifestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &PutManifestMiddlewareTestSuite{})
}
