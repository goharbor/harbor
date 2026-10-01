// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/tests"
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

// TestUnScheduleSharedNumericID tests un-scheduling a policy whose numeric ID is shared by other policies
func (suite *BasicSchedulerTestSuite) TestUnScheduleSharedNumericID() {
	ids := suite.scheduleOnSharedScore("shared_numeric_id", 5, 1000, false)
	defer suite.removePolicies(ids)

	err := suite.scheduler.UnSchedule(ids[0])
	require.NoError(suite.T(), err, "unschedule: nil error expected but got %s", err)

	suite.assertPolicies(ids[0], ids[1:])
}

// TestUnScheduleSharedNumericIDWithStats tests un-scheduling a policy with job stats whose numeric ID is shared by other policies
func (suite *BasicSchedulerTestSuite) TestUnScheduleSharedNumericIDWithStats() {
	ids := suite.scheduleOnSharedScore("shared_numeric_id_stats", 5, 1500, true)
	defer suite.removePolicies(ids)

	err := suite.scheduler.UnSchedule(ids[0])
	require.NoError(suite.T(), err, "unschedule: nil error expected but got %s", err)

	suite.assertPolicies(ids[0], ids[1:])
}

// TestUnScheduleSkipsMalformedMember tests that un-scheduling skips members at the same score that cannot be decoded
func (suite *BasicSchedulerTestSuite) TestUnScheduleSkipsMalformedMember() {
	ids := suite.scheduleOnSharedScore("malformed_neighbour", 1, 1100, false)
	defer suite.removePolicies(ids)

	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	key := rds.KeyPeriodicPolicy(suite.namespace)
	for _, m := range []string{"not-json", `"a string"`} {
		_, err := conn.Do("ZADD", key, 1100, m)
		require.NoError(suite.T(), err, "zadd: nil error expected but got %s", err)
	}
	defer func() {
		_, _ = conn.Do("ZREM", key, "not-json", `"a string"`)
	}()

	err := suite.scheduler.UnSchedule(ids[0])
	require.NoError(suite.T(), err, "unschedule: nil error expected but got %s", err)
	suite.assertPolicies(ids[0], nil)

	n, err := redis.Int(conn.Do("ZCOUNT", key, 1100, 1100))
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), 2, n, "malformed members should be kept")
}

// scheduleOnSharedScore schedules n policies and moves all of them to the given score.
func (suite *BasicSchedulerTestSuite) scheduleOnSharedScore(prefix string, n int, score int64, withStats bool) []string {
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		p := &Policy{
			ID:       fmt.Sprintf("%s_%d", prefix, i),
			JobName:  job.SampleJob,
			CronSpec: "0 10 10 5 * *",
		}
		_, err := suite.scheduler.Schedule(p)
		require.NoError(suite.T(), err, "schedule: nil error expected but got %s", err)
		ids = append(ids, p.ID)

		if withStats {
			_, err = suite.lcmCtl.New(&job.Stats{
				Info: &job.StatsInfo{
					JobID:      p.ID,
					Status:     job.ScheduledStatus.String(),
					JobName:    p.JobName,
					JobKind:    job.KindPeriodic,
					NumericPID: score,
					CronSpec:   p.CronSpec,
				},
			})
			require.NoError(suite.T(), err, "lcm new: nil error expected but got %s", err)
		}
	}

	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	for _, m := range suite.policyMembers(ids) {
		_, err := conn.Do("ZADD", rds.KeyPeriodicPolicy(suite.namespace), score, m)
		require.NoError(suite.T(), err, "zadd: nil error expected but got %s", err)
	}

	return ids
}

// policyMembers returns the raw policy set members of the given policy IDs.
func (suite *BasicSchedulerTestSuite) policyMembers(ids []string) map[string][]byte {
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	members, err := redis.ByteSlices(conn.Do("ZRANGE", rds.KeyPeriodicPolicy(suite.namespace), 0, -1))
	require.NoError(suite.T(), err, "list policies: nil error expected but got %s", err)

	res := make(map[string][]byte)
	for _, m := range members {
		p := &Policy{}
		if err := p.DeSerialize(m); err != nil {
			continue
		}
		for _, id := range ids {
			if p.ID == id {
				res[id] = m
			}
		}
	}

	return res
}

// removePolicies removes the given policies from the policy set without going through the scheduler.
func (suite *BasicSchedulerTestSuite) removePolicies(ids []string) {
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	for _, m := range suite.policyMembers(ids) {
		_, _ = conn.Do("ZREM", rds.KeyPeriodicPolicy(suite.namespace), m)
	}
}

// assertPolicies asserts that removed is not in the policy set and all of kept are.
func (suite *BasicSchedulerTestSuite) assertPolicies(removed string, kept []string) {
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	policies, err := Load(suite.namespace, conn)
	require.NoError(suite.T(), err, "load policies: nil error expected but got %s", err)
	remaining := make([]string, 0, len(policies))
	for _, p := range policies {
		remaining = append(remaining, p.ID)
	}
	if removed != "" {
		assert.NotContains(suite.T(), remaining, removed, "unscheduled policy should be removed")
	}
	for _, id := range kept {
		assert.Contains(suite.T(), remaining, id, "policy sharing the numeric ID should be kept")
	}
}

// setupDirtyJobs adds dirty jobs for testing dirty jobs clear method in the Start()
func (suite *BasicSchedulerTestSuite) setupDirtyJobs() {
	// Add one fake job for next testing
	j := &work.Job{
		Name: job.SampleJob,
		ID:   "jid",
		// Already expired
		EnqueuedAt: time.Now().Unix() - 86400,
		Args:       map[string]any{"image": "sample:latest"},
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
