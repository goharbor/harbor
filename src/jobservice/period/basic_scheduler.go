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
	"math/rand"
	"time"

	"github.com/gocraft/work"
	"github.com/gomodule/redigo/redis"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
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

	pid := time.Now().Unix() + rand.Int63n(10)

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

	conn := bs.pool.Get()
	defer func() {
		_ = conn.Close()
	}()

	numericID, err := bs.locatePolicy(policyID, conn)
	if err != nil {
		return err
	}

	// Get downstream executions of the periodic job
	// And clear these executions
	// This is a try best action, its failure will not cause the unschedule action failed.
	// Failure errors will be only logged here
	eKey := rds.KeyUpstreamJobAndExecutions(bs.namespace, policyID)
	if eIDs, err := getPeriodicExecutions(conn, eKey); err != nil {
		logger.Errorf("Get executions for periodic job %s error: %s", lib.TrimLineBreaks(policyID), err)
	} else {
		if len(eIDs) == 0 {
			logger.Debugf("no stopped executions: %s", lib.TrimLineBreaks(policyID))
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
				// Please pay attention here, the job ID used in the scheduled job queue is
				// the ID of the periodic job (policy).
				if err := bs.client.DeleteScheduledJob(e.Info.RunAt, policyID); err != nil {
					logger.Errorf("Delete scheduled job %s error: %s", eID, err)
				}
			}

			// Mark job status to stopped to block execution.
			// The executions here should not be in the final states,
			// double confirmation: only stop the can-stop ones.
			if job.RunningStatus.After(job.Status(e.Info.Status)) || job.RunningStatus.Equal(job.Status(e.Info.Status)) {
				if err := eTracker.Stop(); err != nil {
					logger.Errorf("Stop execution %s error: %s", eID, err)
				} else {
					logger.Debugf("Stop execution %q of periodic job %s", eID, policyID)
				}
			}
		}
	}

	// REM from redis db
	removed, err := removePolicy(conn, bs.namespace, policyID, numericID)
	if err != nil {
		return errors.Wrap(err, "unschedule periodic job error")
	}

	if removed == 0 {
		logger.Warningf("No periodic job with ID=%s and numeric ID=%d removed from the periodic job policy set", lib.TrimLineBreaks(policyID), numericID)
	}

	return nil
}

// removePolicy removes the policy with the given ID from the periodic policy set.
// Policies scheduled within a few seconds of each other can get the same score,
// so only the member of this policy is removed, not every member with its score.
func removePolicy(conn redis.Conn, namespace string, policyID string, numericID int64) (int64, error) {
	key := rds.KeyPeriodicPolicy(namespace)
	members, err := redis.ByteSlices(conn.Do("ZRANGEBYSCORE", key, numericID, numericID))
	if err != nil {
		return 0, err
	}

	var removed int64
	for _, rawJSON := range members {
		p := &Policy{}
		if err := p.DeSerialize(rawJSON); err != nil || p.ID != policyID {
			continue
		}
		n, err := redis.Int64(conn.Do("ZREM", key, rawJSON))
		if err != nil {
			return removed, err
		}
		removed += n
	}

	return removed, nil
}

// Locate the policy and return the numeric ID.
// First locate policy by tracker, then locate by looping the policy list in case the job stats data is lost
func (bs *basicScheduler) locatePolicy(policyID string, conn redis.Conn) (int64, error) {
	// Handle the corresponding job stats of the given periodic job first.
	tracker, err := bs.ctl.Track(policyID)
	if err != nil {
		// If error is not found error, then switch to the backup approach
		if errs.IsObjectNotFoundError(err) {
			// Loop the policy list to get the policy data
			pl, err := Load(bs.namespace, conn)
			if err != nil {
				return -1, err
			}

			for _, p := range pl {
				if p.ID == policyID && p.NumericID > 0 {
					// Found the policy in the queue and return the numeric ID
					return p.NumericID, nil
				}
			}
		}

		// Still not found or other errors
		return -1, err
	}

	// Switch the job stats to stopped if the job stats existing
	// Should not block the next clear action
	if err := tracker.Stop(); err != nil {
		logger.Errorf("Stop periodic job %s failed with error: %s", lib.TrimLineBreaks(policyID), err)
	}

	return tracker.NumericID()
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
	args := []any{key, 0, "+inf"}

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
