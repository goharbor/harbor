package nydus

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/base"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/nydus"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	htesting "github.com/goharbor/harbor/src/testing"
)

type MiddlewareTestSuite struct {
	htesting.Suite
}

func (suite *MiddlewareTestSuite) SetupTest() {
	suite.Suite.SetupSuite()
}

func (suite *MiddlewareTestSuite) TearDownTest() {
}

func (suite *MiddlewareTestSuite) prepare(name, ref string) (distribution.Manifest, distribution.Descriptor, *http.Request) {
	body := fmt.Sprintf(`
	{
		"schemaVersion": 2,
		"config": {
		  "mediaType": "application/vnd.oci.image.config.v1+json",
		  "digest": "sha256:f7d0778a3c468a5203e95a9efd4d67ecef0d2a04866bb3320f0d5d637812aaee",
		  "size": 466
		},
		"layers": [
		  {
			"mediaType": "application/vnd.oci.image.layer.nydus.blob.v1",
			"digest": "sha256:fd9923a8e2bdc53747dbba3311be876a1deff4658785830e6030c5a8287acf74 ",
			"size": 3011,
			"annotations": {
			  "containerd.io/snapshot/nydus-blob": "true"
			}
		  },
		  {
			"mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
			"digest": "sha256:d49bf6d7db9dac935b99d4c2c846b0d280f550aae62012f888d5a6e3ca59a589",
			"size": 459,
			"annotations": {
				"containerd.io/snapshot/nydus-blob-ids": "[\"fd9923a8e2bdc53747dbba3311be876a1deff4658785830e6030c5a8287acf74\"]",
				"containerd.io/snapshot/nydus-bootstrap": "true",
				"containerd.io/snapshot/nydus-rafs-version": "5"
			}
		  }
		],
		"annotations": {
			"io.goharbor.artifact.v1alpha1.acceleration.driver.name":"nydus",
			"io.goharbor.artifact.v1alpha1.acceleration.driver.version":"5",
			"io.goharbor.artifact.v1alpha1.acceleration.source.digest":"sha256:f54a58bc1aac5ea1a25d796ae155dc228b3f0e11d046ae276b39c4bf2f13d8c4"
		}
	}
	`)

	manifest, descriptor, err := distribution.UnmarshalManifest("application/vnd.oci.image.manifest.v1+json", []byte(body))
	suite.Nil(err)

	req := suite.NewRequest(http.MethodPut, fmt.Sprintf("/v2/%s/manifests/%s", name, ref), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	info := lib.ArtifactInfo{
		Repository: name,
		Reference:  ref,
		Tag:        "latest-nydus",
		Digest:     descriptor.Digest.String(),
	}

	return manifest, descriptor, req.WithContext(lib.WithArtifactInfo(req.Context(), info))
}

func (suite *MiddlewareTestSuite) addArt(pid, repositoryID int64, repositoryName, dgt string) int64 {
	af := &artifact.Artifact{
		Type:           "Docker-Image",
		ProjectID:      pid,
		RepositoryID:   repositoryID,
		RepositoryName: repositoryName,
		Digest:         dgt,
		Size:           1024,
		PushTime:       time.Now(),
		PullTime:       time.Now(),
	}
	afid, err := pkg.ArtifactMgr.Create(suite.Context(), af)
	suite.Nil(err, fmt.Sprintf("Add artifact failed for %d", repositoryID))
	return afid
}

func (suite *MiddlewareTestSuite) addArtAcc(pid, repositoryID int64, repositoryName, dgt, accdgt string) int64 {
	subaf := &artifact.Artifact{
		Type:           "Docker-Image",
		ProjectID:      pid,
		RepositoryID:   repositoryID,
		RepositoryName: repositoryName,
		Digest:         dgt,
		Size:           1024,
		PushTime:       time.Now(),
		PullTime:       time.Now(),
	}
	_, err := pkg.ArtifactMgr.Create(suite.Context(), subaf)
	suite.Nil(err, fmt.Sprintf("Add artifact failed for %d", repositoryID))

	af := &artifact.Artifact{
		Type:           "Nydus",
		ProjectID:      pid,
		RepositoryID:   repositoryID,
		RepositoryName: repositoryName,
		Digest:         accdgt,
		Size:           1024,
		PushTime:       time.Now(),
		PullTime:       time.Now(),
	}
	afid, err := pkg.ArtifactMgr.Create(suite.Context(), af)
	suite.Nil(err, fmt.Sprintf("Add artifact failed for %d", repositoryID))

	accid, err := accessory.Mgr.Create(suite.Context(), accessorymodel.AccessoryData{
		ID:                1,
		ArtifactID:        afid,
		SubArtifactDigest: subaf.Digest,
		Digest:            accdgt,
		Type:              accessorymodel.TypeNydusAccelerator,
	})
	suite.Nil(err, fmt.Sprintf("Add artifact accesspry failed for %d", repositoryID))
	return accid
}

func (suite *MiddlewareTestSuite) TestNydusAccelerator() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		subArtDigest := "sha256:f54a58bc1aac5ea1a25d796ae155dc228b3f0e11d046ae276b39c4bf2f13d8c4"
		_, descriptor, req := suite.prepare(name, subArtDigest)

		// create sunjectArtifact repository
		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)
		suite.Nil(err)

		// add subject artifact
		suite.addArt(projectID, repoId, name, subArtDigest)

		// add nydus artifact
		artID := suite.addArt(projectID, repoId, name, descriptor.Digest.String())
		suite.Nil(err)

		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		AcceleratorMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		accs, _ := accessory.Mgr.List(suite.Context(), &q.Query{
			Keywords: map[string]interface{}{
				"SubjectArtifactDigest": subArtDigest,
			},
		})
		suite.Equal(1, len(accs))
		suite.Equal(subArtDigest, accs[0].GetData().SubArtifactDigest)
		suite.Equal(artID, accs[0].GetData().ArtifactID)
		suite.True(accs[0].IsHard())
		suite.Equal(model.TypeNydusAccelerator, accs[0].GetData().Type)
	})
}

func (suite *MiddlewareTestSuite) TestNydusAcceleratorDup() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		subArtDigest := "sha256:f54a58bc1aac5ea1a25d796ae155dc228b3f0e11d046ae276b39c4bf2f13d8c4"
		_, descriptor, req := suite.prepare(name, subArtDigest)

		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)
		suite.Nil(err)
		accID := suite.addArtAcc(projectID, repoId, name, subArtDigest, descriptor.Digest.String())

		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		AcceleratorMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		accs, _ := accessory.Mgr.List(suite.Context(), &q.Query{
			Keywords: map[string]interface{}{
				"ID": accID,
			},
		})
		suite.Equal(1, len(accs))
		suite.Equal(descriptor.Digest.String(), accs[0].GetData().Digest)
		suite.True(accs[0].IsHard())
		suite.Equal(model.TypeNydusAccelerator, accs[0].GetData().Type)
	})
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}
