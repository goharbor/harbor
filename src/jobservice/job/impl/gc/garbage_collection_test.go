package gc

import (
	"github.com/goharbor/harbor/src/common/config"
	commom_regctl "github.com/goharbor/harbor/src/common/registryctl"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/artifactrash/model"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	trashtesting "github.com/goharbor/harbor/src/testing/pkg/artifactrash"
	"github.com/goharbor/harbor/src/testing/registryctl"
	"github.com/stretchr/testify/suite"
	"testing"
)

type gcTestSuite struct {
	suite.Suite
	artifactCtl       *artifacttesting.Controller
	artrashMgr        *trashtesting.FakeManager
	registryCtlClient *registryctl.Mockclient

	regCtlInit  func()
	setReadOnly func(cfgMgr *config.CfgManager, switcher bool) error
	getReadOnly func(cfgMgr *config.CfgManager) (bool, error)
}

func (suite *gcTestSuite) SetupTest() {
	suite.artifactCtl = &artifacttesting.Controller{}
	suite.artrashMgr = &trashtesting.FakeManager{}
	suite.registryCtlClient = &registryctl.Mockclient{}

	regCtlInit = func() { commom_regctl.RegistryCtlClient = suite.registryCtlClient }
	setReadOnly = func(cfgMgr *config.CfgManager, switcher bool) error { return nil }
	getReadOnly = func(cfgMgr *config.CfgManager) (bool, error) { return true, nil }
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

func (suite *gcTestSuite) TestDeleteCandidates() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	suite.artrashMgr.On("Flush").Return(nil)
	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	suite.artrashMgr.On("Filter").Return([]model.ArtifactTrash{}, nil)

	gc := &GarbageCollector{
		artCtl:     suite.artifactCtl,
		artrashMgr: suite.artrashMgr,
	}
	suite.Nil(gc.deleteCandidates(ctx))
}

func (suite *gcTestSuite) TestInit() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	mock.OnAnything(ctx, "Get").Return("core url", true)
	ctx.On("GetLogger").Return(logger)

	gc := &GarbageCollector{}
	params := map[string]interface{}{
		"delete_untagged": true,
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)

	params = map[string]interface{}{
		"delete_untagged": "unsupported",
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)

	params = map[string]interface{}{
		"delete_untagged": false,
		"redis_url_reg":   "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.False(gc.deleteUntagged)

	params = map[string]interface{}{
		"redis_url_reg": "redis url",
	}
	suite.Nil(gc.init(ctx, params))
	suite.True(gc.deleteUntagged)
}

func (suite *gcTestSuite) TestRun() {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)
	mock.OnAnything(ctx, "Get").Return("core url", true)

	suite.artrashMgr.On("Flush").Return(nil)
	suite.artifactCtl.On("List").Return([]*artifact.Artifact{
		{
			ID:           1,
			RepositoryID: 1,
		},
	}, nil)
	suite.artifactCtl.On("Delete").Return(nil)
	suite.artrashMgr.On("Filter").Return([]model.ArtifactTrash{}, nil)

	gc := &GarbageCollector{
		artCtl:     suite.artifactCtl,
		artrashMgr: suite.artrashMgr,
		cfgMgr:     config.NewInMemoryManager(),
	}
	params := map[string]interface{}{
		"delete_untagged": false,
		// ToDo add a redis testing pkg, we do have a 'localhost' redis server in UT
		"redis_url_reg": "redis://localhost:6379",
	}

	suite.Nil(gc.Run(ctx, params))
}

func TestGCTestSuite(t *testing.T) {
	suite.Run(t, &gcTestSuite{})
}
