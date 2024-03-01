package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	beegocontext "github.com/beego/beego/v2/server/web/context"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/server/router"
	"github.com/goharbor/harbor/src/testing/mock"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	arttesting "github.com/goharbor/harbor/src/testing/pkg/artifact"
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
			Digest:            digestVal,
			ManifestMediaType: "application/vnd.oci.image.manifest.v1+json",
			MediaType:         "application/vnd.example.sbom",
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
				},
			},
		}, nil)

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
	if index.Manifests[0].ArtifactType != "application/vnd.example.sbom" {
		t.Errorf("Expected response body %s, but got %s", "application/vnd.example.sbom", rec.Body.String())
	}
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
