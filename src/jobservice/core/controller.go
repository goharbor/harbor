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
	"fmt"
	"github.com/pkg/errors"

	"github.com/goharbor/harbor/src/jobservice/logger"

	"github.com/goharbor/harbor/src/jobservice/common/query"
	"github.com/goharbor/harbor/src/jobservice/common/utils"
	"github.com/goharbor/harbor/src/jobservice/errs"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/lcm"
	"github.com/goharbor/harbor/src/jobservice/worker"
	"github.com/robfig/cron"
)

// basicController implement the core interface and provides related job handle methods.
// basicController will coordinate the lower components to complete the process as a commander role.
type basicController struct {
	// Refer the backend worker
	backendWorker worker.Interface
	// Refer the job life cycle management controller
	ctl lcm.Controller
}

// NewController is constructor of basicController.
func NewController(backendWorker worker.Interface, ctl lcm.Controller) Interface {
	return &basicController{
		backendWorker: backendWorker,
		ctl:           ctl,
	}
}

// LaunchJob is implementation of same method in core interface.
func (bc *basicController) LaunchJob(req *job.Request) (res *job.Stats, err error) {
	if err := validJobReq(req); err != nil {
		return nil, errs.BadRequestError(err)
	}

	// Validate job name
	jobType, isKnownJob := bc.backendWorker.IsKnownJob(req.Job.Name)
	if !isKnownJob {
		return nil, errs.BadRequestError(errors.Errorf("job with name '%s' is unknown", req.Job.Name))
	}

	// Validate parameters
	if err := bc.backendWorker.ValidateJobParameters(jobType, req.Job.Parameters); err != nil {
		return nil, errs.BadRequestError(err)
	}

	// Enqueue job regarding of the kind
	switch req.Job.Metadata.JobKind {
	case job.KindScheduled:
		res, err = bc.backendWorker.Schedule(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.ScheduleDelay,
			req.Job.Metadata.IsUnique,
			req.Job.StatusHook,
		)
	case job.KindPeriodic:
		res, err = bc.backendWorker.PeriodicallyEnqueue(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.Cron,
			req.Job.Metadata.IsUnique,
			req.Job.StatusHook,
		)
	default:
		res, err = bc.backendWorker.Enqueue(
			req.Job.Name,
			req.Job.Parameters,
			req.Job.Metadata.IsUnique,
			req.Job.StatusHook,
		)
	}

	return
}

// GetJob is implementation of same method in core interface.
func (bc *basicController) GetJob(jobID string) (*job.Stats, error) {
	if utils.IsEmptyStr(jobID) {
		return nil, errs.BadRequestError(errors.New("empty job ID"))
	}

	t, err := bc.ctl.Track(jobID)
	if err != nil {
		return nil, err
	}

	return t.Job(), nil
}

// StopJob is implementation of same method in core interface.
func (bc *basicController) StopJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errs.BadRequestError(errors.New("empty job ID"))
	}

	return bc.backendWorker.StopJob(jobID)
}

// RetryJob is implementation of same method in core interface.
func (bc *basicController) RetryJob(jobID string) error {
	if utils.IsEmptyStr(jobID) {
		return errs.BadRequestError(errors.New("empty job ID"))
	}

	return bc.backendWorker.RetryJob(jobID)
}

// GetJobLogData is used to return the log text data for the specified job if exists
func (bc *basicController) GetJobLogData(jobID string) ([]byte, error) {
	if utils.IsEmptyStr(jobID) {
		return nil, errs.BadRequestError(errors.New("empty job ID"))
	}

	logData, err := logger.Retrieve(jobID)
	if err != nil {
		return nil, err
	}

	return logData, nil
}

// CheckStatus is implementation of same method in core interface.
func (bc *basicController) CheckStatus() (*worker.Stats, error) {
	return bc.backendWorker.Stats()
}

// GetPeriodicExecutions gets the periodic executions for the specified periodic job
func (bc *basicController) GetPeriodicExecutions(periodicJobID string, query *query.Parameter) ([]*job.Stats, int64, error) {
	if utils.IsEmptyStr(periodicJobID) {
		return nil, 0, errs.BadRequestError(errors.New("nil periodic job ID"))
	}

	t, err := bc.ctl.Track(periodicJobID)
	if err != nil {
		return nil, 0, err
	}

	eIDs, total, err := t.Executions(query)
	if err != nil {
		return nil, 0, err
	}

	res := make([]*job.Stats, 0)
	for _, eID := range eIDs {
		et, err := bc.ctl.Track(eID)
		if err != nil {
			return nil, 0, err
		}

		res = append(res, et.Job())
	}

	return res, total, nil
}

// ScheduledJobs returns the scheduled jobs by page
func (bc *basicController) ScheduledJobs(query *query.Parameter) ([]*job.Stats, int64, error) {
	return bc.backendWorker.ScheduledJobs(query)
}

func validJobReq(req *job.Request) error {
	if req == nil || req.Job == nil {
		return errors.New("empty job request is not allowed")
	}

	if utils.IsEmptyStr(req.Job.Name) {
		return errors.New("name of job must be specified")
	}

	if req.Job.Metadata == nil {
		return errors.New("metadata of job is missing")
	}

	if req.Job.Metadata.JobKind != job.KindGeneric &&
		req.Job.Metadata.JobKind != job.KindPeriodic &&
		req.Job.Metadata.JobKind != job.KindScheduled {
		return errors.Errorf(
			"job kind '%s' is not supported, only support '%s','%s','%s'",
			req.Job.Metadata.JobKind,
			job.KindGeneric,
			job.KindScheduled,
			job.KindPeriodic)
	}

	if req.Job.Metadata.JobKind == job.KindScheduled &&
		req.Job.Metadata.ScheduleDelay == 0 {
		return errors.Errorf("'schedule_delay' must be specified for %s job", job.KindScheduled)
	}

	if req.Job.Metadata.JobKind == job.KindPeriodic {
		if utils.IsEmptyStr(req.Job.Metadata.Cron) {
			return fmt.Errorf("'cron_spec' must be specified if the %s job", job.KindPeriodic)
		}

		if _, err := cron.Parse(req.Job.Metadata.Cron); err != nil {
			return fmt.Errorf("'cron_spec' is not correctly set: %s: %s", req.Job.Metadata.Cron, err)
		}
	}

	return nil
}
