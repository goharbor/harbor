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
	"github.com/stretchr/testify/suite"
)

type PostInitiateBlobUploadMiddlewareTestSuite struct {
	htesting.Suite
}

func (suite *PostInitiateBlobUploadMiddlewareTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.Suite.ClearTables = []string{"project_blob", "blob"}
}

func (suite *PostInitiateBlobUploadMiddlewareTestSuite) TestMountBlob() {
	suite.WithProject(func(projectID int64, projectName string) {
		ctx := suite.Context()

		digest := suite.DigestString()
		_, err := blob.Ctl.Ensure(ctx, digest, "", 512)
		suite.Nil(err)

		suite.WithProject(func(id int64, name string) {
			query := map[string]string{"mount": digest}
			req := suite.NewRequest(http.MethodPost, fmt.Sprintf("/v2/%s/photon/blobs/uploads", name), nil, query)
			res := httptest.NewRecorder()

			next := suite.NextHandler(http.StatusCreated, nil)

			PostInitiateBlobUploadMiddleware()(next).ServeHTTP(res, req)

			exist, err := blob.Ctl.Exist(ctx, digest, blob.IsAssociatedWithProject(id))
			suite.Nil(err)
			suite.True(exist)

			blob, err := blob.Ctl.Get(ctx, digest)
			if suite.Nil(err) {
				suite.Equal(digest, blob.Digest)
				suite.Equal(int64(512), blob.Size)
			}
		})
	})
}

func TestPostInitiateBlobUploadMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &PostInitiateBlobUploadMiddlewareTestSuite{})
}
