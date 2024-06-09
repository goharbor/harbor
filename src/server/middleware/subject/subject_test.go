package subject

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/docker/distribution/manifest/schema2"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/accessory"
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

func (suite *MiddlewareTestSuite) prepare(name, digset string, withoutSub ...bool) (distribution.Manifest, distribution.Descriptor, *http.Request) {
	body := fmt.Sprintf(`
	{
   "schemaVersion":2,
   "mediaType":"application/vnd.oci.image.manifest.v1+json",
   "config":{
      "mediaType":"application/vnd.example.main",
      "size":2,
      "digest":"sha256:44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
   },
   "layers":[
      {
         "mediaType":"application/vnd.example.main.text",
         "size":37,
         "digest":"sha256:45592a729ef6884ea3297e9510d79104f27aeef5f4919b3a921e3abb7f469709"
      }
   ],
   "annotations":{
      "org.example.main.format":"text"
   },
   "subject":{
      "mediaType":"application/vnd.oci.image.manifest.v1+json",
      "size":419,
      "digest":"%s"
   }}`, digset)

	if len(withoutSub) > 0 && withoutSub[0] {
		body = fmt.Sprintf(`
	{
   "schemaVersion":2,
   "mediaType":"application/vnd.oci.image.manifest.v1+json",
   "config":{
      "mediaType":"application/vnd.example.main",
      "size":2,
      "digest":"%s"
   },
   "layers":[
      {
         "mediaType":"application/vnd.example.main.text",
         "size":37,
         "digest":"sha256:45592a729ef6884ea3297e9510d79104f27aeef5f4919b3a921e3abb7f469709"
      }
   ],
   "annotations":{
      "org.example.main.format":"text"
   }}`, digset)
	}

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

func (suite *MiddlewareTestSuite) addArtAcc(pid, repositoryID int64, repositoryName, dgt, accdgt string, createSub ...bool) int64 {
	if len(createSub) > 0 && createSub[0] {
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
	}

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
		SubArtifactDigest: dgt,
		SubArtifactRepo:   repositoryName,
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
		suite.Equal(name, accs[0].GetData().SubArtifactRepo)
		suite.True(accs[0].IsHard())
		suite.Equal(accessorymodel.TypeSubject, accs[0].GetData().Type)
		suite.Equal(subArtDigest, res.Header().Values("OCI-Subject")[0])
	})
}

func (suite *MiddlewareTestSuite) TestSubjectAfterAcc() {
	// add acc, with subject digest
	// add subject
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)

		_, descriptor, req := suite.prepare(name, suite.DigestString(), true)
		suite.Nil(err)

		subArtDigest := descriptor.Digest.String()
		accArtDigest := suite.DigestString()

		accID := suite.addArtAcc(projectID, repoId, name, subArtDigest, accArtDigest)
		subArtID := suite.addArt(projectID, repoId, name, subArtDigest)

		res := httptest.NewRecorder()
		next := suite.NextHandler(http.StatusCreated, map[string]string{"Docker-Content-Digest": subArtDigest})
		Middleware()(next).ServeHTTP(res, req)
		suite.Equal(http.StatusCreated, res.Code)

		accs, err := accessory.Mgr.List(suite.Context(), &q.Query{
			Keywords: map[string]interface{}{
				"SubjectArtifactDigest": subArtDigest,
				"SubjectArtifactRepo":   name,
			},
		})
		suite.Equal(1, len(accs))
		suite.Equal(subArtDigest, accs[0].GetData().SubArtifactDigest)
		suite.Equal(subArtID, accs[0].GetData().SubArtifactID)
		suite.Equal(accID, accs[0].GetData().ID)
	})
}

func (suite *MiddlewareTestSuite) TestSubjectDup() {
	suite.WithProject(func(projectID int64, projectName string) {
		name := fmt.Sprintf("%s/hello-world", projectName)
		_, repoId, err := repository.Ctl.Ensure(suite.Context(), name)

		subArtDigest := suite.DigestString()
		_, descriptor, req := suite.prepare(name, subArtDigest)
		suite.Nil(err)

		accID := suite.addArtAcc(projectID, repoId, name, subArtDigest, descriptor.Digest.String(), true)

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
		suite.Equal(accessorymodel.TypeSubject, accs[0].GetData().Type)
	})
}

func (suite *MiddlewareTestSuite) TestIsNydusImage() {
	makeManifest := func(configType string) string {
		return fmt.Sprintf(`{
			"schemaVersion": 2,
			"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
			"config": {
				"mediaType": "%s",
				"digest": "sha256:e314d79415361272a5ff6919ce70eb1d82ae55641ff60dcd8286b731cae2b5e7",
				"size": 3322
			},
			"layers": [
				{
					"mediaType": "application/vnd.oci.image.layer.nydus.blob.v1",
					"digest": "sha256:bce0a563197a6aae0044f2063bf95f43bb956640b374fbdf0886cbc6926e2b7c",
					"size": 3440759,
					"annotations": {
						"containerd.io/snapshot/nydus-blob": "true"
					}
				},
				{
					"mediaType": "application/vnd.oci.image.layer.nydus.blob.v1",
					"digest": "sha256:7dedc3aaf7177a1d6792efcf1eae1305033fbac8dc48eb0caf49373b5d21475f",
					"size": 337049,
					"annotations": {
						"containerd.io/snapshot/nydus-blob": "true"
					}
				},
				{
					"mediaType": "application/vnd.oci.image.layer.nydus.blob.v1",
					"digest": "sha256:f6bf79efcfc89f657b9705ef9ed77659e413e355efac8c6d3eea49d908c9218a",
					"size": 5810244,
					"annotations": {
						"containerd.io/snapshot/nydus-blob": "true"
					}
				},
				{
					"mediaType": "application/vnd.oci.image.layer.nydus.blob.v1",
					"digest": "sha256:35c290e1471c2f546ba7ca8eb47b334c0234e6a2d2b274c54fe96e016c1913c7",
					"size": 7936,
					"annotations": {
						"containerd.io/snapshot/nydus-blob": "true"
					}
				},
				{
					"mediaType": "application/vnd.oci.image.layer.nydus.blob.v1",
					"digest": "sha256:1f168a347d1c654776644b331e631c3a1208699e2f608e29d8e3fd74e5fd99e8",
					"size": 7728,
					"annotations": {
						"containerd.io/snapshot/nydus-blob": "true"
					}
				},
				{
					"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
					"digest": "sha256:86211e9295fabea433b7186ddfa6fd31af048a2f6fe3cf8d747b6f7ea39c0ea6",
					"size": 35092,
					"annotations": {
						"containerd.io/snapshot/nydus-bootstrap": "true",
						"containerd.io/snapshot/nydus-fs-version": "6"
					}
				}
			],
			"subject": {
				"mediaType": "application/vnd.docker.distribution.manifest.v2+json",
				"digest": "sha256:f4d532d482a050a3bb02886be6d6deda9c22cf8df44b1465f04c8648ee573a70",
				"size": 1363
			}
		}`, configType)
	}
	manifest := &ocispec.Manifest{}
	err := json.Unmarshal([]byte(makeManifest(ocispec.MediaTypeImageConfig)), manifest)
	suite.Nil(err)
	suite.True(isNydusImage(manifest))

	err = json.Unmarshal([]byte(makeManifest(schema2.MediaTypeImageConfig)), manifest)
	suite.Nil(err)
	suite.True(isNydusImage(manifest))
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, &MiddlewareTestSuite{})
}
