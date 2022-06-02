package cosign

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/distribution"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/stretchr/testify/suite"
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
		"schemaVersion":2,
		"config":{
			"mediaType":"application/vnd.oci.image.manifest.v1+json",
			"size":233,
			"digest":"sha256:d4e6059ece7bea95266fd7766353130d4bf3dc21048b8a9783c98b8412618c38"
		},
		"layers":[
			{
				"mediaType":"application/vnd.dev.cosign.simplesigning.v1+json",
				"size":250,
				"digest":"sha256:91a821a0e2412f1b99b07bfe176451bcc343568b761388718abbf38076048564",
				"annotations":{
					"dev.cosignproject.cosign/signature":"MEUCIQD/imXjZJlcV82eXu9y9FJGgbDwVPw7AaGFzqva8G+CgwIgYc4CRvEjwoAwkzGoX+aZxQWCASpv5G+EAWDKOJRLbTQ="
				}
			}
		]
	}`)

	manifest, descriptor, err := distribution.UnmarshalManifest("application/vnd.oci.image.manifest.v1+json", []byte(body))
	suite.Nil(err)

	req := suite.NewRequest(http.MethodPut, fmt.Sprintf("/v2/%s/manifests/%s", name, ref), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	info := lib.ArtifactInfo{
		Repository: name,
		Reference:  ref,
		Tag:        ref,
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
	subafid, err := pkg.ArtifactMgr.Create(suite.Context(), subaf)
	suite.Nil(err, fmt.Sprintf("Add artifact failed for %d", repositoryID))

	af := &artifact.Artifact{
		Type:           "Cosign",
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
		ID:            1,
		ArtifactID:    afid,
		SubArtifactID: subafid,
		Digest:        accdgt,
		Type:          accessorymodel.TypeCosignSignature,
	})
	suite.Nil(err, fmt.Sprintf("Add artifact accesspry failed for %d", repositoryID))
	return accid
}

func (suite *MiddlewareTestSuite) TestCosignSignature() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		subArtDigest := suite.DigestString()
		ref := fmt.Sprintf("%s.sig", strings.ReplaceAll(subArtDigest, "sha256:", "sha256-"))
		_, descriptor, req := suite.prepare(name, ref)

		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)
		suite.Nil(err)
		subjectArtID := suite.addArt(projectID, repoId, name, subArtDigest)
		artID := suite.addArt(projectID, repoId, name, descriptor.Digest.String())
		suite.Nil(err)

		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		CosignSignatureMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		accs, err := accessory.Mgr.List(suite.Context(), &q.Query{
			Keywords: map[string]interface{}{
				"SubjectArtifactID": subjectArtID,
			},
		})
		suite.Equal(1, len(accs))
		suite.Equal(subjectArtID, accs[0].GetData().SubArtifactID)
		suite.Equal(artID, accs[0].GetData().ArtifactID)
		suite.True(accs[0].IsHard())
		suite.Equal(model.TypeCosignSignature, accs[0].GetData().Type)
	})
}

func (suite *MiddlewareTestSuite) TestCosignSignatureDup() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		subArtDigest := suite.DigestString()
		ref := fmt.Sprintf("%s.sig", strings.ReplaceAll(subArtDigest, "sha256:", "sha256-"))
		_, descriptor, req := suite.prepare(name, ref)

		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)
		suite.Nil(err)
		accID := suite.addArtAcc(projectID, repoId, name, subArtDigest, descriptor.Digest.String())

		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		CosignSignatureMiddleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		accs, err := accessory.Mgr.List(suite.Context(), &q.Query{
			Keywords: map[string]interface{}{
				"ID": accID,
			},
		})
		suite.Equal(1, len(accs))
		suite.Equal(descriptor.Digest.String(), accs[0].GetData().Digest)
		suite.True(accs[0].IsHard())
		suite.Equal(model.TypeCosignSignature, accs[0].GetData().Type)
	})
}

func (suite *MiddlewareTestSuite) TestMatchManifestURLPattern() {
	_, _, ok := matchCosignSignaturePattern("/v2/library/hello-world/manifests/.Invalid")
	suite.False(ok)

	_, _, ok = matchCosignSignaturePattern("/v2/")
	suite.False(ok)

	_, _, ok = matchCosignSignaturePattern("/v2/library/hello-world/manifests//")
	suite.False(ok)

	_, _, ok = matchCosignSignaturePattern("/v2/library/hello-world/manifests/###")
	suite.False(ok)

	repository, _, ok := matchCosignSignaturePattern("/v2/library/hello-world/manifests/latest")
	suite.False(ok)

	_, _, ok = matchCosignSignaturePattern("/v2/library/hello-world/manifests/sha256:e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9")
	suite.False(ok)

	repository, reference, ok := matchCosignSignaturePattern("/v2/library/hello-world/manifests/sha256-e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9.sig")
	suite.True(ok)
	suite.Equal("library/hello-world", repository)
	suite.Equal("e5785cb0c62cebbed4965129bae371f0589cadd6d84798fb58c2c5f9e237efd9", reference)
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}
