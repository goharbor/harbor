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
package period

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/rds"

	"github.com/gocraft/work"

	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// BasicSchedulerTestSuite tests functions of basic scheduler
type BasicSchedulerTestSuite struct {
	suite.Suite

	cancel    context.CancelFunc
	namespace string
	pool      *redis.Pool

	lcmCtl    lcm.Controller
	scheduler Scheduler
}

// SetupSuite prepares the test suite
func (suite *BasicSchedulerTestSuite) SetupSuite() {
	ctx, cancel := context.WithCancel(context.WithValue(context.Background(), utils.NodeID, "fake_node_ID"))
	suite.cancel = cancel

	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()

	envCtx := &env.Context{
		SystemContext: ctx,
		WG:            new(sync.WaitGroup),
	}

	suite.lcmCtl = lcm.NewController(
		envCtx,
		suite.namespace,
		suite.pool,
		func(hookURL string, change *job.StatusChange) error { return nil },
	)

	suite.setupDirtyJobs()

	suite.scheduler = NewScheduler(ctx, suite.namespace, suite.pool, suite.lcmCtl)
	suite.scheduler.Start()
}

// TearDownSuite clears the test suite
func (suite *BasicSchedulerTestSuite) TearDownSuite() {
	suite.cancel()

	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_ = tests.ClearAll(suite.namespace, conn)
}

// TestSchedulerTestSuite is entry of go test
func TestSchedulerTestSuite(t *testing.T) {
	suite.Run(t, new(BasicSchedulerTestSuite))
}

// TestScheduler tests scheduling and un-scheduling
func (suite *BasicSchedulerTestSuite) TestScheduler() {
	// Prepare one
	now := time.Now()
	minute := now.Minute()
	if minute+2 >= 60 {
		minute = minute - 2
	}
	coreSpec := fmt.Sprintf("30,50 %d * * * *", minute+2)
	p := &Policy{
		ID:       "fake_policy",
		JobName:  job.SampleJob,
		CronSpec: coreSpec,
	}

	pid, err := suite.scheduler.Schedule(p)
	require.NoError(suite.T(), err, "schedule: nil error expected but got %s", err)
	assert.Condition(suite.T(), func() bool {
		return pid > 0
	}, "schedule: returned pid should >0")

	jobStats := &job.Stats{
		Info: &job.StatsInfo{
			JobID:      p.ID,
			Status:     job.ScheduledStatus.String(),
			JobName:    job.SampleJob,
			JobKind:    job.KindPeriodic,
			NumericPID: pid,
			CronSpec:   coreSpec,
		},
	}
	_, err = suite.lcmCtl.New(jobStats)
	require.NoError(suite.T(), err, "lcm new: nil error expected but got %s", err)

	err = suite.scheduler.UnSchedule(p.ID)
	require.NoError(suite.T(), err, "unschedule: nil error expected but got %s", err)
}

// TestUnSchedule tests un-scheduling a periodic job without job stats
func (suite *BasicSchedulerTestSuite) TestUnSchedule() {
	p := &Policy{
		ID:       "job_id_without_stats",
		JobName:  job.SampleJob,
		CronSpec: "0 10 10 5 * *",
	}

	pid, err := suite.scheduler.Schedule(p)
	require.NoError(suite.T(), err, "schedule: nil error expected but got %s", err)
	assert.Condition(suite.T(), func() bool {
		return pid > 0
	}, "schedule: returned pid should >0")

	// No job stats saved
	err = suite.scheduler.UnSchedule(p.ID)
	require.NoError(suite.T(), err, "unschedule: nil error expected but got %s", err)
}

// setupDirtyJobs adds dirty jobs for testing dirty jobs clear method in the Start()
func (suite *BasicSchedulerTestSuite) setupDirtyJobs() {
	// Add one fake job for next testing
	j := &work.Job{
		Name: job.SampleJob,
		ID:   "jid",
		// Already expired
		EnqueuedAt: time.Now().Unix() - 86400,
		Args:       map[string]interface{}{"image": "sample:latest"},
	}

	rawJSON, err := utils.SerializeJob(j)
	suite.NoError(err, "serialize job model")

	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_, err = conn.Do("ZADD", rds.RedisKeyScheduled(suite.namespace), j.EnqueuedAt, rawJSON)
	suite.NoError(err, "add faked dirty scheduled job")
}
