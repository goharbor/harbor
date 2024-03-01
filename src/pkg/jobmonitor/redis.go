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

package jobmonitor

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"

	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/lib/log"
	libRedis "github.com/goharbor/harbor/src/lib/redis"
)

// JobServicePool job service pool name
const JobServicePool = "JobService"

// batchSize the batch size to list the job in queue
const batchSize = 1000

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
	size, err := redis.Int64(conn.Do("LLEN", redisKeyJobQueue))
	if err != nil {
		log.Infof("fail to get the size of the queue")
		return []string{}, err
	}
	if size == 0 {
		return []string{}, nil
	}

	// use batch to list the job in queue, because the too many object load from a list might cause the redis crash
	for startIndex := int64(0); startIndex < int64(size); startIndex += batchSize {
		endIndex := startIndex + batchSize
		if endIndex > int64(size) {
			endIndex = int64(size)
		}
		jobs, err := redis.Strings(conn.Do("LRANGE", redisKeyJobQueue, startIndex, endIndex))
		if err != nil {
			return []string{}, err
		}
		for _, j := range jobs {
			if err := json.Unmarshal([]byte(j), &jobInfo); err != nil {
				log.Errorf("failed to parse the job info %v, %v", j, err)
				continue
			}
			if len(jobInfo.ID) > 0 {
				jobIDs = append(jobIDs, jobInfo.ID)
			}
		}
	}

	log.Infof("updated %d tasks in pending status to stop", len(jobIDs))
	ret, err := redis.Int64(conn.Do("DEL", redisKeyJobQueue))
	if err != nil {
		return []string{}, err
	}
	go func() {
		// the amount of jobIDs maybe large, so use goroutine to remove the job status tracking info
		r.removeJobStatusInRedis(ctx, jobIDs)
	}()
	if ret < 1 {
		// no job in queue removed
		return []string{}, fmt.Errorf("no job in the queue removed")
	}
	log.Infof("deleted %d keys in waiting queue for %s", ret, jobType)
	log.Debugf("job id to be deleted %v", jobIDs)
	return jobIDs, nil
}

// removeJobStatusInRedis remove job status track information from redis, to avoid performance impact when the jobIDs is too large, use batch to remove
func (r *redisClientImpl) removeJobStatusInRedis(_ context.Context, jobIDs []string) {
	conn := r.redisPool.Get()
	defer conn.Close()
	for _, id := range jobIDs {
		namespace := fmt.Sprintf("{%s}", r.namespace)
		redisKeyStatus := rds.KeyJobStats(namespace, id)
		log.Debugf("delete job status info for job id:%v, key:%v", id, redisKeyStatus)
		_, err := conn.Do("DEL", redisKeyStatus)
		if err != nil {
			log.Warningf("failed to delete the job status info for job %v, %v, continue", id, err)
		}
		redisKeyInProgress := rds.KeyJobTrackInProgress(namespace)
		log.Debugf("delete inprogress info for key:%v, job id:%v", id, redisKeyInProgress)
		_, err = conn.Do("HDEL", redisKeyInProgress, id)
		if err != nil {
			log.Warningf("failed to delete the job info in %v for job %v, %v, continue", rds.KeyJobTrackInProgress(namespace), id, err)
		}
	}
}

func (r *redisClientImpl) AllJobTypes(_ context.Context) ([]string, error) {
	conn := r.redisPool.Get()
	defer conn.Close()
	return redis.Strings(conn.Do("SMEMBERS", fmt.Sprintf("{%s}:known_jobs", r.namespace)))
}

func (r *redisClientImpl) PauseJob(_ context.Context, jobName string) error {
	log.Infof("pause job type:%s", jobName)
	redisKeyJobPaused := fmt.Sprintf("{%s}:jobs:%s:paused", r.namespace, jobName)
	conn := r.redisPool.Get()
	defer conn.Close()
	_, err := conn.Do("SET", redisKeyJobPaused, "1")
	return err
}

func (r *redisClientImpl) UnpauseJob(_ context.Context, jobName string) error {
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
