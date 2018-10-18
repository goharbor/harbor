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

package core

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/models"
	"github.com/goharbor/harbor/src/jobservice/pool"
	"github.com/goharbor/harbor/src/jobservice/utils"
	"github.com/robfig/cron"
)

const (
	hookActivated   = "activated"
	hookDeactivated = "error"
)

// Controller implement the core interface and provides related job handle methods.
// Controller will coordinate the lower components to complete the process as a commander role.
type Controller struct {
	// Refer the backend pool
	backendPool pool.Interface
}

// NewController is constructor of Controller.
func NewController(backendPool pool.Interface) *Controller {
	return &Controller{
		backendPool: backendPool,
	}
}

// LaunchJob is implementation of same method in core interface.
func (c *Controller) LaunchJob(req models.JobRequest) (models.JobStats, error) {
	if err := validJobReq(req); err != nil {
		return models.JobStats{}, err
	}

	// Validate job name
	jobType, isKnownJob := c.backendPool.IsKnownJob(req.Job.Name)
	if !isKnownJob {
		return models.JobStats{}, fmt.Errorf("job with name '%s' is unknown", req.Job.Name)
	}

	// Validate parameters
	if err := c.backendPool.ValidateJobParameters(jobType, req.Job.Parameters); err != nil {
		return models.JobStats{}, err
	}

	// Enqueue job regarding of the kind
	var (
		res models.JobStats
		err error
	)
	switch req.Job.Metadata.JobKind {
	case job.JobKindScheduled:
		res, err = c.backendPool.Schedule(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.ScheduleDelay,
			req.Job.Metadata.IsUnique)
	case job.JobKindPeriodic:
		res, err = c.backendPool.PeriodicallyEnqueue(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.Cron)
	default:
		res, err = c.backendPool.Enqueue(req.Job.Name, req.Job.Parameters, req.Job.Metadata.IsUnique)
	}

	// Register status hook?
	if err == nil {
		if !utils.IsEmptyStr(req.Job.StatusHook) {
			if err := c.backendPool.RegisterHook(res.Stats.JobID, req.Job.StatusHook); err != nil {
				res.Stats.HookStatus = hookDeactivated
			} else {
				res.Stats.HookStatus = hookActivated
			}
		}
	}

	return res, err
}

// GetJob is implementation of same method in core interface.
func (c *Controller) GetJob(jobID string) (models.JobStats, error) {
	if utils.IsEmptyStr(jobID) {
		return models.JobStats{}, errors.New("empty job ID")
	}

	return c.backendPool.GetJobStats(jobID)
}

// StopJob is implementation of same method in core interface.
func (c *Controller) StopJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	return c.backendPool.StopJob(jobID)
}

// CancelJob is implementation of same method in core interface.
func (c *Controller) CancelJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	return c.backendPool.CancelJob(jobID)
}

// RetryJob is implementation of same method in core interface.
func (c *Controller) RetryJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errors.New("empty job ID")
	}

	return c.backendPool.RetryJob(jobID)
}

// GetJobLogData is used to return the log text data for the specified job if exists
func (c *Controller) GetJobLogData(jobID string) ([]byte, error) {
	if utils.IsEmptyStr(jobID) {
		return nil, errors.New("empty job ID")
	}

	logPath := fmt.Sprintf("%s/%s.log", config.GetLogBasePath(), jobID)
	if !utils.FileExists(logPath) {
		return nil, errs.NoObjectFoundError(fmt.Sprintf("%s.log", jobID))
	}

	logData, err := ioutil.ReadFile(logPath)
	if err != nil {
		return nil, err
	}

	return logData, nil
}

// CheckStatus is implementation of same method in core interface.
func (c *Controller) CheckStatus() (models.JobPoolStats, error) {
	return c.backendPool.Stats()
}

func validJobReq(req models.JobRequest) error {
	if req.Job == nil {
		return errors.New("empty job request is not allowed")
	}

	if utils.IsEmptyStr(req.Job.Name) {
		return errors.New("name of job must be specified")
	}

	if req.Job.Metadata == nil {
		return errors.New("metadata of job is missing")
	}

	if req.Job.Metadata.JobKind != job.JobKindGeneric &&
		req.Job.Metadata.JobKind != job.JobKindPeriodic &&
		req.Job.Metadata.JobKind != job.JobKindScheduled {
		return fmt.Errorf(
			"job kind '%s' is not supported, only support '%s','%s','%s'",
			req.Job.Metadata.JobKind,
			job.JobKindGeneric,
			job.JobKindScheduled,
			job.JobKindPeriodic)
	}

	if req.Job.Metadata.JobKind == job.JobKindScheduled &&
		req.Job.Metadata.ScheduleDelay == 0 {
		return fmt.Errorf("'schedule_delay' must be specified if the job kind is '%s'", job.JobKindScheduled)
	}

	if req.Job.Metadata.JobKind == job.JobKindPeriodic {
		if utils.IsEmptyStr(req.Job.Metadata.Cron) {
			return fmt.Errorf("'cron_spec' must be specified if the job kind is '%s'", job.JobKindPeriodic)
		}

		if _, err := cron.Parse(req.Job.Metadata.Cron); err != nil {
			return fmt.Errorf("'cron_spec' is not correctly set: %s", err)
		}
	}

	return nil
}
