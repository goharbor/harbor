package registry

import (
	"context"
	"encoding/json"
	beegocontext "github.com/beego/beego/v2/server/web/context"
	"github.com/goharbor/harbor/src/lib/q"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	basemodel "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/server/router"
	"github.com/goharbor/harbor/src/testing/mock"
	accessorytesting "github.com/goharbor/harbor/src/testing/pkg/accessory"
	arttesting "github.com/goharbor/harbor/src/testing/pkg/artifact"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"net/http"
	"net/http/httptest"
	"testing"
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
			Size:              1000,
			Annotations: map[string]string{
				"name": "test-image",
			},
		}, nil)

	accessoryMock.On("Count", mock.Anything, q.New(q.KeyWords{"SubjectArtifactDigest": digestVal})).
		Return(int64(1), nil)
	accessoryMock.On("List", mock.Anything, q.New(q.KeyWords{"SubjectArtifactDigest": digestVal})).
		Return([]accessorymodel.Accessory{
			&basemodel.Default{
				Data: accessorymodel.AccessoryData{
					ID:                1,
					ArtifactID:        2,
					SubArtifactDigest: digestVal,
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
	if index.Manifests[0].ArtifactType != "signature.cosign" {
		t.Errorf("Expected response body %s, but got %s", "signature.cosign", rec.Body.String())
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
