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
package cworker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	common_dao "github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/job/impl"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/goharbor/harbor/src/jobservice/worker"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// CWorkerTestSuite tests functions of c worker
type CWorkerTestSuite struct {
	suite.Suite

	cWorker worker.Interface
	lcmCtl  lcm.Controller

	namespace string
	pool      *redis.Pool

	cancel  context.CancelFunc
	context *env.Context
}

// SetupSuite prepares test suite
func (suite *CWorkerTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()
	common_dao.PrepareTestForPostgresSQL()

	// Append node ID
	vCtx := context.WithValue(context.Background(), utils.NodeID, utils.GenerateNodeID())
	// Create the root context
	ctx, cancel := context.WithCancel(vCtx)
	suite.cancel = cancel

	envCtx := &env.Context{
		SystemContext: ctx,
		WG:            new(sync.WaitGroup),
		ErrorChan:     make(chan error, 1),
		JobContext:    impl.NewDefaultContext(ctx),
	}
	suite.context = envCtx

	suite.lcmCtl = lcm.NewController(
		envCtx,
		suite.namespace,
		suite.pool,
		func(hookURL string, change *job.StatusChange) error { return nil },
	)

	suite.cWorker = NewWorker(envCtx, suite.namespace, 5, suite.pool, suite.lcmCtl)
	err := suite.cWorker.RegisterJobs(map[string]interface{}{
		"fake_job":          (*fakeJob)(nil),
		"fake_long_run_job": (*fakeLongRunJob)(nil),
	})
	require.NoError(suite.T(), err, "register jobs: nil error expected but got %s", err)

	err = suite.cWorker.Start()
	require.NoError(suite.T(), err, "start redis worker: nil error expected but got %s", err)
}

// TearDownSuite clears the test suite
func (suite *CWorkerTestSuite) TearDownSuite() {
	suite.cancel()

	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_ = tests.ClearAll(suite.namespace, conn)
}

// TestCWorkerTestSuite is entry fo go test
func TestCWorkerTestSuite(t *testing.T) {
	suite.Run(t, new(CWorkerTestSuite))
}

// TestRegisterJobs ...
func (suite *CWorkerTestSuite) TestRegisterJobs() {
	_, ok := suite.cWorker.IsKnownJob("fake_job")
	assert.EqualValues(suite.T(), true, ok, "expected known job but registering 'fake_job' appears to have failed")

	params := make(map[string]interface{})
	params["name"] = "testing:v1"
	err := suite.cWorker.ValidateJobParameters((*fakeJob)(nil), params)
	assert.NoError(suite.T(), err, "validate parameters: nil error expected but got %s", err)
}

// TestEnqueueJob tests enqueue job
func (suite *CWorkerTestSuite) TestEnqueueJob() {
	params := make(job.Parameters)
	params["name"] = "testing:v1"

	stats, err := suite.cWorker.Enqueue("fake_job", params, false, "")
	require.NoError(suite.T(), err, "enqueue job: nil error expected but got %s", err)
	_, err = suite.lcmCtl.New(stats)
	assert.NoError(suite.T(), err, "lcm: nil error expected but got %s", err)
}

// TestEnqueueUniqueJob tests enqueue unique job
func (suite *CWorkerTestSuite) TestEnqueueUniqueJob() {
	params := make(job.Parameters)
	params["name"] = "testing:v2"

	stats, err := suite.cWorker.Enqueue("fake_job", params, true, "http://fake-hook.com:8080")
	require.NoError(suite.T(), err, "enqueue unique job: nil error expected but got %s", err)

	_, err = suite.lcmCtl.New(stats)
	assert.NoError(suite.T(), err, "lcm: nil error expected but got %s", err)
}

// TestScheduleJob tests schedule job
func (suite *CWorkerTestSuite) TestScheduleJob() {
	params := make(job.Parameters)
	params["name"] = "testing:v1"

	runAt := time.Now().Unix() + 1
	stats, err := suite.cWorker.Schedule("fake_job", params, 1, false, "")
	require.NoError(suite.T(), err, "schedule job: nil error expected but got %s", err)
	require.Condition(suite.T(), func() bool {
		return runAt <= stats.Info.RunAt
	}, "expect returned 'RunAt' should be >= '%d' but seems not", runAt)
	_, err = suite.lcmCtl.New(stats)
	assert.NoError(suite.T(), err, "lcm: nil error expected but got %s", err)
}

// TestEnqueuePeriodicJob tests periodic job
func (suite *CWorkerTestSuite) TestEnqueuePeriodicJob() {
	params := make(job.Parameters)
	params["name"] = "testing:v1"

	m := time.Now().Minute()
	if m+2 >= 60 {
		m = m - 2
	}
	_, err := suite.cWorker.PeriodicallyEnqueue(
		"fake_job",
		params,
		fmt.Sprintf("10 %d * * * *", m+2),
		false,
		"http://fake-hook.com:8080",
	)

	require.NoError(suite.T(), err, "periodic job: nil error expected but got %s", err)
}

// TestWorkerStats tests worker stats
func (suite *CWorkerTestSuite) TestWorkerStats() {
	stats, err := suite.cWorker.Stats()
	require.NoError(suite.T(), err, "worker stats: nil error expected but got %s", err)
	assert.Equal(suite.T(), 1, len(stats.Pools), "expected 1 pool but got 0")
}

// TestStopJob test stop job
func (suite *CWorkerTestSuite) TestStopJob() {
	// Stop generic job
	params := make(map[string]interface{})
	params["name"] = "testing:v1"

	genericJob, err := suite.cWorker.Enqueue("fake_long_run_job", params, false, "")
	require.NoError(suite.T(), err, "enqueue job: nil error expected but got %s", err)
	t, err := suite.lcmCtl.New(genericJob)
	require.NoError(suite.T(), err, "new job stats: nil error expected but got %s", err)

	tk := time.NewTicker(417 * time.Millisecond)
	defer tk.Stop()

LOOP:
	for {
		select {
		case <-tk.C:
			latest, err := t.Status()
			require.NoError(suite.T(), err, "get latest status: nil error expected but got %s", err)
			if latest.Compare(job.RunningStatus) >= 0 {
				break LOOP
			}
		case <-time.After(29 * time.Second):
			require.NoError(suite.T(), errors.New("check running status time out"))
			return
		}
	}

	err = suite.cWorker.StopJob(genericJob.Info.JobID)
	require.NoError(suite.T(), err, "stop job: nil error expected but got %s", err)

	// Stop scheduled job
	scheduledJob, err := suite.cWorker.Schedule("fake_long_run_job", params, 120, false, "")
	require.NoError(suite.T(), err, "schedule job: nil error expected but got %s", err)
	t, err = suite.lcmCtl.New(scheduledJob)
	require.NoError(suite.T(), err, "new job stats: nil error expected but got %s", err)

	err = suite.cWorker.StopJob(scheduledJob.Info.JobID)
	require.NoError(suite.T(), err, "stop job: nil error expected but got %s", err)
}

type fakeJob struct{}

func (j *fakeJob) MaxFails() uint {
	return 3
}

func (j *fakeJob) ShouldRetry() bool {
	return false
}

func (j *fakeJob) Validate(params job.Parameters) error {
	if p, ok := params["name"]; ok {
		if p == "testing:v1" || p == "testing:v2" {
			return nil
		}
	}

	return errors.New("validate: testing error")
}

func (j *fakeJob) Run(ctx job.Context, params job.Parameters) error {
	ctx.OPCommand()
	_ = ctx.Checkin("done")

	return nil
}

type fakeLongRunJob struct{}

func (j *fakeLongRunJob) MaxFails() uint {
	return 3
}

func (j *fakeLongRunJob) ShouldRetry() bool {
	return false
}

func (j *fakeLongRunJob) Validate(params job.Parameters) error {
	if p, ok := params["name"]; ok {
		if p == "testing:v1" || p == "testing:v2" {
			return nil
		}
	}

	return errors.New("validate: testing error")
}

func (j *fakeLongRunJob) Run(ctx job.Context, params job.Parameters) error {
	time.Sleep(3 * time.Second)

	if _, stopped := ctx.OPCommand(); stopped {
		return nil
	}

	_ = ctx.Checkin("done")

	return nil
}
