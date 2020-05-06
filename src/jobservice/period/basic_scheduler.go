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
	"time"

	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

// basicScheduler manages the periodic scheduling policies.
type basicScheduler struct {
	context   context.Context
	pool      *redis.Pool
	namespace string
	enqueuer  *enqueuer
	client    *work.Client
	ctl       lcm.Controller
}

// NewScheduler is constructor of basicScheduler
func NewScheduler(ctx context.Context, namespace string, pool *redis.Pool, ctl lcm.Controller) Scheduler {
	return &basicScheduler{
		context:   ctx,
		pool:      pool,
		namespace: namespace,
		enqueuer:  newEnqueuer(ctx, namespace, pool, ctl),
		client:    work.NewClient(namespace, pool),
		ctl:       ctl,
	}
}

// Start the periodic scheduling process
func (bs *basicScheduler) Start() {
	// Run once clean
	// Try best to do
	go bs.clearDirtyJobs()

	// start enqueuer
	bs.enqueuer.start()
}

// Schedule is implementation of the same method in period.Interface
func (bs *basicScheduler) Schedule(p *Policy) (int64, error) {
	if p == nil {
		return -1, errors.New("bad policy object: nil")
	}

	if err := p.Validate(); err != nil {
		return -1, err
	}

	conn := bs.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	// Do the 1st round of enqueuing
	bs.enqueuer.scheduleNextJobs(p, conn)

	// Serialize data
	rawJSON, err := p.Serialize()
	if err != nil {
		return -1, err
	}

	pid := time.Now().Unix()

	// Save to redis db
	if _, err := conn.Do("ZADD", rds.KeyPeriodicPolicy(bs.namespace), pid, rawJSON); err != nil {
		return -1, err
	}

	return pid, nil
}

// UnSchedule is implementation of the same method in period.Interface
func (bs *basicScheduler) UnSchedule(policyID string) error {
	if utils.IsEmptyStr(policyID) {
		return errors.New("bad periodic job ID: nil")
	}

	tracker, err := bs.ctl.Track(policyID)
	if err != nil {
		return err
	}

	// If errors occurred when getting the numeric ID of periodic job,
	// may be because the specified job is not a valid periodic job.
	numericID, err := tracker.NumericID()
	if err != nil {
		return err
	}

	conn := bs.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	// Get the un-scheduling policy object
	bytes, err := redis.Values(conn.Do("ZRANGEBYSCORE", rds.KeyPeriodicPolicy(bs.namespace), numericID, numericID))
	if err != nil {
		return err
	}

	p := &Policy{}
	if len(bytes) > 0 {
		if rawPolicy, ok := bytes[0].([]byte); ok {
			if err := p.DeSerialize(rawPolicy); err != nil {
				return err
			}
		}
	}

	if utils.IsEmptyStr(p.ID) {
		// Deserialize failed
		return errors.Errorf("no valid periodic job policy found: %s:%d", policyID, numericID)
	}

	// REM from redis db
	// Accurately remove the item with the specified score
	if _, err := conn.Do("ZREMRANGEBYSCORE", rds.KeyPeriodicPolicy(bs.namespace), numericID, numericID); err != nil {
		return err
	}

	// Expire periodic job stats
	if err := tracker.Expire(); err != nil {
		logger.Error(err)
	}

	// Switch the job stats to stopped
	// Should not block the next clear action
	err = tracker.Stop()

	// Get downstream executions of the periodic job
	// And clear these executions
	// This is a try best action, its failure will not cause the unschedule action failed.
	// Failure errors will be only logged here
	eKey := rds.KeyUpstreamJobAndExecutions(bs.namespace, policyID)
	if eIDs, err := getPeriodicExecutions(conn, eKey); err != nil {
		logger.Errorf("Get executions for periodic job %s error: %s", policyID, err)
	} else {
		if len(eIDs) == 0 {
			logger.Debugf("no stopped executions: %s", policyID)
		}
		for _, eID := range eIDs {
			eTracker, err := bs.ctl.Track(eID)
			if err != nil {
				logger.Errorf("Track execution %s error: %s", eID, err)
				continue
			}

			e := eTracker.Job()
			// Only need to care the pending and running ones
			// Do clear
			if job.ScheduledStatus == job.Status(e.Info.Status) {
				// Please pay attention here, the job ID used in the scheduled jon queue is
				// the ID of the periodic job (policy).
				if err := bs.client.DeleteScheduledJob(e.Info.RunAt, policyID); err != nil {
					logger.Errorf("Delete scheduled job %s error: %s", eID, err)
				}
			}

			// Mark job status to stopped to block execution.
			// The executions here should not be in the final states,
			// double confirmation: only stop the stopped ones.
			if job.RunningStatus.Compare(job.Status(e.Info.Status)) >= 0 {
				if err := eTracker.Stop(); err != nil {
					logger.Errorf("Stop execution %s error: %s", eID, err)
				}
			}
		}
	}

	return err
}

// Clear all the dirty jobs
// A scheduled job will be marked as dirty job only if the enqueued timestamp has expired a horizon.
// This is a try best action
func (bs *basicScheduler) clearDirtyJobs() {
	conn := bs.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	nowEpoch := time.Now().Unix()
	scope := nowEpoch - int64(enqueuerHorizon/time.Minute)*60

	jobScores, err := rds.GetZsetByScore(conn, rds.RedisKeyScheduled(bs.namespace), []int64{0, scope})
	if err != nil {
		logger.Errorf("Get dirty jobs error: %s", err)
		return
	}

	for _, jobScore := range jobScores {
		j, err := utils.DeSerializeJob(jobScore.JobBytes)
		if err != nil {
			logger.Errorf("Deserialize dirty job error: %s", err)
			continue
		}

		if err = bs.client.DeleteScheduledJob(jobScore.Score, j.ID); err != nil {
			logger.Errorf("Remove dirty scheduled job error: %s", err)
		} else {
			logger.Debugf("Remove dirty scheduled job: %s run at %#v", j.ID, time.Unix(jobScore.Score, 0).String())
		}
	}
}

// Get relevant executions for the periodic job
func getPeriodicExecutions(conn redis.Conn, key string) ([]string, error) {
	args := []interface{}{key, 0, "+inf"}

	list, err := redis.Values(conn.Do("ZRANGEBYSCORE", args...))
	if err != nil {
		return nil, err
	}

	results := make([]string, 0)
	for _, item := range list {
		if eID, ok := item.([]byte); ok {
			results = append(results, string(eID))
		}
	}

	return results, nil
}
