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

package job

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"testing"
)

// TrackerTestSuite tests functions of tracker
type TrackerTestSuite struct {
	suite.Suite

	tracker   *basicTracker
	namespace string
	pool      *redis.Pool
}

// TestTrackerTestSuite is entry of go test
func TestTrackerTestSuite(t *testing.T) {
	suite.Run(t, new(TrackerTestSuite))
}

// SetupSuite prepares test suite
func (suite *TrackerTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()
}

// SetupTest prepares for test cases
func (suite *TrackerTestSuite) SetupTest() {
	suite.tracker = &basicTracker{
		namespace: suite.namespace,
		context:   context.Background(),
		pool:      suite.pool,
		callback:  func(hookURL string, change *StatusChange) error { return nil },
	}
}

// TearDownSuite prepares test suites
func (suite *TrackerTestSuite) TearDownSuite() {
	conn := suite.pool.Get()
	defer conn.Close()

	tests.ClearAll(suite.namespace, conn)
}

// TestTracker tests tracker
func (suite *TrackerTestSuite) TestTracker() {
	jobID := utils.MakeIdentifier()
	mockJobStats := &Stats{
		Info: &StatsInfo{
			JobID:    jobID,
			Status:   SuccessStatus.String(),
			JobKind:  KindGeneric,
			JobName:  SampleJob,
			IsUnique: false,
		},
	}

	suite.tracker.jobStats = mockJobStats
	suite.tracker.jobID = jobID

	err := suite.tracker.Save()
	require.Nil(suite.T(), err, "save: nil error expected but got %s", err)

	s, err := suite.tracker.Status()
	assert.Nil(suite.T(), err, "get status: nil error expected but got %s", err)
	assert.Equal(suite.T(), SuccessStatus, s, "get status: expected pending but got %s", s)

	j := suite.tracker.Job()
	assert.Equal(suite.T(), jobID, j.Info.JobID, "job: expect job ID %s but got %s", jobID, j.Info.JobID)

	err = suite.tracker.Update("web_hook_url", "http://hook.url")
	assert.Nil(suite.T(), err, "update: nil error expected but got %s", err)

	err = suite.tracker.Load()
	assert.Nil(suite.T(), err, "load: nil error expected but got %s", err)
	assert.Equal(
		suite.T(),
		"http://hook.url",
		suite.tracker.jobStats.Info.WebHookURL,
		"web hook: expect %s but got %s",
		"http://hook.url",
		suite.tracker.jobStats.Info.WebHookURL,
	)

	err = suite.tracker.Run()
	assert.Error(suite.T(), err, "run: non nil error expected but got nil")
	err = suite.tracker.CheckIn("check in")
	assert.Nil(suite.T(), err, "check in: nil error expected but got %s", err)
	err = suite.tracker.Succeed()
	assert.Nil(suite.T(), err, "succeed: nil error expected but got %s", err)
	err = suite.tracker.Stop()
	assert.Nil(suite.T(), err, "stop: nil error expected but got %s", err)
	err = suite.tracker.Fail()
	assert.Nil(suite.T(), err, "fail: nil error expected but got %s", err)
}
