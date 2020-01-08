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
package runner

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger/backend"

	"github.com/gocraft/work"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/tests"

	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// RedisRunnerTestSuite tests functions of redis runner
type RedisRunnerTestSuite struct {
	suite.Suite

	lcmCtl lcm.Controller

	envContext *env.Context

	cancel    context.CancelFunc
	namespace string
	pool      *redis.Pool
}

// TestRedisRunnerTestSuite is entry of go test
func TestRedisRunnerTestSuite(t *testing.T) {
	suite.Run(t, new(RedisRunnerTestSuite))
}

// SetupSuite prepares test suite
func (suite *RedisRunnerTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.Background())
	suite.cancel = cancel

	suite.envContext = &env.Context{
		SystemContext: ctx,
		WG:            new(sync.WaitGroup),
		ErrorChan:     make(chan error, 1),
	}

	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()

	suite.lcmCtl = lcm.NewController(
		suite.envContext,
		suite.namespace,
		suite.pool,
		func(hookURL string, change *job.StatusChange) error { return nil },
	)

	fakeStats := &job.Stats{
		Info: &job.StatsInfo{
			JobID:    "FAKE-j",
			JobName:  "fakeParentJob",
			JobKind:  job.KindGeneric,
			Status:   job.PendingStatus.String(),
			IsUnique: false,
		},
	}
	_, err := suite.lcmCtl.New(fakeStats)
	require.NoError(suite.T(), err, "lcm new: nil error expected but got %s", err)
}

// SetupTest prepares test cases
func (suite *RedisRunnerTestSuite) SetupTest() {
	t, err := suite.lcmCtl.Track("FAKE-j")
	require.NoError(suite.T(), err)
	err = t.Update("status", job.PendingStatus.String()) // reset
	assert.NoError(suite.T(), err)
}

// TearDownSuite clears the test suite
func (suite *RedisRunnerTestSuite) TearDownSuite() {
	suite.cancel()
}

// TestJobWrapper tests the redis job wrapper
func (suite *RedisRunnerTestSuite) TestJobWrapper() {
	j := &work.Job{
		ID:         "FAKE-j",
		Name:       "fakeParentJob",
		EnqueuedAt: time.Now().Add(5 * time.Minute).Unix(),
	}

	oldJobLoggerCfg := config.DefaultConfig.JobLoggerConfigs
	defer func() {
		config.DefaultConfig.JobLoggerConfigs = oldJobLoggerCfg
	}()

	config.DefaultConfig.JobLoggerConfigs = []*config.LoggerConfig{
		{
			Name:  "STD_OUTPUT",
			Level: "DEBUG",
			Settings: map[string]interface{}{
				"output": backend.StdErr,
			},
		},
		{
			Name:  "FILE",
			Level: "ERROR",
			Settings: map[string]interface{}{
				"base_dir": os.TempDir(),
			},
			Sweeper: &config.LogSweeperConfig{
				Duration: 5,
				Settings: map[string]interface{}{
					"work_dir": os.TempDir(),
				},
			},
		},
	}

	redisJob := NewRedisJob((*fakeParentJob)(nil), suite.envContext, suite.lcmCtl)
	err := redisJob.Run(j)
	require.NoError(suite.T(), err, "redis job: nil error expected but got %s", err)
}

// TestJobWrapperInvalidTracker tests job runner with invalid job ID
func (suite *RedisRunnerTestSuite) TestJobWrapperInvalidTracker() {
	j := &work.Job{
		ID:         "FAKE-j2",
		Name:       "fakeParentJob",
		EnqueuedAt: time.Now().Add(5 * time.Minute).Unix(),
		Fails:      3,
	}

	redisJob := NewRedisJob((*fakeParentJob)(nil), suite.envContext, suite.lcmCtl)
	err := redisJob.Run(j)
	require.Error(suite.T(), err, "redis job: non nil error expected but got nil")
	assert.Equal(suite.T(), int64(10000000000), j.Fails)
}

// TestJobWrapperPanic tests job runner panic
func (suite *RedisRunnerTestSuite) TestJobWrapperPanic() {
	j := &work.Job{
		ID:         "FAKE-j",
		Name:       "fakePanicJob",
		EnqueuedAt: time.Now().Add(5 * time.Minute).Unix(),
	}

	redisJob := NewRedisJob((*fakePanicJob)(nil), suite.envContext, suite.lcmCtl)
	err := redisJob.Run(j)
	assert.Error(suite.T(), err)
}

// TestJobWrapperStopped tests job runner stopped
func (suite *RedisRunnerTestSuite) TestJobWrapperStopped() {
	j := &work.Job{
		ID:         "FAKE-j",
		Name:       "fakePanicJob",
		EnqueuedAt: time.Now().Add(5 * time.Minute).Unix(),
	}

	t, err := suite.lcmCtl.Track("FAKE-j")
	require.NoError(suite.T(), err)
	err = t.Stop()
	require.NoError(suite.T(), err)

	redisJob := NewRedisJob((*fakeParentJob)(nil), suite.envContext, suite.lcmCtl)
	err = redisJob.Run(j)
	require.NoError(suite.T(), err)
}

type fakeParentJob struct {
}

func (j *fakeParentJob) MaxFails() uint {
	return 1
}

func (j *fakeParentJob) ShouldRetry() bool {
	return false
}

func (j *fakeParentJob) Validate(params job.Parameters) error {
	return nil
}

func (j *fakeParentJob) Run(ctx job.Context, params job.Parameters) error {
	_ = ctx.Checkin("start")
	ctx.OPCommand()
	return nil
}

type fakePanicJob struct {
}

func (j *fakePanicJob) MaxFails() uint {
	return 1
}

func (j *fakePanicJob) ShouldRetry() bool {
	return false
}

func (j *fakePanicJob) Validate(params job.Parameters) error {
	return nil
}

func (j *fakePanicJob) Run(ctx job.Context, params job.Parameters) error {
	panic("for testing")
}
