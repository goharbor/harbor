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
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/lib/log"
	libRedis "github.com/goharbor/harbor/src/lib/redis"
)

// JobServicePool job service pool name
const JobServicePool = "JobService"

// RedisClient defines the job service operations related to redis
type RedisClient interface {
	// AllJobTypes returns all the job types registered in the job service
	AllJobTypes(ctx context.Context) ([]string, error)
	// PauseJob pause the execution of the specified type job, except the running job
	PauseJob(ctx context.Context, jobName string) error
	// UnpauseJob resume the execution of the specified type job
	UnpauseJob(ctx context.Context, jobName string) error
	// StopPendingJobs stop the pending jobs of the specified type, and remove the jobs from the waiting queue
	StopPendingJobs(ctx context.Context, jobType string) (jobIDs []string, err error)
}

type redisClientImpl struct {
	redisPool *redis.Pool
	namespace string
}

// NewRedisClient create a redis client
func NewRedisClient(config *config.RedisPoolConfig) (RedisClient, error) {
	pool, err := redisPool(config)
	if err != nil {
		return nil, err
	}
	return &redisClientImpl{pool, config.Namespace}, nil
}

func redisPool(config *config.RedisPoolConfig) (*redis.Pool, error) {
	return libRedis.GetRedisPool(JobServicePool, config.RedisURL, &libRedis.PoolParam{
		PoolMaxIdle:     0,
		PoolIdleTimeout: time.Duration(config.IdleTimeoutSecond) * time.Second,
	})
}

func (r *redisClientImpl) StopPendingJobs(ctx context.Context, jobType string) (jobIDs []string, err error) {
	jobIDs = []string{}
	log.Infof("job queue cleaned up %s", jobType)
	redisKeyJobQueue := fmt.Sprintf("{%s}:jobs:%v", r.namespace, jobType)
	conn := r.redisPool.Get()
	defer conn.Close()
	var jobInfo struct {
		ID string `json:"id"`
	}
	jobs, err := redis.Strings(conn.Do("LRANGE", redisKeyJobQueue, 0, -1))
	if err != nil {
		return []string{}, err
	}
	if len(jobs) == 0 {
		log.Infof("no pending job for job type %v", jobType)
		return []string{}, nil
	}
	for _, j := range jobs {
		if err := json.Unmarshal([]byte(j), &jobInfo); err != nil {
			log.Errorf("failed to parse the job info %v, %v", j, err)
			continue
		}
		jobIDs = append(jobIDs, jobInfo.ID)
	}
	log.Infof("updated %d tasks in pending status to stop", len(jobIDs))
	ret, err := redis.Int64(conn.Do("DEL", redisKeyJobQueue))
	if err != nil {
		return []string{}, err
	}
	if ret < 1 {
		// no job in queue removed
		return []string{}, fmt.Errorf("no job in the queue removed")
	}
	log.Infof("deleted %d keys in waiting queue for %s", ret, jobType)
	log.Debugf("job id to be deleted %v", jobIDs)
	return jobIDs, nil
}

func (r *redisClientImpl) AllJobTypes(ctx context.Context) ([]string, error) {
	conn := r.redisPool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("SMEMBERS", fmt.Sprintf("{%s}:known_jobs", r.namespace)))
}

func (r *redisClientImpl) PauseJob(ctx context.Context, jobName string) error {
	log.Infof("pause job type:%s", jobName)
	redisKeyJobPaused := fmt.Sprintf("{%s}:jobs:%s:paused", r.namespace, jobName)
	conn := r.redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", redisKeyJobPaused, "1")
	return err
}

func (r *redisClientImpl) UnpauseJob(ctx context.Context, jobName string) error {
	log.Infof("unpause job %s", jobName)
	redisKeyJobPaused := fmt.Sprintf("{%s}:jobs:%s:paused", r.namespace, jobName)
	conn := r.redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("DEL", redisKeyJobPaused)
	return err
}

// JobServiceRedisClient function to create redis client for job service
func JobServiceRedisClient() (RedisClient, error) {
	cfg, err := job.GlobalClient.GetJobServiceConfig()
	if err != nil {
		return nil, err
	}
	config := cfg.RedisPoolConfig
	return NewRedisClient(config)
}
