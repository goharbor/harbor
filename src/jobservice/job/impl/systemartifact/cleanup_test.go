package systemartifact

import (
	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/systemartifact"
	"github.com/stretchr/testify/suite"
	"testing"
)

type SystemArtifactCleanupSuite struct {
	suite.Suite
	sysArtifactMgr *systemartifact.Manager
	job            *Cleanup
}

func (suite *SystemArtifactCleanupSuite) SetupTest() {
	suite.sysArtifactMgr = &systemartifact.Manager{}
	suite.job = &Cleanup{sysArtifactManager: suite.sysArtifactMgr}
}

func (suite *SystemArtifactCleanupSuite) TestRun() {
	mock.OnAnything(suite.sysArtifactMgr, "Cleanup").Return(int64(100), int64(100), nil)
	params := job.Parameters{}
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.NoError(err)
	// assert that job manager is invoked in this mode
	suite.sysArtifactMgr.AssertCalled(suite.T(), "Cleanup", mock.Anything)
}

func TestSystemArtifactCleanupSuite(t *testing.T) {
	suite.Run(t, &SystemArtifactCleanupSuite{})
}
