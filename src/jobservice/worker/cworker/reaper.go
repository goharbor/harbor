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

package cworker

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/jobservice/common/rds"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/gomodule/redigo/redis"
)

const (
	maxUpdateDuration       = 24 * time.Hour
	reapLoopInterval        = 1 * time.Hour
	initialReapLoopInterval = 5 * time.Minute
)

// reaper is designed to reap the outdated job stats and web hook
type reaper struct {
	context   context.Context
	namespace string
	pool      *redis.Pool
	lcmCtl    lcm.Controller
	jobTypes  []string
}

// start the reap process
// Non blocking call
func (r *reaper) start() {
	// Start the interval job stats sync loop
	go func() {
		defer logger.Info("Reaper is stopped")

		tm := time.NewTimer(initialReapLoopInterval)
		defer tm.Stop()

		logger.Info("Reaper is started")
		for {
			select {
			case <-tm.C:
				tm.Reset(reapLoopInterval)
				if err := r.syncOutdatedStats(); err != nil {
					// Just log
					logger.Error(err)
				}
			case <-r.context.Done():
				return // Terminated
			}
		}
	}()

	// Start re-enqueue in-progress jobs process.
	// Only run once at the start point.
	go func() {
		// Wait for a short while and then start
		<-time.After(5 * time.Second)
		if err := r.reEnqueueInProgressJobs(); err != nil {
			logger.Error(err)
		}
	}()
}

// reEnqueueInProgressJobs is an enhancement for reaper process of upstream project.
// Mainly fix the issue of failing to re-enqueue the jobs in the dead worker pool.
// This process only needs to be executed once when worker pool is starting.
func (r *reaper) reEnqueueInProgressJobs() error {
	// Debug
	logger.Info("Start: Reap in-progress jobs from the dead pools")
	defer logger.Info("End: Reap in-progress jobs")

	currentPools, err := r.getCurrentWorkerPools()
	if err != nil {
		return errors.Wrap(err, "re-enqueue in progress jobs")
	}

	h := func(k string, v int64) (err error) {
		if v <= 0 {
			// Do nothing
			return nil
		}

		// If the worker pool is in the current pool list, ignore it as the default reap
		// process will cover the re-enqueuing work.
		if currentPools[k] {
			// Do nothing
			return nil
		}

		// Re-enqueue jobs
		if err := r.requeueInProgressJobs(k, r.jobTypes); err != nil {
			return errors.Wrap(err, "in-progress jobs reap handler")
		}

		return nil
	}

	for _, jt := range r.jobTypes {
		lk := rds.KeyJobLockInfo(r.namespace, jt)
		if err := r.scanLocks(lk, h); err != nil {
			// Log error and continue
			logger.Errorf("Re-enqueue in progress jobs error: %v", err)
			continue
		}
	}

	return nil
}

// syncOutdatedStats ensures the job status is correctly updated and
// the related status change hook events are successfully fired.
func (r *reaper) syncOutdatedStats() error {
	// Debug
	logger.Info("Start: reap outdated job stats")
	defer logger.Info("End: reap outdated job stats")

	// Loop all the in progress jobs to check if they're really in progress or
	// status is hung.
	h := func(k string, v int64) (err error) {
		defer func() {
			if errs.IsObjectNotFoundError(err) {
				// As the job stats is lost and we don't have chance to restore it, then directly discard it.
				// Un-track the in-progress record
				if e := r.unTrackInProgress(k); e != nil {
					// Wrap error
					err = errors.Wrap(e, err.Error())
				}
			}

			if err != nil {
				err = errors.Wrap(err, "sync outdated stats handler error")
			}
		}()

		var t job.Tracker
		t, err = r.lcmCtl.Track(k)
		if err != nil {
			return
		}

		// Compare and check if the status and the ACKed status are consistent
		diff := compare(t.Job().Info)
		if diff == 0 {
			// Status and ACKed status are consistent
			if job.Status(t.Job().Info.Status).Final() {
				// Final status
				// The inprogress track record is not valid as everything is done and consistent.
				// It should not happen. However, if it is really happened, we can fix it here.
				if err = r.unTrackInProgress(t.Job().Info.JobID); err != nil {
					return
				}
			} else {
				// Ongoing, check the update timestamp to make sure it is not hung
				if time.Unix(t.Job().Info.UpdateTime, 0).Add(maxUpdateDuration).Before(time.Now()) {
					// Status hung
					// Mark job status to error state
					if err = t.Fail(); err != nil {
						return
					}

					// Log and exit
					logger.Infof("Reaper: mark job %s failed as job is still not finished in 1 day", t.Job().Info.JobID)
				}

				// Exit as it is still a valid ongoing job
			}
		} else if diff > 0 {
			// The hook event of current job status is not ACKed
			// Resend hook event by set the status again
			if err = t.FireHook(); err != nil {
				return
			}

			// Success and exit
			logger.Infof(
				"Reaper: fire hook again for job %s as job status change is not ACKed: %s(rev=%d)",
				t.Job().Info.JobID,
				t.Job().Info.Status,
				t.Job().Info.Revision,
			)
		} else {
			// Current status is outdated, update it with ACKed status.
			if err = t.UpdateStatusWithRetry(job.Status(t.Job().Info.HookAck.Status)); err != nil {
				return
			}

			// Success and exit
			logger.Infof(
				"Reaper: update the status of job %s to the ACKed status: %s(%d)",
				t.Job().Info.JobID,
				t.Job().Info.HookAck.Status,
				t.Job().Info.Revision,
			)
		}

		return nil
	}

	if err := r.scanLocks(rds.KeyJobTrackInProgress(r.namespace), h); err != nil {
		return errors.Wrap(err, "reaper error")
	}

	return nil
}

// scanLocks gets the lock info from the specified key by redis scan way
func (r *reaper) scanLocks(key string, handler func(k string, v int64) error) error {
	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("Failed to close redis connection: %v", err)
		}
	}()

	var cursor int64

	for {
		reply, err := redis.Values(conn.Do("HSCAN", key, cursor, "MATCH", "*", "COUNT", 100))
		if err != nil {
			return errors.Wrap(err, "scan locks")
		}

		if len(reply) != 2 {
			return errors.New("expect 2 elements returned")
		}

		// Get next cursor
		cursor, err = strconv.ParseInt(string(reply[0].([]uint8)), 10, 16)
		if err != nil {
			return errors.Wrap(err, "scan locks")
		}

		if values, ok := reply[1].([]interface{}); ok {
			for i := 0; i < len(values); i += 2 {
				k := string(values[i].([]uint8))
				lc, err := strconv.ParseInt(string(values[i+1].([]uint8)), 10, 16)
				if err != nil {
					// Ignore and continue
					logger.Errorf("Malformed lock object for %s: %v", k, err)
					continue
				}
				// Call handler to handle the data
				if err := handler(k, lc); err != nil {
					// Log and ignore the error
					logger.Errorf("Failed to call reap handler: %v", err)
				}
			}
		}

		// Check if we have reached the end
		if cursor == 0 {
			return nil
		}
	}
}

// unTrackInProgress del the key in the progress track queue
func (r *reaper) unTrackInProgress(jobID string) error {
	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("Failed to close redis connection: %s: %v", "untrack in progress job", err)
		}
	}()

	_, err := conn.Do("HDEL", rds.KeyJobTrackInProgress(r.namespace), jobID)
	if err != nil {
		return errors.Wrap(err, "untrack in progress record")
	}

	return nil
}

// getCurrentWorkerPools returns the IDs of the current worker pools
func (r *reaper) getCurrentWorkerPools() (map[string]bool, error) {
	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("Failed to close redis connection: %s : %v", "get current worker pools", err)
		}
	}()

	// Get the current worker pools
	workerPoolIDs, err := redis.Strings(conn.Do("SMEMBERS", rds.KeyWorkerPools(r.namespace)))
	if err != nil {
		return nil, errors.Wrap(err, "get current workpools")
	}

	m := make(map[string]bool)
	for _, id := range workerPoolIDs {
		m[id] = true
	}

	return m, nil
}

func (r *reaper) requeueInProgressJobs(poolID string, jobTypes []string) error {
	numKeys := len(jobTypes)
	redisRequeueScript := rds.RedisLuaReenqueueScript(numKeys)
	var scriptArgs = make([]interface{}, 0, numKeys+1)

	for _, jobType := range jobTypes {
		// pops from in progress, push into job queue and decrement the queue lock
		scriptArgs = append(
			scriptArgs,
			rds.KeyInProgressQueue(r.namespace, jobType, poolID),
			rds.KeyJobs(r.namespace, jobType),
			rds.KeyJobLock(r.namespace, jobType),
			rds.KeyJobLockInfo(r.namespace, jobType),
		) // KEYS[1-4 * N]
	}
	scriptArgs = append(scriptArgs, poolID) // ARGV[1]

	conn := r.pool.Get()
	defer func() {
		if err := conn.Close(); err != nil {
			logger.Errorf("Failed to close redis connection: %s : %s", "re enqueue in-progress jobs", err)
		}
	}()

	// Keep moving jobs until all queues are empty
	for {
		values, err := redis.Values(redisRequeueScript.Do(conn, scriptArgs...))
		if err == redis.ErrNil {
			return nil
		} else if err != nil {
			return err
		}

		if len(values) != 3 {
			return fmt.Errorf("need 3 elements back")
		}
	}
}

// compare the status and the status in the ack
// 0: status == ack.status
// >0: status > ack.status
// <0: status < ack.status
//
// compare based on:
// revision:status_code:check_in
func compare(j *job.StatsInfo) int {
	// No ack existing
	if j.HookAck == nil {
		return 1
	}

	// Compare revision
	rev := j.Revision - j.HookAck.Revision
	if rev != 0 {
		return (int)(rev)
	}

	// Revision is same, then compare the status
	switch {
	case job.Status(j.Status).Before(job.Status(j.HookAck.Status)):
		return -1
	case job.Status(j.Status).After(job.Status(j.HookAck.Status)):
		return 1
	}

	// Revision and status are same, then compare the checkin
	return (int)(j.CheckInAt - j.HookAck.CheckInAt)
}
