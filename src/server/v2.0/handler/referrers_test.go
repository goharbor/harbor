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

package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	beegocontext "github.com/beego/beego/v2/server/web/context"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/config"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/router"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	"github.com/goharbor/harbor/src/testing/mock"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	arttesting "github.com/goharbor/harbor/src/testing/pkg/artifact"
	regtesting "github.com/goharbor/harbor/src/testing/pkg/registry"
)

// newTestRequestContext creates a request context with mock security and beego params
func newTestRequestContext(req *http.Request, params map[string]string) *http.Request {
	input := &beegocontext.BeegoInput{}
	for k, v := range params {
		input.SetParam(k, v)
	}
	ctx := context.WithValue(req.Context(), router.ContextKeyInput{}, input)

	// Set up mock security context with admin access
	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(true)
	secCtx.On("IsSysAdmin").Return(true)
	mock.OnAnything(secCtx, "Can").Return(true)
	ctx = security.NewContext(ctx, secCtx)

	*req = *(req.WithContext(ctx))
	return req
}

var testOCIManifest = `{
	"schemaVersion": 2,
	"mediaType": "application/vnd.oci.image.manifest.v1+json",
	"config": {
	   "mediaType": "application/vnd.cncf.notary.signature",
	   "digest": "sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03",
	   "size": 123
	},
	"layers": [
	   {
		  "mediaType": "application/vnd.cncf.notary.signature",
		  "digest": "sha256:e258d248fda94c63753607f7c4494ee0fcbe92f1a76bfdac795c9d84101eb317",
		  "size": 1234
	   }
	],
	"annotations": {
	   "name": "test-signature"
	}
}`

func TestReferrersAPIHandlerOK(t *testing.T) {
	rec := httptest.NewRecorder()
	digestVal := "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
	req, err := http.NewRequest("GET",
		"/api/v2.0/projects/testproject/repositories/testrepo/artifacts/"+digestVal+"/referrers", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = newTestRequestContext(req, map[string]string{
		":project_name": "testproject",
		":repo_name":    "testrepo",
		":reference":    digestVal,
	})

	artifactMock := &arttesting.Manager{}
	accessoryMock := &accessorytesting.Manager{}

	artifactMock.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).
		Return(&artifact.Artifact{
			Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
			ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
			MediaType:         "application/vnd.cncf.notary.signature",
			ArtifactType:      "application/vnd.cncf.notary.signature",
			RepositoryName:    "testproject/testrepo",
			Size:              1000,
			Annotations: map[string]string{
				"name": "test-signature",
			},
		}, nil)

	accessoryMock.On("Count", mock.Anything, mock.Anything).
		Return(int64(1), nil)
	accessoryMock.On("List", mock.Anything, mock.Anything).
		Return([]accessorymodel.Accessory{
			&basemodel.Default{
				Data: accessorymodel.AccessoryData{
					ID:                1,
					ArtifactID:        2,
					SubArtifactDigest: digestVal,
					SubArtifactRepo:   "testproject/testrepo",
					Type:              accessorymodel.TypeNotationSignature,
					Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
				},
			},
		}, nil)

	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(testOCIManifest))
	if err != nil {
		t.Fatal(err)
	}

	regCliMock := &regtesting.Client{}
	config.DefaultMgr().Set(context.TODO(), "cache_enabled", false)
	mock.OnAnything(regCliMock, "PullManifest").Return(manifest, "", nil)

	handler := &referrersAPIHandler{
		BaseAPI:          &BaseAPI{},
		artifactManager:  artifactMock,
		accessoryManager: accessoryMock,
		registryClient:   regCliMock,
	}

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, referrersMediaType, rec.Header().Get("Content-Type"))

	index := &ocispec.Index{}
	err = json.Unmarshal(rec.Body.Bytes(), index)
	assert.NoError(t, err)
	assert.Equal(t, referrersSchemaVersion, index.SchemaVersion)
	assert.Len(t, index.Manifests, 1)
	assert.Equal(t, "application/vnd.cncf.notary.signature", index.Manifests[0].ArtifactType)
}

func TestReferrersAPIHandlerEmpty(t *testing.T) {
	rec := httptest.NewRecorder()
	digestVal := "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
	req, err := http.NewRequest("GET",
		"/api/v2.0/projects/testproject/repositories/testrepo/artifacts/"+digestVal+"/referrers", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = newTestRequestContext(req, map[string]string{
		":project_name": "testproject",
		":repo_name":    "testrepo",
		":reference":    digestVal,
	})

	accessoryMock := &accessorytesting.Manager{}
	accessoryMock.On("Count", mock.Anything, mock.Anything).
		Return(int64(0), nil)
	accessoryMock.On("List", mock.Anything, mock.Anything).
		Return([]accessorymodel.Accessory{}, nil)

	handler := &referrersAPIHandler{
		BaseAPI:          &BaseAPI{},
		artifactManager:  &arttesting.Manager{},
		accessoryManager: accessoryMock,
	}

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	index := &ocispec.Index{}
	err = json.Unmarshal(rec.Body.Bytes(), index)
	assert.NoError(t, err)
	assert.Empty(t, index.Manifests)
}

func TestReferrersAPIHandlerInvalidDigest(t *testing.T) {
	rec := httptest.NewRecorder()
	req, err := http.NewRequest("GET",
		"/api/v2.0/projects/testproject/repositories/testrepo/artifacts/invalid/referrers", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = newTestRequestContext(req, map[string]string{
		":project_name": "testproject",
		":repo_name":    "testrepo",
		":reference":    "invalid",
	})

	handler := &referrersAPIHandler{
		BaseAPI:          &BaseAPI{},
		artifactManager:  &arttesting.Manager{},
		accessoryManager: &accessorytesting.Manager{},
	}

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestReferrersAPIHandlerWithArtifactTypeFilter(t *testing.T) {
	rec := httptest.NewRecorder()
	digestVal := "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
	req, err := http.NewRequest("GET",
		"/api/v2.0/projects/testproject/repositories/testrepo/artifacts/"+digestVal+"/referrers?artifactType=application/vnd.cncf.notary.signature", nil)
	if err != nil {
		t.Fatal(err)
	}
	req = newTestRequestContext(req, map[string]string{
		":project_name": "testproject",
		":repo_name":    "testrepo",
		":reference":    digestVal,
	})

	artifactMock := &arttesting.Manager{}
	accessoryMock := &accessorytesting.Manager{}

	artifactMock.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).
		Return(&artifact.Artifact{
			Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
			ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
			MediaType:         "application/vnd.cncf.notary.signature",
			ArtifactType:      "application/vnd.cncf.notary.signature",
			RepositoryName:    "testproject/testrepo",
			Size:              1000,
		}, nil)

	accessoryMock.On("Count", mock.Anything, mock.Anything).
		Return(int64(1), nil)
	accessoryMock.On("List", mock.Anything, mock.Anything).
		Return([]accessorymodel.Accessory{
			&basemodel.Default{
				Data: accessorymodel.AccessoryData{
					ID:                1,
					ArtifactID:        2,
					SubArtifactDigest: digestVal,
					SubArtifactRepo:   "testproject/testrepo",
					Type:              accessorymodel.TypeNotationSignature,
					Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
				},
			},
		}, nil)

	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(testOCIManifest))
	if err != nil {
		t.Fatal(err)
	}

	regCliMock := &regtesting.Client{}
	config.DefaultMgr().Set(context.TODO(), "cache_enabled", false)
	mock.OnAnything(regCliMock, "PullManifest").Return(manifest, "", nil)

	handler := &referrersAPIHandler{
		BaseAPI:          &BaseAPI{},
		artifactManager:  artifactMock,
		accessoryManager: accessoryMock,
		registryClient:   regCliMock,
	}

	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "artifactType", rec.Header().Get("OCI-Filters-Applied"))

	index := &ocispec.Index{}
	err = json.Unmarshal(rec.Body.Bytes(), index)
	assert.NoError(t, err)
	assert.Len(t, index.Manifests, 1)
}
