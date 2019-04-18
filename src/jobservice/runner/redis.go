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

package runner

import (
	"github.com/goharbor/harbor/src/jobservice/job/impl"
	"runtime"

	"fmt"
	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/pkg/errors"
	"time"
)

// RedisJob is a job wrapper to wrap the job.Interface to the style which can be recognized by the redis worker.
type RedisJob struct {
	job     interface{}    // the real job implementation
	context *env.Context   // context
	ctl     lcm.Controller // life cycle controller
}

// NewRedisJob is constructor of RedisJob
func NewRedisJob(job interface{}, ctx *env.Context, ctl lcm.Controller) *RedisJob {
	return &RedisJob{
		job:     job,
		context: ctx,
		ctl:     ctl,
	}
}

// Run the job
func (rj *RedisJob) Run(j *work.Job) (err error) {
	var (
		runningJob  job.Interface
		execContext job.Context
		tracker     lcm.Tracker
		markStopped *bool = bp(false)
	)

	// Defer to log the exit result
	defer func() {
		if !*markStopped {
			if err == nil {
				logger.Infof("Job '%s:%s' exit with success", j.Name, j.ID)
			} else {
				// log error
				logger.Errorf("Job '%s:%s' exit with error: %s\n", j.Name, j.ID, err)
			}
		}
	}()

	// Track the running job now
	jID := j.ID
	if isPeriodicJobExecution(j) {
		jID = fmt.Sprintf("%s@%d", j.ID, j.EnqueuedAt)
	}

	if tracker, err = rj.ctl.Track(jID); err != nil {
		// As tracker creation failed, there is no way to mark the job status change.
		// Also a non nil error return consumes a fail. If all retries are failed here,
		// it will cause the job to be zombie one (pending forever).
		// Here we will avoid the job to consume a fail and let it retry again and again.
		// However, to avoid a forever retry, we will check the FailedAt timestamp.
		now := time.Now().Unix()
		if j.FailedAt == 0 || now-j.FailedAt < 2*24*3600 {
			j.Fails--
		}

		return
	}

	if job.RunningStatus.Compare(job.Status(tracker.Job().Info.Status)) >= 0 {
		// Probably jobs has been stopped by directly mark status to stopped.
		// Directly exit and no retry
		markStopped = bp(true)
		return nil
	}

	// Defer to switch status
	defer func() {
		// Switch job status based on the returned error.
		// The err happened here should not override the job run error, just log it.
		if err != nil {
			if errs.IsJobStoppedError(err) {
				if er := tracker.Stop(); er != nil {
					logger.Errorf("Mark job status to stopped error: %s", err)
				}
			} else {
				if er := tracker.Fail(); er != nil {
					logger.Errorf("Mark job status to failure error: %s", err)
				}
			}

			return
		}

		// Mark job status to success.
		if er := tracker.Succeed(); er != nil {
			logger.Errorf("Mark job status to success error: %s", err)
		}
	}()

	// Defer to handle runtime error
	defer func() {
		if r := recover(); r != nil {
			// Log the stack
			buf := make([]byte, 1<<16)
			size := runtime.Stack(buf, false)
			err = errors.Errorf("runtime error: %s; stack: %s", r, buf[0:size])
			logger.Errorf("Run job %s:%s error: %s", j.Name, j.ID, err)
		}
	}()

	// Build job context
	if rj.context.JobContext == nil {
		rj.context.JobContext = impl.NewDefaultContext(rj.context.SystemContext)
		if execContext, err = rj.context.JobContext.Build(tracker); err != nil {
			return
		}
	}
	// Defer to close logger stream
	defer func() {
		// Close open io stream first
		if closer, ok := execContext.GetLogger().(logger.Closer); ok {
			if er := closer.Close(); er != nil {
				logger.Errorf("Close job logger failed: %s", er)
			}
		}
	}()

	// Wrap job
	runningJob = Wrap(rj.job)
	// Set status to run
	if err = tracker.Run(); err != nil {
		return
	}
	// Run the job
	err = runningJob.Run(execContext, j.Args)
	// Handle retry
	rj.retry(runningJob, j, (err != nil && errs.IsJobStoppedError(err)))
	// Handle periodic job execution
	if isPeriodicJobExecution(j) {
		if er := tracker.PeriodicExecutionDone(); er != nil {
			// Just log it
			logger.Error(er)
		}
	}

	return
}

func (rj *RedisJob) retry(j job.Interface, wj *work.Job, stopped bool) {
	if stopped || !j.ShouldRetry() {
		// Cancel retry immediately
		// Make it big enough to avoid retrying
		wj.Fails = 10000000000
		return
	}
}

func isPeriodicJobExecution(j *work.Job) bool {
	if isPeriodic, ok := j.Args["_job_kind_periodic_"]; ok {
		if isPeriodicV, yes := isPeriodic.(bool); yes && isPeriodicV {
			return true
		}
	}

	return false
}

func bp(b bool) *bool {
	return &b
}
