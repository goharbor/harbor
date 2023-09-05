//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package jobmonitor

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/config"
)

type RedisClientTestSuite struct {
	suite.Suite
	redisClient redisClientImpl
	redisURL    string
}

func (suite *RedisClientTestSuite) SetupSuite() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		suite.FailNow("REDIS_HOST is not specified")
	}
	suite.redisURL = fmt.Sprintf("redis://%s:6379", redisHost)
	pool, err := redisPool(&config.RedisPoolConfig{RedisURL: suite.redisURL, Namespace: "{jobservice_namespace}", IdleTimeoutSecond: 30})
	suite.redisClient = redisClientImpl{
		redisPool: pool,
		namespace: "{harbor_job_service_namespace}",
	}
	if err != nil {
		suite.FailNow("failed to create redis client", err)
	}
}

func (suite *RedisClientTestSuite) TearDownSuite() {
}

func (suite *RedisClientTestSuite) TestUntrackJobStatusInBatch() {
	// create key and value
	jobIDs := make([]string, 0)
	conn := suite.redisClient.redisPool.Get()
	defer conn.Close()
	for i := 0; i < 100; i++ {
		k := utils.GenerateRandomStringWithLen(10)
		jobIDs = append(jobIDs, k)
		key := rds.KeyJobStats(fmt.Sprintf("{%s}", suite.redisClient.namespace), k)
		v := utils.GenerateRandomStringWithLen(10)
		_, err := conn.Do("HSET", key, k, v)
		if err != nil {
			suite.FailNow("can not insert data to redis", err)
		}
	}
	suite.redisClient.removeJobStatusInRedis(context.Background(), jobIDs)
	key := rds.KeyJobStats(fmt.Sprintf("{%s}", suite.redisClient.namespace), "*")
	result, err := conn.Do("KEYS", key)
	if err != nil {
		suite.FailNow("can not get data from redis", err)
	}
	remains, err := redis.Values(result, err)
	if err != nil {
		suite.FailNow("can not get data from redis", err)
	}
	suite.Equal(0, len(remains))
}

func TestRedisClientTestSuite(t *testing.T) {
	suite.Run(t, &RedisClientTestSuite{})
}
