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

package mgt

import (
	"context"
	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// BasicManagerTestSuite tests the function of basic manager
type BasicManagerTestSuite struct {
	suite.Suite

	namespace string
	pool      *redis.Pool

	manager Manager
}

// SetupSuite prepares the test suite
func (suite *BasicManagerTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()
	suite.manager = NewManager(context.TODO(), suite.namespace, suite.pool)

	// Mock data
	periodicJob := &job.Stats{
		Info: &job.StatsInfo{
			JobID:    "1000",
			JobName:  job.SampleJob,
			JobKind:  job.KindPeriodic,
			Status:   job.ScheduledStatus.String(),
			IsUnique: false,
			CronSpec: "* */10 * * * *",
		},
	}

	t := job.NewBasicTrackerWithStats(context.TODO(), periodicJob, suite.namespace, suite.pool, nil)
	err := t.Save()
	require.NoError(suite.T(), err)

	execution := &job.Stats{
		Info: &job.StatsInfo{
			JobID:         "1001",
			JobKind:       job.KindScheduled,
			JobName:       job.SampleJob,
			Status:        job.PendingStatus.String(),
			IsUnique:      false,
			RunAt:         time.Now().Add(5 * time.Minute).Unix(),
			UpstreamJobID: "1000",
		},
	}
	t = job.NewBasicTrackerWithStats(context.TODO(), execution, suite.namespace, suite.pool, nil)
	err = t.Save()
	require.NoError(suite.T(), err)
}

// TearDownSuite clears the test suite
func (suite *BasicManagerTestSuite) TearDownSuite() {
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	_ = tests.ClearAll(suite.namespace, conn)
}

// TestBasicManagerTestSuite is entry of go test
func TestBasicManagerTestSuite(t *testing.T) {
	suite.Run(t, new(BasicManagerTestSuite))
}

// TestBasicManagerGetJobs tests get jobs
func (suite *BasicManagerTestSuite) TestBasicManagerGetJobs() {
	jobs, _, err := suite.manager.GetJobs(&query.Parameter{
		PageSize:   25,
		PageNumber: 1,
	})
	require.NoError(suite.T(), err)
	assert.Condition(suite.T(), func() bool {
		return len(jobs) > 0
	})
}

// TestGetPeriodicExecutions tests get periodic executions
func (suite *BasicManagerTestSuite) TestGetPeriodicExecutions() {
	extras := make(query.ExtraParameters)
	extras.Set(query.ExtraParamKeyNonStoppedOnly, true)

	jobs, total, err := suite.manager.GetPeriodicExecution("1000", &query.Parameter{
		PageSize:   20,
		PageNumber: 1,
		Extras:     extras,
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), int64(1), int64(len(jobs)))

	t := job.NewBasicTrackerWithID(context.TODO(), "1001", suite.namespace, suite.pool, nil)
	err = t.Load()
	require.NoError(suite.T(), err)
	err = t.PeriodicExecutionDone()
	require.NoError(suite.T(), err)

	jobs, total, err = suite.manager.GetPeriodicExecution("1000", &query.Parameter{
		PageSize:   20,
		PageNumber: 1,
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), int64(1), int64(len(jobs)))
}

// TestGetScheduledJobs tests get scheduled jobs
func (suite *BasicManagerTestSuite) TestGetScheduledJobs() {
	enqueuer := work.NewEnqueuer(suite.namespace, suite.pool)
	scheduledJob, err := enqueuer.EnqueueIn(job.SampleJob, 1000, make(map[string]interface{}))
	require.NoError(suite.T(), err)
	stats := &job.Stats{
		Info: &job.StatsInfo{
			JobID:   scheduledJob.ID,
			JobName: job.SampleJob,
			JobKind: job.KindScheduled,
			Status:  job.ScheduledStatus.String(),
			RunAt:   scheduledJob.RunAt,
		},
	}

	t := job.NewBasicTrackerWithStats(context.TODO(), stats, suite.namespace, suite.pool, nil)
	err = t.Save()
	require.NoError(suite.T(), err)

	list, total, err := suite.manager.GetScheduledJobs(&query.Parameter{
		PageNumber: 1,
		PageSize:   10,
	})
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), total)
	assert.Equal(suite.T(), int64(1), int64(len(list)))
}

// TestGetJob tests get job
func (suite *BasicManagerTestSuite) TestGetJob() {
	j, err := suite.manager.GetJob("1001")
	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), j.Info.JobID, "1001")
}

// TestSaveJob tests saving job
func (suite *BasicManagerTestSuite) TestSaveJob() {
	newJob := &job.Stats{
		Info: &job.StatsInfo{
			JobID:    "1002",
			JobKind:  job.KindPeriodic,
			JobName:  job.SampleJob,
			Status:   job.PendingStatus.String(),
			IsUnique: false,
		},
	}

	err := suite.manager.SaveJob(newJob)
	require.NoError(suite.T(), err)
}
