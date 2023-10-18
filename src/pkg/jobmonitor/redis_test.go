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
	"encoding/json"
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

func (s *RedisClientTestSuite) SetupSuite() {
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		s.FailNow("REDIS_HOST is not specified")
	}
	s.redisURL = fmt.Sprintf("redis://%s:6379", redisHost)
	pool, err := redisPool(&config.RedisPoolConfig{RedisURL: s.redisURL, Namespace: "{jobservice_namespace}", IdleTimeoutSecond: 30})
	s.redisClient = redisClientImpl{
		redisPool: pool,
		namespace: "{harbor_job_service_namespace}",
	}
	if err != nil {
		s.FailNow("failed to create redis client", err)
	}
}

func (s *RedisClientTestSuite) TearDownSuite() {
}

func (s *RedisClientTestSuite) TestUntrackJobStatusInBatch() {
	// create key and value
	jobIDs := make([]string, 0)
	conn := s.redisClient.redisPool.Get()
	defer conn.Close()
	for i := 0; i < 100; i++ {
		k := utils.GenerateRandomStringWithLen(10)
		jobIDs = append(jobIDs, k)
		key := rds.KeyJobStats(fmt.Sprintf("{%s}", s.redisClient.namespace), k)
		v := utils.GenerateRandomStringWithLen(10)
		_, err := conn.Do("HSET", key, k, v)
		if err != nil {
			s.FailNow("can not insert data to redis", err)
		}
	}

	s.redisClient.removeJobStatusInRedis(context.Background(), jobIDs)
	key := rds.KeyJobStats(fmt.Sprintf("{%s}", s.redisClient.namespace), "*")
	result, err := conn.Do("KEYS", key)
	if err != nil {
		s.FailNow("can not get data from redis", err)
	}
	remains, err := redis.Values(result, err)
	if err != nil {
		s.FailNow("can not get data from redis", err)
	}
	s.Equal(0, len(remains))
}

func (s *RedisClientTestSuite) TestStopPendingJobs() {
	redisKeyJobQueue := fmt.Sprintf("{%s}:jobs:%v", "{harbor_job_service_namespace}", "REPLICATION")
	// create key and value
	type jobInfo struct {
		ID     string `json:"id"`
		Params string `json:"params"`
	}
	conn := s.redisClient.redisPool.Get()
	defer conn.Close()
	for i := 0; i < 100; i++ {
		job := jobInfo{
			ID:     utils.GenerateRandomStringWithLen(10),
			Params: utils.GenerateRandomStringWithLen(10),
		}
		val, err := json.Marshal(&job)
		if err != nil {
			s.Errorf(err, "failed to marshal job info")
		}
		_, err = conn.Do("LPUSH", redisKeyJobQueue, val)
		if err != nil {
			s.FailNow("can not insert data to redis", err)
		}
	}
	// job without id
	for i := 0; i < 10; i++ {
		job := jobInfo{
			Params: utils.GenerateRandomStringWithLen(10),
		}
		val, err := json.Marshal(&job)
		if err != nil {
			s.Errorf(err, "failed to marshal job info")
		}
		_, err = conn.Do("LPUSH", redisKeyJobQueue, val)
		if err != nil {
			s.FailNow("can not insert data to redis", err)
		}
	}

	jobIDs, err := s.redisClient.StopPendingJobs(context.Background(), "REPLICATION")
	if err != nil {
		s.FailNow("failed to stop pending jobs", err)
	}
	s.Assert().Equal(100, len(jobIDs))
}

func TestRedisClientTestSuite(t *testing.T) {
	suite.Run(t, &RedisClientTestSuite{})
}
