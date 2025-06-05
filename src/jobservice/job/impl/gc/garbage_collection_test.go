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
	"testing"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/stretchr/testify/suite"

	commom_regctl "github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	pkgart "github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	pkg_blob "github.com/goharbor/harbor/src/pkg/blob/models"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	htesting "github.com/goharbor/harbor/src/testing"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	trashtesting "github.com/goharbor/harbor/src/testing/pkg/artifactrash"
	"github.com/goharbor/harbor/src/testing/pkg/blob"
	"github.com/goharbor/harbor/src/testing/registryctl"
)

type gcTestSuite struct {
	htesting.Suite
	artifactCtl       *artifacttesting.Controller
	artrashMgr        *trashtesting.Manager
	registryCtlClient *registryctl.Client
	projectCtl        *projecttesting.Controller
	blobMgr           *blob.Manager

	originalProjectCtl project.Controller

	regCtlInit func()
}

func (suite *gcTestSuite) SetupTest() {
	suite.artifactCtl = &artifacttesting.Controller{}
	suite.artrashMgr = &trashtesting.Manager{}
	suite.registryCtlClient = &registryctl.Client{}
	suite.blobMgr = &blob.Manager{}
	suite.projectCtl = &projecttesting.Controller{}

	suite.originalProjectCtl = project.Ctl
	project.Ctl = suite.projectCtl

	regCtlInit = func() { commom_regctl.RegistryCtlClient = suite.registryCtlClient }
}

func (suite *gcTestSuite) TearDownTest() {
	project.Ctl = suite.originalProjectCtl
}

func (suite *gcTestSuite) TestMaxFails() {
	gc := &GarbageCollector{}
	suite.Equal(uint(1), gc.MaxFails())
}

func (suite *gcTestSuite) TestShouldRetry() {
	gc := &GarbageCollector{}
	suite.False(gc.ShouldRetry())
}

func (suite *gcTestSuite) TestValidate() {
	gc := &GarbageCollector{}
	suite.Nil(gc.Validate(nil))
}

func (suite *gcTestSuite) TestDeletedArt() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			Artifact: pkgart.Artifact{
				ID:           1,
				RepositoryID: 1,
			},
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	mock.OnAnything(suite.artrashMgr, "Filter").Return([]model.ArtifactTrash{
		{
			ID:                1,
			Digest:            suite.DigestString(),
			ManifestMediaType: schema2.MediaTypeManifest,
		},
	}, nil)

	gc := &GarbageCollector{
		artCtl:     suite.artifactCtl,
		artrashMgr: suite.artrashMgr,
	}

	arts, err := gc.deletedArt(ctx)
	suite.Nil(err)
	suite.Equal(1, len(arts))
}

func (suite *gcTestSuite) TestRemoveUntaggedBlobs() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.NilCommand, false)

	mock.OnAnything(suite.projectCtl, "List").Return([]*proModels.Project{
		{
			ProjectID: 1234,
			Name:      "test GC",
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "List").Return([]*pkg_blob.Blob{
		{
			ID:     1234,
			Digest: "sha256:1234",
			Size:   1234,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "CleanupAssociationsForProject").Return(nil)

	gc := &GarbageCollector{
		blobMgr: suite.blobMgr,
	}

	suite.NotPanics(func() {
		gc.markOrSweepUntaggedBlobs(ctx)
	})
}

func (suite *gcTestSuite) TestInit() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	mock.OnAnything(ctx, "Get").Return("core url", true)
	ctx.On("GetLogger").Return(logger)

	gc := &GarbageCollector{
		registryCtlClient: suite.registryCtlClient,
	}
	params := map[string]any{
		"delete_untagged": true,
		"redis_url_reg":   "redis url",
		"time_window":     1,
		"workers":         float64(3),
	}

	mock.OnAnything(gc.registryCtlClient, "Health").Return(nil)
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)
	suite.Equal(3, gc.workers)

	params = map[string]any{
		"delete_untagged": "unsupported",
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)

	params = map[string]any{
		"delete_untagged": false,
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.False(gc.deleteUntagged)

	params = map[string]any{
		"redis_url_reg": "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)
}

func (suite *gcTestSuite) TestStop() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	mock.OnAnything(ctx, "Get").Return("core url", true)
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.StopCommand, true)

	mock.OnAnything(suite.artifactCtl, "List").Return([]*artifact.Artifact{
		{
			Artifact: pkgart.Artifact{
				ID:           1,
				RepositoryID: 1,
			},
		},
	}, nil)

	gc := &GarbageCollector{
		registryCtlClient: suite.registryCtlClient,
		artCtl:            suite.artifactCtl,
		deleteUntagged:    true,
	}

	suite.Equal(errGcStop, gc.mark(ctx))
}

func (suite *gcTestSuite) TestRun() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.NilCommand, true)
	mock.OnAnything(ctx, "Get").Return("core url", true)
	mock.OnAnything(ctx, "Checkin").Return(nil)

	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			Artifact: pkgart.Artifact{
				ID:           1,
				RepositoryID: 1,
			},
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	mock.OnAnything(suite.artrashMgr, "Filter").Return([]model.ArtifactTrash{}, nil)

	mock.OnAnything(suite.projectCtl, "List").Return([]*proModels.Project{
		{
			ProjectID: 12345,
			Name:      "test GC",
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "List").Return([]*pkg_blob.Blob{
		{
			ID:     12345,
			Digest: "sha256:12345",
			Size:   12345,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "CleanupAssociationsForProject").Return(nil)

	mock.OnAnything(suite.blobMgr, "UselessBlobs").Return([]*pkg_blob.Blob{
		{
			ID:          1,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeManifest,
		},
		{
			ID:          2,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeLayer,
		},
		{
			ID:          3,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeManifest,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "UpdateBlobStatus").Return(int64(1), nil)

	mock.OnAnything(suite.blobMgr, "Delete").Return(nil)

	mock.OnAnything(suite.registryCtlClient, "Health").Return(nil)

	gc := &GarbageCollector{
		artCtl:            suite.artifactCtl,
		artrashMgr:        suite.artrashMgr,
		blobMgr:           suite.blobMgr,
		registryCtlClient: suite.registryCtlClient,
	}
	params := map[string]any{
		"delete_untagged": false,
		"redis_url_reg":   tests.GetRedisURL(),
		"time_window":     1,
		"workers":         3,
	}

	mock.OnAnything(gc.registryCtlClient, "DeleteBlob").Return(nil)
	suite.Nil(gc.Run(ctx, params))
}

func (suite *gcTestSuite) TestMark() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.NilCommand, false)

	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			Artifact: pkgart.Artifact{
				ID:           1,
				RepositoryID: 1,
			},
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	mock.OnAnything(suite.artrashMgr, "Filter").Return([]model.ArtifactTrash{
		{
			ID:                1,
			Digest:            suite.DigestString(),
			ManifestMediaType: schema2.MediaTypeManifest,
		},
	}, nil)

	mock.OnAnything(suite.projectCtl, "List").Return([]*proModels.Project{
		{
			ProjectID: 1234,
			Name:      "test GC",
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "List").Return([]*pkg_blob.Blob{
		{
			ID:     1234,
			Digest: suite.DigestString(),
			Size:   1234,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "CleanupAssociationsForProject").Return(nil)

	mock.OnAnything(suite.blobMgr, "UselessBlobs").Return([]*pkg_blob.Blob{
		{
			ID:          1,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeManifest,
		},
		{
			ID:          2,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeLayer,
		},
		{
			ID:          3,
			Digest:      suite.DigestString(),
			ContentType: schema2.MediaTypeManifest,
		},
	}, nil)

	mock.OnAnything(suite.blobMgr, "UpdateBlobStatus").Return(int64(1), nil)

	gc := &GarbageCollector{
		artCtl:     suite.artifactCtl,
		artrashMgr: suite.artrashMgr,
		blobMgr:    suite.blobMgr,
	}

	suite.Nil(gc.mark(ctx))
}

func (suite *gcTestSuite) TestSweep() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	ctx.On("OPCommand").Return(job.NilCommand, false)
	mock.OnAnything(ctx, "Checkin").Return(nil)

	mock.OnAnything(suite.blobMgr, "UpdateBlobStatus").Return(int64(1), nil)
	mock.OnAnything(suite.blobMgr, "Delete").Return(nil)

	gc := &GarbageCollector{
		artCtl:            suite.artifactCtl,
		artrashMgr:        suite.artrashMgr,
		blobMgr:           suite.blobMgr,
		registryCtlClient: suite.registryCtlClient,
		deleteSet: []*pkg_blob.Blob{
			{
				ID:          1,
				Digest:      suite.DigestString(),
				ContentType: schema2.MediaTypeLayer,
			},
		},
		workers: 3,
	}

	mock.OnAnything(gc.registryCtlClient, "DeleteBlob").Return(nil)
	suite.Nil(gc.sweep(ctx))
}

func (suite *gcTestSuite) TestSaveRes() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	mock.OnAnything(ctx, "Checkin").Return(nil)
	suite.Nil(saveGCRes(ctx, 123456, 100, 100))
}

func TestGCTestSuite(t *testing.T) {
	t.Setenv("UTTEST", "true")
	suite.Run(t, &gcTestSuite{})
}
