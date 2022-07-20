package systemartifact

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/systemartifact"
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

func (suite *SystemArtifactCleanupSuite) TestRunFailure() {
	mock.OnAnything(suite.sysArtifactMgr, "Cleanup").Return(int64(0), int64(0), errors.New("test error"))
	params := job.Parameters{}
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.Error(err)
	// assert that job manager is invoked in this mode
	suite.sysArtifactMgr.AssertCalled(suite.T(), "Cleanup", mock.Anything)
}

func (suite *SystemArtifactCleanupSuite) TestMaxFails() {
	suite.Equal(uint(1), suite.job.MaxFails())
}

func (suite *SystemArtifactCleanupSuite) TestMaxConcurrency() {
	suite.Equal(uint(1), suite.job.MaxCurrency())
}

func (suite *SystemArtifactCleanupSuite) TestShouldRetry() {
	suite.Equal(true, suite.job.ShouldRetry())
}

func (suite *SystemArtifactCleanupSuite) TestValidate() {
	suite.NoError(suite.job.Validate(job.Parameters{}))
}

func TestSystemArtifactCleanupSuite(t *testing.T) {
	suite.Run(t, &SystemArtifactCleanupSuite{})
}
