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
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/pkg/blob/models"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/goharbor/harbor/src/testing/mock"
	distributiontesting "github.com/goharbor/harbor/src/testing/pkg/distribution"
	"github.com/stretchr/testify/suite"
)

type PutManifestMiddlewareTestSuite struct {
	RequestMiddlewareTestSuite

	unmarshalManifest func(r *http.Request) (distribution.Manifest, distribution.Descriptor, error)
	manifest          distribution.Manifest
}

func (suite *PutManifestMiddlewareTestSuite) SetupTest() {
	suite.RequestMiddlewareTestSuite.SetupTest()

	suite.unmarshalManifest = unmarshalManifest
	suite.manifest = &distributiontesting.Manifest{}

	mock.OnAnything(suite.manifest, "References").Return([]distribution.Descriptor{
		{Digest: "blob1", Size: 10, MediaType: schema2.MediaTypeLayer},
		{Digest: "blob2", Size: 20, MediaType: schema2.MediaTypeLayer},
		{Digest: "blob3", Size: 30, MediaType: schema2.MediaTypeForeignLayer},
		{Digest: "blob4", Size: 40, MediaType: schema2.MediaTypeForeignLayer},
	})

	unmarshalManifest = func(r *http.Request) (distribution.Manifest, distribution.Descriptor, error) {
		return suite.manifest, distribution.Descriptor{Digest: "digest", Size: 100}, nil
	}
}

func (suite *PutManifestMiddlewareTestSuite) TearDownTest() {
	suite.RequestMiddlewareTestSuite.TearDownTest()

	unmarshalManifest = suite.unmarshalManifest
}

func (suite *PutManifestMiddlewareTestSuite) TestMiddleware() {
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
		// manifest not associated with project and blobs are already associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()
		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(nil, nil).Once()
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

	{
		// manifest not associated with project and some blobs are not associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()

		missing := func(ctx context.Context, projectID int64, blobs []*models.Blob) []*models.Blob {
			return blobs[:1]
		}

		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(missing, nil).Once()
		mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
			resources := args.Get(3).(types.ResourceList)
			suite.Len(resources, 2)
			suite.Equal(resources[types.ResourceStorage], int64(100+10))
			suite.Equal(resources[types.ResourceCount], int64(1))

			f := args.Get(4).(func() error)
			f()
		})

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

	{
		// manifest not associated with project and some blobs include foreign layers are not associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()

		missing := func(ctx context.Context, projectID int64, blobs []*models.Blob) []*models.Blob {
			return blobs[1:]
		}

		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(missing, nil).Once()
		mock.OnAnything(suite.quotaController, "Request").Return(nil).Once().Run(func(args mock.Arguments) {
			resources := args.Get(3).(types.ResourceList)
			suite.Len(resources, 2)
			suite.Equal(resources[types.ResourceStorage], int64(100+20))
			suite.Equal(resources[types.ResourceCount], int64(1))

			f := args.Get(4).(func() error)
			f()
		})

		req := httptest.NewRequest(http.MethodPut, "/v2/library/photon/manifests/2.0", nil)
		rr := httptest.NewRecorder()

		PutManifestMiddleware()(next).ServeHTTP(rr, req)
		suite.Equal(http.StatusOK, rr.Code)
	}

	{
		// manifest not associated with project and only foreign layers are not associated with project
		mock.OnAnything(suite.blobController, "Exist").Return(false, nil).Once()

		missing := func(ctx context.Context, projectID int64, blobs []*models.Blob) []*models.Blob {
			return blobs[2:]
		}

		mock.OnAnything(suite.blobController, "FindMissingAssociationsForProject").Return(missing, nil).Once()
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
