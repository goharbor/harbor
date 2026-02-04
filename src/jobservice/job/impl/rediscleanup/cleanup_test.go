package rediscleanup

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/rediscleanup"
)

type RedisCleanupSuite struct {
	suite.Suite
	redisCleanupMgr *rediscleanup.Manager
	job             *Cleanup
}

func (suite *RedisCleanupSuite) SetupTest() {
	suite.redisCleanupMgr = &rediscleanup.Manager{}
	suite.job = &Cleanup{redisCleanupManager: suite.redisCleanupMgr}
}

func (suite *RedisCleanupSuite) TestRun() {
	mock.OnAnything(suite.redisCleanupMgr, "CleanupInvalidBlobSizeKeys").Return(nil)
	params := job.Parameters{}
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.NoError(err)
	// assert that job manager is invoked in this mode
	suite.redisCleanupMgr.AssertCalled(suite.T(), "CleanupInvalidBlobSizeKeys", mock.Anything)
}

func (suite *RedisCleanupSuite) TestRunFailure() {
	mock.OnAnything(suite.redisCleanupMgr, "CleanupInvalidBlobSizeKeys").Return(errors.New("test error"))
	params := job.Parameters{}
	ctx := &mockjobservice.MockJobContext{}

	err := suite.job.Run(ctx, params)
	suite.Error(err)
	// assert that job manager is invoked in this mode
	suite.redisCleanupMgr.AssertCalled(suite.T(), "CleanupInvalidBlobSizeKeys", mock.Anything)
}

func (suite *RedisCleanupSuite) TestMaxFails() {
	suite.Equal(uint(1), suite.job.MaxFails())
}

func (suite *RedisCleanupSuite) TestMaxConcurrency() {
	suite.Equal(uint(1), suite.job.MaxCurrency())
}

func (suite *RedisCleanupSuite) TestShouldRetry() {
	suite.Equal(true, suite.job.ShouldRetry())
}

func (suite *RedisCleanupSuite) TestValidate() {
	suite.NoError(suite.job.Validate(job.Parameters{}))
}

func TestRedisCleanupSuite(t *testing.T) {
	suite.Run(t, &RedisCleanupSuite{})
}
