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
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// PolicyStoreTestSuite tests functions of policy store
type PolicyStoreTestSuite struct {
	suite.Suite

	namespace string
	pool      *redis.Pool
}

// TestPolicyStoreTestSuite is entry of go test
func TestPolicyStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyStoreTestSuite))
}

// SetupSuite prepares test suite
func (suite *PolicyStoreTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()
}

// TearDownSuite clears the test suite
func (suite *PolicyStoreTestSuite) TearDownSuite() {
	conn := suite.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			suite.NoError(err, "close redis connection")
		}
	}()

	if err := tests.ClearAll(suite.namespace, conn); err != nil {
		suite.NoError(err, "clear redis namespace")
	}
}

// TestLoad tests load policy from backend
func (suite *PolicyStoreTestSuite) TestLoad() {
	// Prepare one
	p := &Policy{
		ID:       "fake_policy",
		JobName:  job.SampleJob,
		CronSpec: "5 * * * * *",
	}
	rawData, err := p.Serialize()
	assert.Nil(suite.T(), err, "prepare data: nil error expected but got %s", err)
	key := rds.KeyPeriodicPolicy(suite.namespace)

	conn := suite.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			suite.NoError(err, "close redis connection")
		}
	}()

	_, err = conn.Do("ZADD", key, time.Now().Unix(), rawData)
	assert.Nil(suite.T(), err, "add data: nil error expected but got %s", err)

	ps, err := Load(suite.namespace, conn)
	suite.NoError(err, "load: nil error expected but got %s", err)
	suite.Equal(1, len(ps), "count of loaded policies")
	suite.Greater(ps[0].NumericID, int64(0), "numericID of the policy <> 0")
}

// TestPolicy tests policy itself
func (suite *PolicyStoreTestSuite) TestPolicy() {
	p1 := &Policy{
		ID:       "fake_policy_1",
		JobName:  job.SampleJob,
		CronSpec: "5 * * * * *",
	}

	bytes, err := p1.Serialize()
	assert.Nil(suite.T(), err, "policy serialize: nil error expected but got %s", err)
	p2 := &Policy{}
	err = p2.DeSerialize(bytes)
	assert.Nil(suite.T(), err, "policy deserialize: nil error expected but got %s", err)
	assert.Equal(suite.T(), "5 * * * * *", p2.CronSpec)
	err = p2.Validate()
	assert.Nil(suite.T(), err, "policy validate: nil error expected but got %s", err)
}

// TestInvalidPolicy tests invalid policy
func (suite *PolicyStoreTestSuite) TestInvalidPolicy() {
	p := &Policy{}
	suite.Error(p.Validate(), "error should be returned for empty ID")
	p.ID = "pid"
	suite.Error(p.Validate(), "error should be returned for empty job name")
	p.JobName = "GC"
	p.WebHookURL = "webhook"
	suite.Error(p.Validate(), "error should be returned for invalid webhook")
	p.WebHookURL = "https://webhook.com"
	suite.Error(p.Validate(), "error should be returned for invalid cron spec")
	p.CronSpec = "0 10 10 * * *"
	suite.NoError(p.Validate(), "validation passed")
}
