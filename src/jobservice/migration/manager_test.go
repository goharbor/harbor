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

package migration

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/tests"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ManagerTestSuite tests functions of manager
type ManagerTestSuite struct {
	suite.Suite

	pool      *redis.Pool
	namespace string

	manager Manager

	jobID      string
	numbericID int64
}

// TestManagerTestSuite is entry of executing ManagerTestSuite
func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

// SetupAllSuite sets up env for test suite
func (suite *ManagerTestSuite) SetupSuite() {
	suite.pool = tests.GiveMeRedisPool()
	suite.namespace = tests.GiveMeTestNamespace()

	suite.manager = New(suite.pool, suite.namespace)
}

// SetupTestSuite sets up env for each test case
func (suite *ManagerTestSuite) SetupTest() {
	// Mock fake data
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	id := utils.MakeIdentifier()
	suite.jobID = id
	// Mock stats of periodic job
	args := []interface{}{
		rds.KeyJobStats(suite.namespace, id),
		"status_hook",
		"http://core:8080/hook",
		"id",
		id,
		"name",
		job.ImageGC,
		"kind",
		job.KindPeriodic,
		"unique",
		0,
		"status",
		job.SuccessStatus.String(), // v1.6 issue
		"ref_link",
		fmt.Sprintf("/api/v1/jobs/%s", id),
		"enqueue_time",
		time.Now().Unix(),
		"update_time",
		time.Now().Unix(),
		"run_at",
		time.Now().Add(5 * time.Minute).Unix(),
		"cron_spec",
		"0 0 17 * * *",
		"multiple_executions", // V1.7
		1,
	}
	reply, err := redis.String(conn.Do("HMSET", args...))
	require.NoError(suite.T(), err, "mock job stats data error")
	require.Equal(suite.T(), "ok", strings.ToLower(reply), "ok expected")

	// Mock periodic job policy object
	params := make(map[string]interface{})
	params["redis_url_reg"] = "redis://redis:6379/1"

	policy := make(map[string]interface{})
	policy["job_name"] = job.ImageGC
	policy["job_params"] = params
	policy["cron_spec"] = "0 0 17 * * *"

	rawJSON, err := json.Marshal(&policy)
	require.NoError(suite.T(), err, "mock periodic job policy error")

	policy["cron_spec"] = "0 0 8 * * *"
	duplicatedRawJSON, err := json.Marshal(&policy)
	require.NoError(suite.T(), err, "mock duplicated periodic job policy error")

	score := time.Now().Unix()
	suite.numbericID = score
	zaddArgs := []interface{}{
		rds.KeyPeriodicPolicy(suite.namespace),
		score,
		rawJSON,
		score - 10,
		duplicatedRawJSON, // duplicated one
	}
	count, err := redis.Int(conn.Do("ZADD", zaddArgs...))
	require.NoError(suite.T(), err, "add raw policy error")
	require.Equal(suite.T(), 2, count)

	// Mock key score mapping
	keyScoreArgs := []interface{}{
		fmt.Sprintf("%s%s", rds.KeyNamespacePrefix(suite.namespace), "period:key_score"),
		score,
		id,
	}

	count, err = redis.Int(conn.Do("ZADD", keyScoreArgs...))
	require.NoError(suite.T(), err, "add key score mapping error")
	require.Equal(suite.T(), 1, count)
}

// SetupTestSuite clears up env for each test case
func (suite *ManagerTestSuite) TearDownTest() {
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	err := tests.ClearAll(suite.namespace, conn)
	assert.NoError(suite.T(), err, "clear all of redis db error")
}

// TestManager test the basic functions of the manager
func (suite *ManagerTestSuite) TestManager() {
	require.NotNil(suite.T(), suite.manager, "nil migration manager")

	suite.manager.Register(PolicyMigratorFactory)
	err := suite.manager.Migrate()
	require.NoError(suite.T(), err, "migrating rdb error")

	// Check data
	conn := suite.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	count, err := redis.Int(conn.Do("ZCARD", rds.KeyPeriodicPolicy(suite.namespace)))
	assert.NoError(suite.T(), err, "get count of policies error")
	assert.Equal(suite.T(), 1, count)

	innerConn := suite.pool.Get()
	p, err := getPeriodicPolicy(suite.numbericID, innerConn, suite.namespace)
	assert.NoError(suite.T(), err, "get migrated policy error")
	assert.NotEmpty(suite.T(), p.ID, "ID of policy")
	assert.NotEmpty(suite.T(), p.WebHookURL, "Web hook URL of policy")

	key := fmt.Sprintf("%s%s", rds.KeyNamespacePrefix(suite.namespace), "period:key_score")
	count, err = redis.Int(conn.Do("EXISTS", key))
	assert.NoError(suite.T(), err, "check existence of key score mapping error")
	assert.Equal(suite.T(), 0, count)

	hmGetArgs := []interface{}{
		rds.KeyJobStats(suite.namespace, suite.jobID),
		"id",
		"status",
		"web_hook_url",
		"numeric_policy_id",
		"multiple_executions",
		"status_hook",
	}
	fields, err := redis.Values(conn.Do("HMGET", hmGetArgs...))
	assert.NoError(suite.T(), err, "check migrated job stats error")
	assert.Equal(suite.T(), suite.jobID, toString(fields[0]), "check job ID")
	assert.Equal(suite.T(), job.ScheduledStatus.String(), toString(fields[1]), "check job status")
	assert.Equal(suite.T(), "http://core:8080/hook", toString(fields[2]), "check web hook URL")
	assert.Equal(suite.T(), suite.numbericID, toInt(fields[3]), "check numberic ID")
	assert.Nil(suite.T(), fields[4], "'multiple_executions' removed")
	assert.Nil(suite.T(), fields[5], "'status_hook' removed")
}
