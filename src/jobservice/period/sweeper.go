// Copyright 2018 The Harbor Authors. All rights reserved.

package period

import (
	"fmt"
	"time"

	"github.com/gocraft/work"

	"github.com/garyburd/redigo/redis"
	"github.com/vmware/harbor/src/jobservice/logger"
	"github.com/vmware/harbor/src/jobservice/utils"
)

//Sweeper take charge of clearing the outdated data such as scheduled jobs etc..
//Currently, only used in redis worker pool.
type Sweeper struct {
	redisPool *redis.Pool
	client    *work.Client
	namespace string
}

//NewSweeper is constructor of Sweeper.
func NewSweeper(namespace string, pool *redis.Pool, client *work.Client) *Sweeper {
	return &Sweeper{
		namespace: namespace,
		redisPool: pool,
		client:    client,
	}
}

//ClearOutdatedScheduledJobs clears the outdated scheduled jobs.
//Try best to do
func (s *Sweeper) ClearOutdatedScheduledJobs() error {
	//Check if other workpool has done the action
	conn := s.redisPool.Get()
	defer conn.Close()

	//Lock
	r, err := conn.Do("SET", utils.KeyPeriodicLock(s.namespace), time.Now().Unix(), "EX", 30, "NX")
	defer func() {
		//Make sure it can be unlocked if it is not expired yet
		conn.Do("DEL", utils.KeyPeriodicLock(s.namespace))
	}()
	if err != nil {
		return err
	}

	if r == nil {
		//Action is already locked by other workerpool
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

	//Unlock
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
