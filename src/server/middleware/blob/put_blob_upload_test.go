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
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goharbor/harbor/src/controller/blob"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type PutBlobUploadMiddlewareTestSuite struct {
	htesting.Suite
}

func (suite *PutBlobUploadMiddlewareTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"project_blob", "blob"}
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestDataInBody() {
	suite.WithProject(func(projectID int64, projectName string) {
		req := suite.NewRequest(http.MethodPut, fmt.Sprintf("/v2/%s/photon/blobs/uploads/%s", projectName, uuid.New().String()), nil)
		req.Header.Set("Content-Length", "512")
		res := httptest.NewRecorder()

		digest := suite.DigestString()

		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": digest})
		PutBlobUploadMiddleware()(next).ServeHTTP(res, req)

		exist, err := blob.Ctl.Exist(suite.Context(), digest, blob.IsAssociatedWithProject(projectID))
		suite.Nil(err)
		suite.True(exist)

		blob, err := blob.Ctl.Get(suite.Context(), digest)
		suite.Nil(err)
		suite.Equal(digest, blob.Digest)
		suite.Equal(int64(512), blob.Size)
	})
}

func (suite *PutBlobUploadMiddlewareTestSuite) TestWithoutBody() {
	suite.WithProject(func(projectID int64, projectName string) {
		sessionID := uuid.New().String()
		path := fmt.Sprintf("/v2/%s/photon/blobs/uploads/%s", projectName, sessionID)

		{
			req := httptest.NewRequest(http.MethodPatch, path, nil)
			res := httptest.NewRecorder()

			next := suite.NextHandler(http.StatusAccepted, map[string]string{"Range": "0-511"})
			PatchBlobUploadMiddleware()(next).ServeHTTP(res, req)
			suite.Equal(http.StatusAccepted, res.Code)
		}

		req := suite.NewRequest(http.MethodPut, path, nil)
		res := httptest.NewRecorder()

		digest := suite.DigestString()

		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": digest})
		PutBlobUploadMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		exist, err := blob.Ctl.Exist(suite.Context(), digest, blob.IsAssociatedWithProject(projectID))
		suite.Nil(err)
		suite.True(exist)

		blob, err := blob.Ctl.Get(suite.Context(), digest)
		if suite.Nil(err) {
			suite.Equal(digest, blob.Digest)
			suite.Equal(int64(512), blob.Size)
		}
	})
}

func TestPutBlobUploadMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &PutBlobUploadMiddlewareTestSuite{})
}
