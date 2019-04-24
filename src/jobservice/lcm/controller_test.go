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

package lcm

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"sync"
	"testing"
	"time"
)

// LcmControllerTestSuite tests functions of life cycle controller
type LcmControllerTestSuite struct {
	suite.Suite

	namespace string
	pool      *redis.Pool
	ctl       Controller
	cancel    context.CancelFunc
}

// SetupSuite prepares test suite
func (suite *LcmControllerTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()

	ctx, cancel := context.WithCancel(context.Background())
	suite.cancel = cancel
	envCtx := &env.Context{
		SystemContext: ctx,
		WG:            new(sync.WaitGroup),
	}
	suite.ctl = NewController(envCtx, suite.namespace, suite.pool, func(hookURL string, change *job.StatusChange) error { return nil })
}

// TearDownSuite clears test suite
func (suite *LcmControllerTestSuite) TearDownSuite() {
	suite.cancel()
}

// TestLcmControllerTestSuite is entry of go test
func TestLcmControllerTestSuite(t *testing.T) {
	suite.Run(t, new(LcmControllerTestSuite))
}

// TestNewAndTrack tests controller.New() and controller.Track()
func (suite *LcmControllerTestSuite) TestNewAndTrack() {
	jobID := utils.MakeIdentifier()
	suite.newsStats(jobID)

	t, err := suite.ctl.Track(jobID)
	require.Nil(suite.T(), err, "lcm track: nil error expected but got %s", err)
	assert.Equal(suite.T(), job.SampleJob, t.Job().Info.JobName, "lcm new: expect job name %s but got %s", job.SampleJob, t.Job().Info.JobName)
}

// TestNew tests controller.Serve()
func (suite *LcmControllerTestSuite) TestServe() {
	// Prepare mock data
	jobID := utils.MakeIdentifier()
	suite.newsStats(jobID)

	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()
	simpleChange := &job.SimpleStatusChange{
		JobID:        jobID,
		TargetStatus: job.RunningStatus.String(),
	}
	rawJSON, err := json.Marshal(simpleChange)
	require.Nil(suite.T(), err, "json marshal: nil error expected but got %s", err)
	key := rds.KeyStatusUpdateRetryQueue(suite.namespace)
	args := []interface{}{key, "NX", time.Now().Unix(), rawJSON}
	_, err = conn.Do("ZADD", args...)
	require.Nil(suite.T(), err, "prepare mock data: nil error expected but got %s", err)

	err = suite.ctl.Serve()
	require.NoError(suite.T(), err, "lcm: nil error expected but got %s", err)
	<-time.After(1 * time.Second)

	count, err := redis.Int(conn.Do("ZCARD", key))
	require.Nil(suite.T(), err, "get total dead status: nil error expected but got %s", err)
	assert.Equal(suite.T(), 0, count)
}

// newsStats create job stats
func (suite *LcmControllerTestSuite) newsStats(jobID string) {
	stats := &job.Stats{
		Info: &job.StatsInfo{
			JobID:    jobID,
			JobKind:  job.KindGeneric,
			JobName:  job.SampleJob,
			IsUnique: true,
			Status:   job.PendingStatus.String(),
		},
	}

	t, err := suite.ctl.New(stats)
	require.Nil(suite.T(), err, "lcm new: nil error expected but got %s", err)
	assert.Equal(suite.T(), jobID, t.Job().Info.JobID, "lcm new: expect job ID %s but got %s", jobID, t.Job().Info.JobID)
}
