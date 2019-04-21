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
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

// PolicyStoreTestSuite tests functions of policy store
type PolicyStoreTestSuite struct {
	suite.Suite

	store     *policyStore
	namespace string
	pool      *redis.Pool
	cancel    context.CancelFunc
}

// TestPolicyStoreTestSuite is entry of go test
func TestPolicyStoreTestSuite(t *testing.T) {
	suite.Run(t, new(PolicyStoreTestSuite))
}

// SetupSuite prepares test suite
func (suite *PolicyStoreTestSuite) SetupSuite() {
	suite.namespace = tests.GiveMeTestNamespace()
	suite.pool = tests.GiveMeRedisPool()
	ctx, cancel := context.WithCancel(context.Background())
	suite.cancel = cancel

	suite.store = newPolicyStore(ctx, suite.namespace, suite.pool)
}

// TearDownSuite clears the test suite
func (suite *PolicyStoreTestSuite) TearDownSuite() {
	suite.cancel()

	conn := suite.pool.Get()
	defer conn.Close()

	tests.ClearAll(suite.namespace, conn)
}

// TestStore tests policy store serve
func (suite *PolicyStoreTestSuite) TestServe() {
	var err error

	defer func() {
		suite.store.stopChan <- true
		assert.Nil(suite.T(), err, "serve exit: nil error expected but got %s", err)
	}()

	go func() {
		err = suite.store.serve()
	}()
	<-time.After(1 * time.Second)
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
	defer conn.Close()

	_, err = conn.Do("ZADD", key, time.Now().Unix(), rawData)
	assert.Nil(suite.T(), err, "add data: nil error expected but got %s", err)

	err = suite.store.load()
	assert.Nil(suite.T(), err, "load: nil error expected but got %s", err)

	p1 := &Policy{
		ID:       "fake_policy_1",
		JobName:  job.SampleJob,
		CronSpec: "5 * * * * *",
	}
	m := &message{
		Event: changeEventSchedule,
		Data:  p1,
	}
	err = suite.store.sync(m)
	assert.Nil(suite.T(), err, "sync schedule: nil error expected but got %s", err)

	count := 0
	suite.store.Iterate(func(id string, p *Policy) bool {
		count++
		return true
	})
	assert.Equal(suite.T(), 2, count, "expected 2 policies but got %d", count)

	m1 := &message{
		Event: changeEventUnSchedule,
		Data:  p1,
	}
	err = suite.store.sync(m1)
	assert.Nil(suite.T(), err, "sync unschedule: nil error expected but got %s", err)

	count = 0
	suite.store.Iterate(func(id string, p *Policy) bool {
		count++
		return true
	})
	assert.Equal(suite.T(), 1, count, "expected 1 policies but got %d", count)
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
