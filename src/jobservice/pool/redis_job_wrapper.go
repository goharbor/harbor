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

package pool

import (
	"errors"
	"fmt"
	"runtime"
	"time"

	"github.com/goharbor/harbor/src/jobservice/job/impl"

	"github.com/gocraft/work"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/opm"
	"github.com/goharbor/harbor/src/jobservice/utils"
)

// RedisJob is a job wrapper to wrap the job.Interface to the style which can be recognized by the redis pool.
type RedisJob struct {
	job          interface{}         // the real job implementation
	context      *env.Context        // context
	statsManager opm.JobStatsManager // job stats manager
}

// NewRedisJob is constructor of RedisJob
func NewRedisJob(j interface{}, ctx *env.Context, statsManager opm.JobStatsManager) *RedisJob {
	return &RedisJob{
		job:          j,
		context:      ctx,
		statsManager: statsManager,
	}
}

// Run the job
func (rj *RedisJob) Run(j *work.Job) error {
	var (
		cancelled          = false
		buildContextFailed = false
		runningJob         job.Interface
		err                error
		execContext        env.JobContext
	)

	defer func() {
		if err == nil {
			logger.Infof("Job '%s:%s' exit with success", j.Name, j.ID)
			return // nothing need to do
		}

		// log error
		logger.Errorf("Job '%s:%s' exit with error: %s\n", j.Name, j.ID, err)

		if buildContextFailed || rj.shouldDisableRetry(runningJob, j, cancelled) {
			j.Fails = 10000000000 // Make it big enough to avoid retrying
			now := time.Now().Unix()
			go func() {
				timer := time.NewTimer(2 * time.Second) // make sure the failed job is already put into the dead queue
				defer timer.Stop()

				<-timer.C

				rj.statsManager.DieAt(j.ID, now)
			}()
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("Runtime error: %s", r)

			// Log the stack
			buf := make([]byte, 1<<16)
			size := runtime.Stack(buf, false)
			logger.Errorf("Runtime error happened when executing job %s:%s: %s", j.Name, j.ID, buf[0:size])

			// record runtime error status
			rj.jobFailed(j.ID)
		}
	}()

	// Wrap job
	runningJob = Wrap(rj.job)

	execContext, err = rj.buildContext(j)
	if err != nil {
		buildContextFailed = true
		goto FAILED // no need to retry
	}

	defer func() {
		// Close open io stream first
		if closer, ok := execContext.GetLogger().(logger.Closer); ok {
			closer.Close()
		}
	}()

	// Start to run
	rj.jobRunning(j.ID)

	// Inject data
	err = runningJob.Run(execContext, j.Args)

	// update the proper status
	if err == nil {
		rj.jobSucceed(j.ID)
		return nil
	}

	if errs.IsJobStoppedError(err) {
		rj.jobStopped(j.ID)
		return nil // no need to put it into the dead queue for resume
	}

	if errs.IsJobCancelledError(err) {
		rj.jobCancelled(j.ID)
		cancelled = true
		return err // need to resume
	}

FAILED:
	rj.jobFailed(j.ID)
	return err
}

func (rj *RedisJob) jobRunning(jobID string) {
	rj.statsManager.SetJobStatus(jobID, job.JobStatusRunning)
}

func (rj *RedisJob) jobFailed(jobID string) {
	rj.statsManager.SetJobStatus(jobID, job.JobStatusError)
}

func (rj *RedisJob) jobStopped(jobID string) {
	rj.statsManager.SetJobStatus(jobID, job.JobStatusStopped)
}

func (rj *RedisJob) jobCancelled(jobID string) {
	rj.statsManager.SetJobStatus(jobID, job.JobStatusCancelled)
}

func (rj *RedisJob) jobSucceed(jobID string) {
	rj.statsManager.SetJobStatus(jobID, job.JobStatusSuccess)
}

func (rj *RedisJob) buildContext(j *work.Job) (env.JobContext, error) {
	// Build job execution context
	jData := env.JobData{
		ID:        j.ID,
		Name:      j.Name,
		Args:      j.Args,
		ExtraData: make(map[string]interface{}),
	}

	checkOPCmdFuncFactory := func(jobID string) job.CheckOPCmdFunc {
		return func() (string, bool) {
			cmd, err := rj.statsManager.CtlCommand(jobID)
			if err != nil {
				return "", false
			}
			return cmd, true
		}
	}

	jData.ExtraData["opCommandFunc"] = checkOPCmdFuncFactory(j.ID)

	checkInFuncFactory := func(jobID string) job.CheckInFunc {
		return func(message string) {
			rj.statsManager.CheckIn(jobID, message)
		}
	}

	jData.ExtraData["checkInFunc"] = checkInFuncFactory(j.ID)

	launchJobFuncFactory := func(jobID string) job.LaunchJobFunc {
		funcIntf := rj.context.SystemContext.Value(utils.CtlKeyOfLaunchJobFunc)
		return func(jobReq models.JobRequest) (models.JobStats, error) {
			launchJobFunc, ok := funcIntf.(job.LaunchJobFunc)
			if !ok {
				return models.JobStats{}, errors.New("no launch job func provided")
			}

			jobName := ""
			if jobReq.Job != nil {
				jobName = jobReq.Job.Name
			}
			if j.Name == jobName {
				return models.JobStats{}, errors.New("infinite job creating loop may exist")
			}

			res, err := launchJobFunc(jobReq)
			if err != nil {
				return models.JobStats{}, err
			}

			if err := rj.statsManager.Update(jobID, "multiple_executions", true); err != nil {
				logger.Error(err)
			}

			if err := rj.statsManager.Update(res.Stats.JobID, "upstream_job_id", jobID); err != nil {
				logger.Error(err)
			}

			rj.statsManager.AttachExecution(jobID, res.Stats.JobID)

			logger.Infof("Launch sub job %s:%s for upstream job %s", res.Stats.JobName, res.Stats.JobID, jobID)
			return res, nil
		}
	}

	jData.ExtraData["launchJobFunc"] = launchJobFuncFactory(j.ID)

	// Use default context
	if rj.context.JobContext == nil {
		rj.context.JobContext = impl.NewDefaultContext(rj.context.SystemContext)
	}

	return rj.context.JobContext.Build(jData)
}

func (rj *RedisJob) shouldDisableRetry(j job.Interface, wj *work.Job, cancelled bool) bool {
	maxFails := j.MaxFails()
	if maxFails == 0 {
		maxFails = 4 // Consistent with backend worker pool
	}
	fails := wj.Fails
	fails++ // as the fail is not returned to backend pool yet

	if cancelled && fails < int64(maxFails) {
		return true
	}

	if !cancelled && fails < int64(maxFails) && !j.ShouldRetry() {
		return true
	}

	return false
}
