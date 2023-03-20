package subject

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
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/subject"
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

func (suite *MiddlewareTestSuite) prepare(name, subject string) (distribution.Manifest, distribution.Descriptor, *http.Request) {
	body := fmt.Sprintf(`
	{
   "schemaVersion":2,
   "mediaType":"application/vnd.oci.image.manifest.v1+json",
   "config":{
      "mediaType":"application/vnd.example.sbom",
      "size":2,
      "digest":"sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
   },
   "layers":[
      {
         "mediaType":"application/vnd.example.sbom.text",
         "size":37,
         "digest":"sha256:45592a729ef6884ea3297e9510d79104f27aeef5f4919b3a921e3abb7f469709"
      }
   ],
   "annotations":{
      "org.example.sbom.format":"text"
   },
   "subject":{
      "mediaType":"application/vnd.oci.image.manifest.v1+json",
      "size":419,
      "digest":"%s"
   }}`, subject)

	manifest, descriptor, err := distribution.UnmarshalManifest("application/vnd.oci.image.manifest.v1+json", []byte(body))
	suite.Nil(err)

	req := suite.NewRequest(http.MethodPut, fmt.Sprintf("/v2/%s/manifests/%s", name, descriptor.Digest.String()), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/vnd.oci.image.manifest.v1+json")
	info := lib.ArtifactInfo{
		Repository: name,
		Reference:  descriptor.Digest.String(),
		Tag:        descriptor.Digest.String(),
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
		Type:           "Subject",
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
		Type:              accessorymodel.TypeSubject,
	})
	suite.Nil(err, fmt.Sprintf("Add artifact accesspry failed for %d", repositoryID))
	return accid
}

func (suite *MiddlewareTestSuite) TestSubject() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)

		subArtDigest := suite.DigestString()
		suite.addArt(projectID, repoId, name, subArtDigest)

		_, descriptor, req := suite.prepare(name, subArtDigest)
		suite.Nil(err)
		artID := suite.addArt(projectID, repoId, name, descriptor.Digest.String())
		suite.Nil(err)

		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		Middleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		accs, err := accessory.Mgr.List(suite.Context(), &q.Query{
			Keywords: map[string]interface{}{
				"SubjectArtifactDigest": subArtDigest,
			},
		})
		suite.Equal(1, len(accs))
		suite.Equal(subArtDigest, accs[0].GetData().SubArtifactDigest)
		suite.Equal(artID, accs[0].GetData().ArtifactID)
		suite.True(accs[0].IsHard())
		suite.Equal(model.TypeSubject, accs[0].GetData().Type)
	})
}

func (suite *MiddlewareTestSuite) TestSubjectDup() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)

		subArtDigest := suite.DigestString()
		_, descriptor, req := suite.prepare(name, subArtDigest)
		suite.Nil(err)

		accID := suite.addArtAcc(projectID, repoId, name, subArtDigest, descriptor.Digest.String())

		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": descriptor.Digest.String()})
		Middleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		accs, err := accessory.Mgr.List(suite.Context(), &q.Query{
			Keywords: map[string]interface{}{
				"ID": accID,
			},
		})
		suite.Equal(1, len(accs))
		suite.Equal(descriptor.Digest.String(), accs[0].GetData().Digest)
		suite.True(accs[0].IsHard())
		suite.Equal(model.TypeSubject, accs[0].GetData().Type)
	})
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}
