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
	"encoding/json"
	"errors"
	"strconv"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/opm"

	"github.com/robfig/cron"

	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
)

const (
	// EventSchedulePeriodicPolicy is for scheduling periodic policy event
	EventSchedulePeriodicPolicy = "schedule"
	// EventUnSchedulePeriodicPolicy is for unscheduling periodic policy event
	EventUnSchedulePeriodicPolicy = "unschedule"
)

// RedisPeriodicScheduler manages the periodic scheduling policies.
type RedisPeriodicScheduler struct {
	context   *env.Context
	redisPool *redis.Pool
	namespace string
	pstore    *periodicJobPolicyStore
	enqueuer  *periodicEnqueuer
}

// NewRedisPeriodicScheduler is constructor of RedisPeriodicScheduler
func NewRedisPeriodicScheduler(ctx *env.Context, namespace string, redisPool *redis.Pool, statsManager opm.JobStatsManager) *RedisPeriodicScheduler {
	pstore := &periodicJobPolicyStore{
		lock:     new(sync.RWMutex),
		policies: make(map[string]*PeriodicJobPolicy),
	}
	enqueuer := newPeriodicEnqueuer(namespace, redisPool, pstore, statsManager)

	return &RedisPeriodicScheduler{
		context:   ctx,
		redisPool: redisPool,
		namespace: namespace,
		pstore:    pstore,
		enqueuer:  enqueuer,
	}
}

// Start to serve
func (rps *RedisPeriodicScheduler) Start() {
	defer func() {
		logger.Info("Redis scheduler is stopped")
	}()

	// Load existing periodic job policies
	if err := rps.Load(); err != nil {
		// exit now
		rps.context.ErrorChan <- err
		return
	}

	// start enqueuer
	rps.enqueuer.start()
	defer rps.enqueuer.stop()
	logger.Info("Redis scheduler is started")

	// blocking here
	<-rps.context.SystemContext.Done()
}

// Schedule is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) Schedule(jobName string, params models.Parameters, cronSpec string) (string, int64, error) {
	if utils.IsEmptyStr(jobName) {
		return "", 0, errors.New("empty job name is not allowed")
	}
	if utils.IsEmptyStr(cronSpec) {
		return "", 0, errors.New("cron spec is not set")
	}

	// Get next run time
	schedule, err := cron.Parse(cronSpec)
	if err != nil {
		return "", 0, err
	}

	// Although the ZSET can guarantee no duplicated items, we still need to check the existing
	// of the job policy to avoid publish duplicated ones to other nodes as we
	// use transaction commands.
	jobPolicy := &PeriodicJobPolicy{
		JobName:       jobName,
		JobParameters: params,
		CronSpec:      cronSpec,
	}
	// Serialize data
	rawJSON, err := jobPolicy.Serialize()
	if err != nil {
		return "", 0, nil
	}

	// Check existing
	// If existing, treat as a succeed submitting and return the exitsing id
	if score, ok := rps.exists(string(rawJSON)); ok {
		// Ignore error
		id, _ := rps.getIDByScore(score)
		return "", 0, errs.ConflictError(id)
	}

	uuid, score := utils.MakePeriodicPolicyUUID()
	// Set back policy ID
	jobPolicy.PolicyID = uuid
	notification := &models.Message{
		Event: EventSchedulePeriodicPolicy,
		Data:  jobPolicy,
	}
	rawJSON2, err := json.Marshal(notification)
	if err != nil {
		return "", 0, err
	}

	// Save to redis db and publish notification via redis transaction
	conn := rps.redisPool.Get()
	defer conn.Close()

	err = conn.Send("MULTI")
	if err != nil {
		return "", 0, err
	}
	err = conn.Send("ZADD", utils.KeyPeriodicPolicy(rps.namespace), score, rawJSON)
	if err != nil {
		return "", 0, err
	}
	err = conn.Send("ZADD", utils.KeyPeriodicPolicyScore(rps.namespace), score, uuid)
	if err != nil {
		return "", 0, err
	}
	err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(rps.namespace), rawJSON2)
	if err != nil {
		return "", 0, err
	}

	if _, err := conn.Do("EXEC"); err != nil {
		return "", 0, err
	}

	return uuid, schedule.Next(time.Now()).Unix(), nil
}

// UnSchedule is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) UnSchedule(cronJobPolicyID string) error {
	if utils.IsEmptyStr(cronJobPolicyID) {
		return errors.New("cron job policy ID is empty")
	}

	score, err := rps.getScoreByID(cronJobPolicyID)
	if err == redis.ErrNil {
		return errs.NoObjectFoundError(err.Error())
	}

	if err != nil {
		return err
	}

	notification := &models.Message{
		Event: EventUnSchedulePeriodicPolicy,
		Data: &PeriodicJobPolicy{
			PolicyID: cronJobPolicyID, // Only ID required
		},
	}

	rawJSON, err := json.Marshal(notification)
	if err != nil {
		return err
	}

	// REM from redis db
	conn := rps.redisPool.Get()
	defer conn.Close()

	err = conn.Send("MULTI")
	if err != nil {
		return err
	}
	err = conn.Send("ZREMRANGEBYSCORE", utils.KeyPeriodicPolicy(rps.namespace), score, score) // Accurately remove the item with the specified score
	if err != nil {
		return err
	}
	err = conn.Send("ZREMRANGEBYSCORE", utils.KeyPeriodicPolicyScore(rps.namespace), score, score) // Remove key score mapping
	if err != nil {
		return err
	}
	err = conn.Send("PUBLISH", utils.KeyPeriodicNotification(rps.namespace), rawJSON)
	if err != nil {
		return err
	}

	_, err = conn.Do("EXEC")

	return err
}

// Load data from zset
func (rps *RedisPeriodicScheduler) Load() error {
	conn := rps.redisPool.Get()
	defer conn.Close()

	// Let's build key score mapping locally first
	bytes, err := redis.MultiBulk(conn.Do("ZRANGE", utils.KeyPeriodicPolicyScore(rps.namespace), 0, -1, "WITHSCORES"))
	if err != nil {
		return err
	}
	keyScoreMap := make(map[int64]string)
	for i, l := 0, len(bytes); i < l; i = i + 2 {
		pid := string(bytes[i].([]byte))
		rawScore := bytes[i+1].([]byte)
		score, err := strconv.ParseInt(string(rawScore), 10, 64)
		if err != nil {
			// Ignore
			continue
		}
		keyScoreMap[score] = pid
	}

	bytes, err = redis.MultiBulk(conn.Do("ZRANGE", utils.KeyPeriodicPolicy(rps.namespace), 0, -1, "WITHSCORES"))
	if err != nil {
		return err
	}

	allPeriodicPolicies := make([]*PeriodicJobPolicy, 0, len(bytes)/2)
	for i, l := 0, len(bytes); i < l; i = i + 2 {
		rawPolicy := bytes[i].([]byte)
		rawScore := bytes[i+1].([]byte)
		policy := &PeriodicJobPolicy{}

		if err := policy.DeSerialize(rawPolicy); err != nil {
			// Ignore error which means the policy data is not valid
			// Only logged
			logger.Warningf("failed to deserialize periodic policy with error:%s; raw data: %s\n", err, rawPolicy)
			continue
		}
		score, err := strconv.ParseInt(string(rawScore), 10, 64)
		if err != nil {
			// Ignore error which means the policy data is not valid
			// Only logged
			logger.Warningf("failed to parse the score of the periodic policy with error:%s\n", err)
			continue
		}

		// Set back the policy ID
		if pid, ok := keyScoreMap[score]; ok {
			policy.PolicyID = pid
		} else {
			// Something wrong, should not be happened
			// ignore here
			continue
		}

		allPeriodicPolicies = append(allPeriodicPolicies, policy)

		logger.Infof("Load periodic job policy %s for job %s: %s", policy.PolicyID, policy.JobName, policy.CronSpec)
	}

	if len(allPeriodicPolicies) > 0 {
		rps.pstore.addAll(allPeriodicPolicies)
	}

	logger.Infof("Load %d periodic job policies", len(allPeriodicPolicies))
	return nil
}

// Clear is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) Clear() error {
	conn := rps.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("ZREMRANGEBYRANK", utils.KeyPeriodicPolicy(rps.namespace), 0, -1)

	return err
}

// AcceptPeriodicPolicy is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) AcceptPeriodicPolicy(policy *PeriodicJobPolicy) error {
	if policy == nil || utils.IsEmptyStr(policy.PolicyID) {
		return errors.New("nil periodic policy")
	}

	rps.pstore.add(policy)

	return nil
}

// RemovePeriodicPolicy is implementation of the same method in period.Interface
func (rps *RedisPeriodicScheduler) RemovePeriodicPolicy(policyID string) *PeriodicJobPolicy {
	if utils.IsEmptyStr(policyID) {
		return nil
	}

	return rps.pstore.remove(policyID)
}

func (rps *RedisPeriodicScheduler) exists(rawPolicy string) (int64, bool) {
	if utils.IsEmptyStr(rawPolicy) {
		return 0, false
	}

	conn := rps.redisPool.Get()
	defer conn.Close()

	count, err := redis.Int64(conn.Do("ZSCORE", utils.KeyPeriodicPolicy(rps.namespace), rawPolicy))
	return count, err == nil
}

func (rps *RedisPeriodicScheduler) getScoreByID(id string) (int64, error) {
	conn := rps.redisPool.Get()
	defer conn.Close()

	return redis.Int64(conn.Do("ZSCORE", utils.KeyPeriodicPolicyScore(rps.namespace), id))
}

func (rps *RedisPeriodicScheduler) getIDByScore(score int64) (string, error) {
	conn := rps.redisPool.Get()
	defer conn.Close()

	ids, err := redis.Strings(conn.Do("ZRANGEBYSCORE", utils.KeyPeriodicPolicyScore(rps.namespace), score, score))
	if err != nil {
		return "", err
	}

	return ids[0], nil
}
