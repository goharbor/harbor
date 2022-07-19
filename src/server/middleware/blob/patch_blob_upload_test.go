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

package blob

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type PatchBlobUploadMiddlewareTestSuite struct {
	suite.Suite
}

func (suite *PatchBlobUploadMiddlewareTestSuite) TestMiddleware() {
	next := func(rangeHeader string) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			w.Header().Set("Range", rangeHeader)
		})
	}

	sessionID := uuid.New().String()
	path := fmt.Sprintf("/v2/library/photon/blobs/uploads/%s", sessionID)

	req := httptest.NewRequest(http.MethodPatch, path, nil)
	res := httptest.NewRecorder()
	PatchBlobUploadMiddleware()(next("bad value")).ServeHTTP(res, req)
	suite.Equal(http.StatusInternalServerError, res.Code)

	req = httptest.NewRequest(http.MethodPatch, path, nil)
	res = httptest.NewRecorder()
	PatchBlobUploadMiddleware()(next("0-511")).ServeHTTP(res, req)
	suite.Equal(http.StatusAccepted, res.Code)

	size, err := blob.Ctl.GetAcceptedBlobSize(context.TODO(), sessionID)
	suite.Nil(err)
	suite.Equal(int64(512), size)
}

func TestPatchBlobUploadMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &PatchBlobUploadMiddlewareTestSuite{})
}
