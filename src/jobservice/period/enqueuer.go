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
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/opm"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/gomodule/redigo/redis"
	"github.com/robfig/cron"
)

const (
	periodicEnqueuerSleep   = 2 * time.Minute
	periodicEnqueuerHorizon = 4 * time.Minute
)

type periodicEnqueuer struct {
	namespace        string
	pool             *redis.Pool
	policyStore      *periodicJobPolicyStore
	stopChan         chan struct{}
	doneStoppingChan chan struct{}
	statsManager     opm.JobStatsManager
}

func newPeriodicEnqueuer(namespace string, pool *redis.Pool, policyStore *periodicJobPolicyStore, statsManager opm.JobStatsManager) *periodicEnqueuer {
	return &periodicEnqueuer{
		namespace:        namespace,
		pool:             pool,
		policyStore:      policyStore,
		statsManager:     statsManager,
		stopChan:         make(chan struct{}),
		doneStoppingChan: make(chan struct{}),
	}
}

func (pe *periodicEnqueuer) start() {
	go pe.loop()
	logger.Info("Periodic enqueuer is started")
}

func (pe *periodicEnqueuer) stop() {
	pe.stopChan <- struct{}{}
	<-pe.doneStoppingChan
}

func (pe *periodicEnqueuer) loop() {
	defer func() {
		logger.Info("Periodic enqueuer is stopped")
	}()
	// Begin reaping periodically
	timer := time.NewTimer(periodicEnqueuerSleep + time.Duration(rand.Intn(30))*time.Second)
	defer timer.Stop()

	if pe.shouldEnqueue() {
		err := pe.enqueue()
		if err != nil {
			logger.Errorf("periodic_enqueuer.loop.enqueue:%s\n", err)
		}
	} else {
		logger.Debug("Enqueue condition not matched, do nothing.")
	}

	for {
		select {
		case <-pe.stopChan:
			pe.doneStoppingChan <- struct{}{}
			return
		case <-timer.C:
			timer.Reset(periodicEnqueuerSleep + time.Duration(rand.Intn(30))*time.Second)
			if pe.shouldEnqueue() {
				err := pe.enqueue()
				if err != nil {
					logger.Errorf("periodic_enqueuer.loop.enqueue:%s\n", err)
				}
			} else {
				logger.Debug("Enqueue condition not matched, do nothing.")
			}
		}
	}
}

func (pe *periodicEnqueuer) enqueue() error {
	now := time.Now().Unix()

	conn := pe.pool.Get()
	defer conn.Close()

	// Set last periodic enqueue timestamp in advance to avoid duplicated enqueue actions
	if _, err := conn.Do("SET", utils.RedisKeyLastPeriodicEnqueue(pe.namespace), now); err != nil {
		return err
	}

	// Avoid schedule in the same time.
	lockerKey := fmt.Sprintf("%s:%s", utils.KeyPeriod(pe.namespace), "lock")
	lockerID := utils.MakeIdentifier()

	// Acquire a locker with 30s expiring time
	if err := acquireLock(conn, lockerKey, lockerID, 30); err != nil {
		return err
	}
	defer func() {
		if err := releaseLock(conn, lockerKey, lockerID); err != nil {
			logger.Errorf("release enqueue locker failed: %s", err)
		}
	}()

	nowTime := time.Unix(now, 0)
	horizon := nowTime.Add(periodicEnqueuerHorizon)

	for _, pl := range pe.policyStore.list() {
		schedule, err := cron.Parse(pl.CronSpec)
		if err != nil {
			// The cron spec should be already checked at top components.
			// Just in cases, if error occurred, ignore it
			logger.Errorf("[Ignore] Invalid corn spec in periodic policy %s %s: %s", pl.JobName, pl.PolicyID, err)
			continue
		}

		executions := []string{}
		for t := schedule.Next(nowTime); t.Before(horizon); t = schedule.Next(t) {
			epoch := t.Unix()
			scheduledExecutionID := utils.MakeIdentifier()
			executions = append(executions, scheduledExecutionID)

			// Create an execution (job) based on the periodic job template (policy)
			job := &work.Job{
				Name: pl.JobName,
				ID:   scheduledExecutionID,

				// This is technically wrong, but this lets the bytes be identical for the same periodic job instance.
				// If we don't do this, we'd need to use a different approach -- probably giving each periodic job its own
				// history of the past 100 periodic jobs, and only scheduling a job if it's not in the history.
				EnqueuedAt: epoch,
				Args:       pl.JobParameters, // Pass parameters to scheduled job here
			}

			rawJSON, err := utils.SerializeJob(job)
			if err != nil {
				return err
			}

			// Place the time slots for the job (policy)
			// If the slot is already there, error will be returned.
			expireTime := (epoch - nowTime.Unix()) + 5
			slot := fmt.Sprintf("%s:%s@%d", utils.KeyPeriodicJobTimeSlots(pe.namespace), pl.PolicyID, epoch)
			if err := placeSlot(conn, slot, epoch, expireTime); err != nil {
				// Logged and continue
				logger.Errorf("Failed to place time slot '%s@%d': %s", pl.PolicyID, epoch, err)
				continue
			}

			_, err = conn.Do("ZADD", utils.RedisKeyScheduled(pe.namespace), epoch, rawJSON)
			if err != nil {
				return err
			}

			logger.Infof("Schedule job %s:%s for policy %s at %d\n", job.Name, job.ID, pl.PolicyID, epoch)

			// Try to save the stats of new scheduled execution (job).
			pe.createExecution(pl.PolicyID, pl.JobName, scheduledExecutionID, epoch)

			// Get web hook from the periodic job (policy)
			webHookURL, err := pe.statsManager.GetHook(pl.PolicyID)
			if err == nil {
				// Register hook for the execution
				if err := pe.statsManager.RegisterHook(scheduledExecutionID, webHookURL, false); err != nil {
					// Just logged
					logger.Errorf("Failed to register web hook '%s' for periodic job (execution) '%s' with error: %s", webHookURL, scheduledExecutionID, err)
				}
			} else {
				// Just a warning
				logger.Warningf("Failed to retrieve web hook for periodic job (policy) %s: %s", pl.PolicyID, err)
			}
		}
		// Link the upstream job (policy) with the created executions
		if len(executions) > 0 {
			if err := pe.statsManager.AttachExecution(pl.PolicyID, executions...); err != nil {
				// Just logged it
				logger.Errorf("Link upstream job with executions failed: %s", err)
			}
		}
		// Directly use redis conn to update the periodic job (policy) status
		// Do not care the result
		conn.Do("HMSET", utils.KeyJobStats(pe.namespace, pl.PolicyID), "status", job.JobStatusScheduled, "update_time", time.Now().Unix())
	}

	return nil
}

func (pe *periodicEnqueuer) createExecution(upstreamJobID, upstreamJobName, executionID string, runAt int64) {
	execution := models.JobStats{
		Stats: &models.JobStatData{
			JobID:         executionID,
			JobName:       upstreamJobName,
			Status:        job.JobStatusPending,
			JobKind:       job.JobKindScheduled,
			EnqueueTime:   time.Now().Unix(),
			UpdateTime:    time.Now().Unix(),
			RefLink:       fmt.Sprintf("/api/v1/jobs/%s", executionID),
			RunAt:         runAt,
			UpstreamJobID: upstreamJobID,
		},
	}

	pe.statsManager.Save(execution)
}

func (pe *periodicEnqueuer) shouldEnqueue() bool {
	conn := pe.pool.Get()
	defer conn.Close()

	lastEnqueue, err := redis.Int64(conn.Do("GET", utils.RedisKeyLastPeriodicEnqueue(pe.namespace)))
	if err == redis.ErrNil {
		return true
	} else if err != nil {
		logger.Errorf("periodic_enqueuer.should_enqueue:%s\n", err)
		return true
	}

	return lastEnqueue < (time.Now().Unix() - int64(periodicEnqueuerSleep/time.Minute))
}

func placeSlot(conn redis.Conn, key string, value interface{}, expireTime int64) error {
	args := []interface{}{key, value, "NX", "EX", expireTime}
	res, err := conn.Do("SET", args...)
	if err != nil {
		return err
	}
	// Existing, the value can not be overrid
	if res == nil {
		return fmt.Errorf("key %s is already set with value %v", key, value)
	}

	return nil
}

func acquireLock(conn redis.Conn, lockerKey string, lockerID string, expireTime int64) error {
	return placeSlot(conn, lockerKey, lockerID, expireTime)
}

func releaseLock(conn redis.Conn, lockerKey string, lockerID string) error {
	theID, err := redis.String(conn.Do("GET", lockerKey))
	if err != nil {
		return err
	}

	if theID == lockerID {
		_, err := conn.Do("DEL", lockerKey)
		return err
	}

	return errors.New("locker ID mismatch")
}
