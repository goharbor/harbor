package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	beegocontext "github.com/beego/beego/v2/server/web/context"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/router"
	"github.com/goharbor/harbor/src/testing/mock"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	arttesting "github.com/goharbor/harbor/src/testing/pkg/artifact"
	testmanifest "github.com/goharbor/harbor/src/testing/pkg/cached/manifest/redis"
	regtesting "github.com/goharbor/harbor/src/testing/pkg/registry"
)

var (
	OCIManifest = `{ 
		"schemaVersion": 2,
		"mediaType": "application/vnd.oci.image.manifest.v1+json",
		"config": {
		   "mediaType": "application/vnd.example.sbom",
		   "digest": "sha256:5891b5b522d5df086d0ff0b110fbd9d21bb4fc7163af34d08286a2e846f6be03",
		   "size": 123
		},
		"layers": [
		   {
			  "mediaType": "application/vnd.example.data.v1.tar+gzip",
			  "digest": "sha256:e258d248fda94c63753607f7c4494ee0fcbe92f1a76bfdac795c9d84101eb317",
			  "size": 1234
		   }
		],
		"annotations": {
		   "name": "test-image"
		}
	 }`
)

func TestReferrersHandlerOK(t *testing.T) {
	rec := httptest.NewRecorder()
	digestVal := "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
	req, err := http.NewRequest("GET", "/v2/test/repository/referrers/sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b", nil)
	if err != nil {
		t.Fatal(err)
	}
	input := &beegocontext.BeegoInput{}
	input.SetParam(":reference", digestVal)
	*req = *(req.WithContext(context.WithValue(req.Context(), router.ContextKeyInput{}, input)))

	artifactMock := &arttesting.Manager{}
	accessoryMock := &accessorytesting.Manager{}

	artifactMock.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).
		Return(&artifact.Artifact{
			Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
			ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
			MediaType:         "application/vnd.example.sbom",
			ArtifactType:      "application/vnd.example.sbom",
			Size:              1000,
			Annotations: map[string]string{
				"name": "test-image",
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
					SubArtifactRepo:   "goharbor",
					Type:              accessorymodel.TypeCosignSignature,
					Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
				},
			},
		}, nil)

	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(OCIManifest))
	if err != nil {
		t.Fatal(err)
	}
	regCliMock := &regtesting.Client{}
	config.DefaultMgr().Set(context.TODO(), "cache_enabled", false)
	mock.OnAnything(regCliMock, "PullManifest").Return(manifest, "", nil)

	handler := &referrersHandler{
		artifactManager:  artifactMock,
		accessoryManager: accessoryMock,
		registryClient:   regCliMock,
	}

	handler.ServeHTTP(rec, req)

	// check that the response has the expected status code (200 OK)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}
	index := &ocispec.Index{}
	json.Unmarshal(rec.Body.Bytes(), index)
	if index.Manifests[0].ArtifactType != "application/vnd.example.sbom" {
		t.Errorf("Expected response body %s, but got %s", "application/vnd.example.sbom", rec.Body.String())
	}
	_, content, _ := manifest.Payload()
	assert.Equal(t, int64(len(content)), index.Manifests[0].Size)
}

func TestReferrersHandlerSavetoCache(t *testing.T) {
	rec := httptest.NewRecorder()
	digestVal := "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
	req, err := http.NewRequest("GET", "/v2/test/repository/referrers/sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b", nil)
	if err != nil {
		t.Fatal(err)
	}
	input := &beegocontext.BeegoInput{}
	input.SetParam(":reference", digestVal)
	*req = *(req.WithContext(context.WithValue(req.Context(), router.ContextKeyInput{}, input)))

	artifactMock := &arttesting.Manager{}
	accessoryMock := &accessorytesting.Manager{}

	artifactMock.On("GetByDigest", mock.Anything, mock.Anything, mock.Anything).
		Return(&artifact.Artifact{
			Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
			ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
			MediaType:         "application/vnd.example.sbom",
			ArtifactType:      "application/vnd.example.sbom",
			Size:              1000,
			Annotations: map[string]string{
				"name": "test-image",
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
					SubArtifactRepo:   "goharbor",
					Type:              accessorymodel.TypeCosignSignature,
					Digest:            "sha256:4911bb745e19a6b5513755f3d033f10ef10c34b40edc631809e28be8a7c005f6",
				},
			},
		}, nil)

	manifest, _, err := distribution.UnmarshalManifest(v1.MediaTypeImageManifest, []byte(OCIManifest))
	if err != nil {
		t.Fatal(err)
	}

	// cache_enabled pull from cahce
	config.DefaultMgr().Set(context.TODO(), "cache_enabled", true)
	cacheManagerMock := &testmanifest.CachedManager{}
	mock.OnAnything(cacheManagerMock, "Get").Return(nil, fmt.Errorf("unable to do stuff: %w", cache.ErrNotFound))
	regCliMock := &regtesting.Client{}
	mock.OnAnything(regCliMock, "PullManifest").Return(manifest, "", nil)
	mock.OnAnything(cacheManagerMock, "Save").Return(nil)

	handler := &referrersHandler{
		artifactManager:  artifactMock,
		accessoryManager: accessoryMock,
		registryClient:   regCliMock,
		maniCacheManager: cacheManagerMock,
	}

	handler.ServeHTTP(rec, req)

	// check that the response has the expected status code (200 OK)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}
	index := &ocispec.Index{}
	json.Unmarshal(rec.Body.Bytes(), index)
	if index.Manifests[0].ArtifactType != "application/vnd.example.sbom" {
		t.Errorf("Expected response body %s, but got %s", "application/vnd.example.sbom", rec.Body.String())
	}
	_, content, _ := manifest.Payload()
	assert.Equal(t, int64(len(content)), index.Manifests[0].Size)
}

func TestReferrersHandlerEmpty(t *testing.T) {
	rec := httptest.NewRecorder()
	digestVal := "sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b"
	req, err := http.NewRequest("GET", "/v2/test/repository/referrers/sha256:6c3c624b58dbbcd3c0dd82b4c53f04194d1247c6eebdaab7c610cf7d66709b3b", nil)
	if err != nil {
		t.Fatal(err)
	}
	input := &beegocontext.BeegoInput{}
	input.SetParam(":reference", digestVal)
	*req = *(req.WithContext(context.WithValue(req.Context(), router.ContextKeyInput{}, input)))

	artifactMock := &arttesting.Manager{}
	accessoryMock := &accessorytesting.Manager{}

	accessoryMock.On("Count", mock.Anything, mock.Anything).
		Return(int64(0), nil)
	accessoryMock.On("List", mock.Anything, mock.Anything).
		Return([]accessorymodel.Accessory{}, nil)

	handler := &referrersHandler{
		artifactManager:  artifactMock,
		accessoryManager: accessoryMock,
	}

	handler.ServeHTTP(rec, req)

	// check that the response has the expected status code (200 OK)
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}
	index := &ocispec.Index{}
	json.Unmarshal([]byte(rec.Body.String()), index)
	if index.SchemaVersion != 0 && len(index.Manifests) != -0 {
		t.Errorf("Expected empty response body, but got %s", rec.Body.String())
	}
}

func TestReferrersHandler400(t *testing.T) {
	rec := httptest.NewRecorder()
	digestVal := "invalid"
	req, err := http.NewRequest("GET", "/v2/test/repository/referrers/invalid", nil)
	if err != nil {
		t.Fatal(err)
	}
	input := &beegocontext.BeegoInput{}
	input.SetParam(":reference", digestVal)
	*req = *(req.WithContext(context.WithValue(req.Context(), router.ContextKeyInput{}, input)))

	artifactMock := &arttesting.Manager{}
	accessoryMock := &accessorytesting.Manager{}
	handler := &referrersHandler{
		artifactManager:  artifactMock,
		accessoryManager: accessoryMock,
	}

	handler.ServeHTTP(rec, req)
	// check that the response has the expected status code (200 OK)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, rec.Code)
	}
}
