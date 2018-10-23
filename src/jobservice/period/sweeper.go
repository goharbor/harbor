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
	"fmt"
	"time"

	"github.com/gocraft/work"

	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
)

// Sweeper take charge of clearing the outdated data such as scheduled jobs etc..
// Currently, only used in redis worker pool.
type Sweeper struct {
	redisPool *redis.Pool
	client    *work.Client
	namespace string
}

// NewSweeper is constructor of Sweeper.
func NewSweeper(namespace string, pool *redis.Pool, client *work.Client) *Sweeper {
	return &Sweeper{
		namespace: namespace,
		redisPool: pool,
		client:    client,
	}
}

// ClearOutdatedScheduledJobs clears the outdated scheduled jobs.
// Try best to do
func (s *Sweeper) ClearOutdatedScheduledJobs() error {
	// Check if other workpool has done the action
	conn := s.redisPool.Get()
	defer conn.Close()

	// Lock
	r, err := conn.Do("SET", utils.KeyPeriodicLock(s.namespace), time.Now().Unix(), "EX", 30, "NX")
	defer func() {
		// Make sure it can be unlocked if it is not expired yet
		if _, err := conn.Do("DEL", utils.KeyPeriodicLock(s.namespace)); err != nil {
			logger.Errorf("Unlock key '%s' failed with error: %s\n", utils.KeyPeriodicLock(s.namespace), err.Error())
		}
	}()
	if err != nil {
		return err
	}

	if r == nil {
		// Action is already locked by other workerpool
		logger.Info("Ignore clear outdated scheduled jobs")
		return nil
	}

	nowEpoch := time.Now().Unix()
	jobScores, err := utils.GetZsetByScore(s.redisPool, utils.RedisKeyScheduled(s.namespace), []int64{0, nowEpoch})
	if err != nil {
		return err
	}

	allErrors := make([]error, 0)
	for _, jobScore := range jobScores {
		j, err := utils.DeSerializeJob(jobScore.JobBytes)
		if err != nil {
			allErrors = append(allErrors, err)
			continue
		}

		if err = s.client.DeleteScheduledJob(jobScore.Score, j.ID); err != nil {
			allErrors = append(allErrors, err)
		}

		logger.Infof("Clear outdated scheduled job: %s run at %#v\n", j.ID, time.Unix(jobScore.Score, 0).String())
	}

	// Unlock
	if len(allErrors) == 0 {
		return nil
	}

	if len(allErrors) == 1 {
		return allErrors[0]
	}

	errorSummary := allErrors[0].Error()
	for index, e := range allErrors {
		if index == 0 {
			continue
		}

		errorSummary = fmt.Sprintf("%s, %s", errorSummary, e)
	}
	return fmt.Errorf("%s", errorSummary)
}
