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

package gc

import (
	"context"
	"testing"
	"time"

	"github.com/docker/distribution/manifest/schema2"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/accessory"
	accessorymodel "github.com/goharbor/harbor/src/pkg/accessory/model"
	_ "github.com/goharbor/harbor/src/pkg/accessory/model/cosign"
	pkgart "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/artifactrash"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	"github.com/goharbor/harbor/src/pkg/blob"
	blobModels "github.com/goharbor/harbor/src/pkg/blob/models"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/pkg/tag"
	tagmodel "github.com/goharbor/harbor/src/pkg/tag/model/tag"
	htesting "github.com/goharbor/harbor/src/testing"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
)

var errRollback = errors.New("rollback")

// dbJobContext runs the job against the database through the given context
type dbJobContext struct {
	*mockjobservice.MockJobContext
	ctx context.Context
}

func (c *dbJobContext) SystemContext() context.Context {
	return c.ctx
}

type parityTestSuite struct {
	htesting.Suite
	ctx      context.Context
	blobMgr  blob.Manager
	projects map[string]int64
	repos    map[string]int64
	// the digests of all blobs the fixture created
	created map[string]bool
}

// TestDryRunMatchesRealRun runs the mark phase as a dry-run and for real on the same data and expects both to pick the
// same blobs. Everything happens in one transaction that is rolled back.
func (suite *parityTestSuite) TestDryRunMatchesRealRun() {
	err := orm.WithTransaction(func(ctx context.Context) error {
		suite.ctx = ctx
		suite.blobMgr = blob.NewManager()
		suite.projects = map[string]int64{}
		suite.repos = map[string]int64{}
		suite.created = map[string]bool{}
		expected := suite.fixture()

		dryRun := suite.mark(true)
		realRun := suite.mark(false)
		suite.ElementsMatch(expected, dryRun)
		suite.ElementsMatch(realRun, dryRun)
		return errRollback
	})(suite.Context())
	suite.ErrorIs(err, errRollback)
}

// fixture creates the artifacts and blobs, and returns the digests that GC is expected to delete
func (suite *parityTestSuite) fixture() []string {
	var expected []string
	deleted := func(img *image) {
		expected = append(expected, img.blobs...)
	}

	// the same image in two projects, untagged in one of them, sharing its base layer with a tagged image
	base := suite.blob(true)
	x := suite.image("a", "app", nil, base)
	suite.copyImage("b", "app", x, "1.0")
	suite.image("a", "other", []string{"1.0"}, base)
	deleted(suite.image("a", "junk", nil))

	// an untagged index with a tagged child, a child of another tagged index and a child of nothing else
	c := suite.image("a", "idx", []string{"c"})
	d := suite.image("a", "idx", nil)
	e := suite.image("a", "idx", nil)
	deleted(suite.index("a", "idx", nil, c, d, e))
	suite.index("a", "idx", []string{"i2"}, e)
	deleted(d)

	// a child shared by two untagged indexes is deleted along with the second one
	f := suite.image("a", "idx2", nil)
	deleted(suite.index("a", "idx2", nil, f))
	deleted(suite.index("a", "idx2", nil, f))
	deleted(f)

	// an untagged image with a signature
	w := suite.image("a", "signed", nil)
	deleted(w)
	deleted(suite.accessory(w))

	// a tagged image with a signature
	suite.accessory(suite.image("a", "signed", []string{"1.0"}))

	// blobs pushed inside the time window
	suite.imageWithBlobs("a", "recent", nil, suite.blob(false), suite.blob(false), suite.blob(false))

	// orphan blobs of an aborted push, one of them also used in another project
	orphan := suite.blob(true)
	suite.associate(orphan, "a")
	expected = append(expected, orphan)
	shared := suite.blob(true)
	suite.associate(shared, "a")
	suite.image("b", "keep", []string{"1.0"}, shared)

	// a blob that a previous GC left in status deleting
	deleting := suite.blob(true)
	suite.exec(`UPDATE blob SET status = ? WHERE digest = ?`, blobModels.StatusDeleting, deleting)

	return expected
}

type image struct {
	id      int64
	project string
	repo    string
	digest  string
	// the manifest and all the blobs only this image uses
	blobs []string
}

func (suite *parityTestSuite) project(name string) int64 {
	if id, ok := suite.projects[name]; ok {
		return id
	}
	id, err := suite.AddProject(suite.ctx, &proModels.Project{Name: "gc-parity-" + name + suite.RandString(6), OwnerID: 1})
	suite.Require().NoError(err)
	suite.projects[name] = id
	return id
}

func (suite *parityTestSuite) repository(project, name string) (int64, string) {
	repoName := project + "/" + name
	if id, ok := suite.repos[repoName]; ok {
		return id, repoName
	}
	id, err := pkg.RepositoryMgr.Create(suite.ctx, &repomodel.RepoRecord{Name: repoName, ProjectID: suite.project(project)})
	suite.Require().NoError(err)
	suite.repos[repoName] = id
	return id, repoName
}

func (suite *parityTestSuite) blob(old bool) string {
	digest := suite.DigestString()
	_, err := suite.blobMgr.Create(suite.ctx, digest, schema2.MediaTypeLayer, 1024)
	suite.Require().NoError(err)
	suite.created[digest] = true
	if old {
		suite.exec(`UPDATE blob SET update_time = now() - interval '3 hours' WHERE digest = ?`, digest)
	}
	return digest
}

func (suite *parityTestSuite) associate(digest, project string) {
	b, err := suite.blobMgr.Get(suite.ctx, digest)
	suite.Require().NoError(err)
	_, err = suite.blobMgr.AssociateWithProject(suite.ctx, b.ID, suite.project(project))
	suite.Require().NoError(err)
}

// image creates an image with its own config and layer on top of the shared layers
func (suite *parityTestSuite) image(project, repo string, tags []string, sharedLayers ...string) *image {
	return suite.imageWithBlobs(project, repo, tags, suite.blob(true), suite.blob(true), suite.blob(true), sharedLayers...)
}

func (suite *parityTestSuite) imageWithBlobs(project, repo string, tags []string, manifest, config, layer string, sharedLayers ...string) *image {
	img := &image{project: project, repo: repo, digest: manifest, blobs: []string{manifest, config, layer}}
	img.id = suite.artifact(project, repo, schema2.MediaTypeManifest, manifest, tags, nil, append([]string{config, layer}, sharedLayers...)...)
	return img
}

func (suite *parityTestSuite) copyImage(project, repo string, img *image, tags ...string) {
	blobs, err := suite.blobMgr.GetByArt(suite.ctx, img.digest)
	suite.Require().NoError(err)
	var digests []string
	for _, b := range blobs {
		if b.Digest != img.digest {
			digests = append(digests, b.Digest)
		}
	}
	suite.artifact(project, repo, schema2.MediaTypeManifest, img.digest, tags, nil, digests...)
}

func (suite *parityTestSuite) index(project, repo string, tags []string, children ...*image) *image {
	manifest := suite.blob(true)
	var references []*pkgart.Reference
	var childManifests []string
	for _, child := range children {
		references = append(references, &pkgart.Reference{ChildID: child.id, ChildDigest: child.digest})
		childManifests = append(childManifests, child.digest)
	}
	id := suite.artifact(project, repo, v1.MediaTypeImageIndex, manifest, tags, references, childManifests...)
	return &image{id: id, project: project, repo: repo, digest: manifest, blobs: []string{manifest}}
}

func (suite *parityTestSuite) accessory(subject *image) *image {
	sig := suite.image(subject.project, subject.repo, nil)
	_, repoName := suite.repository(subject.project, subject.repo)
	_, err := accessory.Mgr.Create(suite.ctx, accessorymodel.AccessoryData{
		ArtifactID:        sig.id,
		SubArtifactID:     subject.id,
		SubArtifactRepo:   repoName,
		SubArtifactDigest: subject.digest,
		Type:              accessorymodel.TypeCosignSignature,
		Digest:            sig.digest,
		Size:              1024,
	})
	suite.Require().NoError(err)
	return sig
}

// artifact creates the artifact and associates its manifest and blobs with it and its project, the way a push does
func (suite *parityTestSuite) artifact(project, repo, mediaType, manifest string, tags []string, references []*pkgart.Reference, blobs ...string) int64 {
	repoID, repoName := suite.repository(project, repo)
	id, err := pkg.ArtifactMgr.Create(suite.ctx, &pkgart.Artifact{
		Type:              "IMAGE",
		MediaType:         schema2.MediaTypeImageConfig,
		ManifestMediaType: mediaType,
		ProjectID:         suite.project(project),
		RepositoryID:      repoID,
		RepositoryName:    repoName,
		Digest:            manifest,
		PushTime:          time.Now(),
		References:        references,
	})
	suite.Require().NoError(err)
	for _, digest := range append([]string{manifest}, blobs...) {
		_, err := suite.blobMgr.AssociateWithArtifact(suite.ctx, digest, manifest)
		suite.Require().NoError(err)
		suite.associate(digest, project)
	}
	for _, name := range tags {
		_, err := tag.Mgr.Create(suite.ctx, &tagmodel.Tag{RepositoryID: repoID, ArtifactID: id, Name: name, PushTime: time.Now()})
		suite.Require().NoError(err)
	}
	return id
}

func (suite *parityTestSuite) exec(sql string, args ...any) {
	o, err := orm.FromContext(suite.ctx)
	suite.Require().NoError(err)
	_, err = o.Raw(sql, args...).Exec()
	suite.Require().NoError(err)
}

// mark runs the mark phase and returns the digests of the fixture's blobs it picked
func (suite *parityTestSuite) mark(dryRun bool) []string {
	jobCtx := &dbJobContext{MockJobContext: &mockjobservice.MockJobContext{}, ctx: suite.ctx}
	jobCtx.On("GetLogger").Return(&mockjobservice.MockJobLogger{})
	jobCtx.On("OPCommand").Return(job.NilCommand, false)
	mock.OnAnything(jobCtx, "Checkin").Return(nil)

	gc := &GarbageCollector{
		artCtl:          artifact.Ctl,
		artMgr:          pkg.ArtifactMgr,
		artrashMgr:      artifactrash.NewManager(),
		blobMgr:         suite.blobMgr,
		logger:          &mockjobservice.MockJobLogger{},
		deleteUntagged:  true,
		dryRun:          dryRun,
		timeWindowHours: 2,
		trashedArts:     map[string][]model.ArtifactTrash{},
	}
	suite.Require().NoError(gc.mark(jobCtx))

	// other tests may have left useless blobs behind, they are picked by both runs alike
	var digests []string
	for _, b := range gc.deleteSet {
		if suite.created[b.Digest] {
			digests = append(digests, b.Digest)
		}
	}
	return digests
}

func TestParityTestSuite(t *testing.T) {
	suite.Run(t, &parityTestSuite{})
}
