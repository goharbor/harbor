// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/api/scan"
	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	common_job "github.com/goharbor/harbor/src/common/job"
	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api/models"
	utils_core "github.com/goharbor/harbor/src/core/utils"
	"github.com/pkg/errors"
)

// AJAPI manages the CRUD of admin job and its schedule, any API wants to handle manual and cron job like ScanAll and GC cloud reuse it.
type AJAPI struct {
	BaseController
}

// Prepare validates the URL and parms, it needs the system admin permission.
func (aj *AJAPI) Prepare() {
	aj.BaseController.Prepare()
}

// updateSchedule update a schedule of admin job.
func (aj *AJAPI) updateSchedule(ajr models.AdminJobReq) {
	if ajr.Schedule.Type == models.ScheduleManual {
		aj.SendInternalServerError((fmt.Errorf("fail to update admin job schedule as wrong schedule type: %s", ajr.Schedule.Type)))
		return
	}

	query := &common_models.AdminJobQuery{
		Name: ajr.Name,
		Kind: common_job.JobKindPeriodic,
	}
	jobs, err := dao.GetAdminJobs(query)
	if err != nil {
		aj.SendInternalServerError(err)
		return
	}
	if len(jobs) != 1 {
		aj.SendInternalServerError(errors.New("fail to update admin job schedule as we found more than one schedule in system, please ensure that only one schedule left for your job"))
		return
	}

	// stop the scheduled job and remove it.
	if err = utils_core.GetJobServiceClient().PostAction(jobs[0].UUID, common_job.JobActionStop); err != nil {
		_, ok := err.(*common_job.StatusBehindError)
		if !ok {
			if e, ok := err.(*common_http.Error); !ok || e.Code != http.StatusNotFound {
				aj.SendInternalServerError(err)
				return
			}
		}
	}

	if err = dao.DeleteAdminJob(jobs[0].ID); err != nil {
		aj.SendInternalServerError(err)
		return
	}

	// Set schedule to None means to cancel the schedule, won't add new job.
	if ajr.Schedule.Type != models.ScheduleNone {
		aj.submit(&ajr)
	}
}

// get get a execution of admin job by ID
func (aj *AJAPI) get(id int64) {
	jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
		ID: id,
	})
	if err != nil {
		aj.SendInternalServerError(fmt.Errorf("failed to get admin jobs: %v", err))
		return
	}
	if len(jobs) == 0 {
		aj.SendNotFoundError(errors.New("no admin job found"))
		return
	}

	adminJobRep, err := convertToAdminJobRep(jobs[0])
	if err != nil {
		aj.SendInternalServerError(fmt.Errorf("failed to convert admin job response: %v", err))
		return
	}

	aj.Data["json"] = adminJobRep
	aj.ServeJSON()
}

// list list all executions of admin job by name
func (aj *AJAPI) list(name string) {
	jobs, err := dao.GetTop10AdminJobsOfName(name)
	if err != nil {
		aj.SendInternalServerError(fmt.Errorf("failed to get admin jobs: %v", err))
		return
	}

	AdminJobReps := []*models.AdminJobRep{}
	for _, job := range jobs {
		AdminJobRep, err := convertToAdminJobRep(job)
		if err != nil {
			aj.SendInternalServerError(fmt.Errorf("failed to convert admin job response: %v", err))
			return
		}
		AdminJobReps = append(AdminJobReps, &AdminJobRep)
	}

	aj.Data["json"] = AdminJobReps
	aj.ServeJSON()
}

// getSchedule gets admin job schedule ...
func (aj *AJAPI) getSchedule(name string) {
	adminJobSchedule := models.AdminJobSchedule{}

	jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
		Name: name,
		Kind: common_job.JobKindPeriodic,
	})
	if err != nil {
		aj.SendInternalServerError(fmt.Errorf("failed to get admin jobs: %v", err))
		return
	}
	if len(jobs) > 1 {
		aj.SendInternalServerError(errors.New("get more than one scheduled admin job, make sure there has only one"))
		return
	}

	if len(jobs) != 0 {
		adminJobRep, err := convertToAdminJobRep(jobs[0])
		if err != nil {
			aj.SendInternalServerError(fmt.Errorf("failed to convert admin job response: %v", err))
			return
		}
		adminJobSchedule.Schedule = adminJobRep.Schedule
	}

	aj.Data["json"] = adminJobSchedule
	aj.ServeJSON()
}

// getLog ...
func (aj *AJAPI) getLog(id int64) {
	job, err := dao.GetAdminJob(id)
	if err != nil {
		log.Errorf("Failed to load job data for job: %d, error: %v", id, err)
		aj.SendInternalServerError(errors.New("Failed to get Job data"))
		return
	}
	if job == nil {
		log.Errorf("Failed to get admin job: %d", id)
		aj.SendNotFoundError(errors.New("Failed to get Job"))
		return
	}

	var jobID string
	// to get the latest execution job id, then to query job log.
	if job.Kind == common_job.JobKindPeriodic {
		exes, err := utils_core.GetJobServiceClient().GetExecutions(job.UUID)
		if err != nil {
			aj.SendInternalServerError(err)
			return
		}
		if len(exes) == 0 {
			aj.SendNotFoundError(errors.New("no execution log found"))
			return
		}
		// get the latest terminal status execution.
		for _, exe := range exes {
			if exe.Info.Status == "Error" || exe.Info.Status == "Success" {
				jobID = exe.Info.JobID
				break
			}
		}
		// no execution found
		if jobID == "" {
			aj.SendNotFoundError(errors.New("no execution log found"))
			return
		}

	} else {
		jobID = job.UUID
	}

	logBytes, err := utils_core.GetJobServiceClient().GetJobLog(jobID)
	if err != nil {
		if httpErr, ok := err.(*common_http.Error); ok {
			aj.RenderError(httpErr.Code, "")
			log.Errorf(fmt.Sprintf("failed to get log of job %d: %d %s",
				id, httpErr.Code, httpErr.Message))
			return
		}
		aj.SendInternalServerError(fmt.Errorf("Failed to get job logs, uuid: %s, error: %v", job.UUID, err))
		return
	}
	aj.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	aj.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = aj.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		aj.SendInternalServerError(fmt.Errorf("Failed to write job logs, uuid: %s, error: %v", job.UUID, err))
	}
}

// submit submits a job to job service per request
func (aj *AJAPI) submit(ajr *models.AdminJobReq) {
	// when the schedule is saved as None without any schedule, just return 200 and do nothing.
	if ajr.Schedule == nil || ajr.Schedule.Type == models.ScheduleNone {
		return
	}

	// cannot post multiple schedule for admin job.
	if ajr.IsPeriodic() {
		jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
			Name: ajr.Name,
			Kind: common_job.JobKindPeriodic,
		})
		if err != nil {
			aj.SendInternalServerError(fmt.Errorf("failed to get admin jobs: %v", err))
			return
		}
		if len(jobs) != 0 {
			aj.SendPreconditionFailedError(errors.New("fail to set schedule for admin job as always had one, please delete it firstly then to re-schedule"))
			return
		}
	} else {
		// So far, it should be a generic job for the manually trigger case.
		// Only needs to care the 1st generic job.
		// Check if there are still ongoing scan jobs triggered by the previous admin job.
		// TODO: REPLACE WITH TASK MANAGER METHODS IN FUTURE
		jb, err := aj.getLatestAdminJob(ajr.Name, common_job.JobKindGeneric)
		if err != nil {
			aj.SendInternalServerError(errors.Wrap(err, "AJAPI"))
			return
		}

		if jb != nil {
			// With a reasonable timeout duration
			if jb.UpdateTime.Add(2 * time.Hour).After(time.Now()) {
				if isOnGoing(jb.Status) {
					err := errors.Errorf("reject job submitting: job %s with ID %d is %s", jb.Name, jb.ID, jb.Status)
					aj.SendInternalServerError(errors.Wrap(err, "submit : AJAPI"))
					return
				}

				// For scan all job, check more
				if jb.Name == common_job.ImageScanAllJob {
					// Get the overall stats with the ID of the previous job
					stats, err := scan.DefaultController.GetStats(fmt.Sprintf("%d", jb.ID))
					if err != nil {
						aj.SendInternalServerError(errors.Wrap(err, "submit : AJAPI"))
						return
					}

					if stats.Total != stats.Completed {
						// Not all scan processes are completed
						err := errors.Errorf("scan processes started by %s job with ID %d is in progress: %s", jb.Name, jb.ID, progress(stats.Completed, stats.Total))
						aj.SendPreconditionFailedError(errors.Wrap(err, "submit : AJAPI"))
						return
					}
				}
			}
		}
	}

	id, err := dao.AddAdminJob(&common_models.AdminJob{
		Name: ajr.Name,
		Kind: ajr.JobKind(),
		Cron: ajr.CronString(),
	})
	if err != nil {
		aj.SendInternalServerError(err)
		return
	}
	ajr.ID = id
	job := ajr.ToJob()

	// submit job to job service
	log.Debugf("submitting admin job to job service")
	uuid, err := utils_core.GetJobServiceClient().SubmitJob(job)
	if err != nil {
		if err := dao.DeleteAdminJob(id); err != nil {
			log.Debugf("Failed to delete admin job, err: %v", err)
		}
		aj.ParseAndHandleError("failed to submit admin job", err)
		return
	}
	if err := dao.SetAdminJobUUID(id, uuid); err != nil {
		aj.SendInternalServerError(err)
		return
	}
}

func (aj *AJAPI) getLatestAdminJob(name, kind string) (*common_models.AdminJob, error) {
	query := &common_models.AdminJobQuery{
		Name: name,
		Kind: kind,
	}
	query.Size = 1
	query.Page = 1

	jbs, err := dao.GetAdminJobs(query)

	if err != nil {
		return nil, err
	}

	if len(jbs) == 0 {
		// Not exist
		return nil, nil
	}

	// Return the latest one (with biggest ID)
	return jbs[0], nil
}

func convertToAdminJobRep(job *common_models.AdminJob) (models.AdminJobRep, error) {
	if job == nil {
		return models.AdminJobRep{}, nil
	}

	AdminJobRep := models.AdminJobRep{
		ID:           job.ID,
		Name:         job.Name,
		Kind:         job.Kind,
		Status:       job.Status,
		CreationTime: job.CreationTime,
		UpdateTime:   job.UpdateTime,
	}

	if len(job.Cron) > 0 {
		schedule, err := models.ConvertSchedule(job.Cron)
		if err != nil {
			return models.AdminJobRep{}, err
		}
		AdminJobRep.Schedule = &schedule
	}
	return AdminJobRep, nil
}

func progress(completed, total uint) string {
	if total == 0 {
		return fmt.Sprintf("0%s", "%")
	}

	v := float64(completed)
	vv := float64(total)

	p := (int)(math.Round((v / vv) * 100))

	return fmt.Sprintf("%d%s", p, "%")
}

func isOnGoing(status string) bool {
	return status == common_models.JobRunning ||
		status == common_models.JobScheduled ||
		status == common_models.JobPending
}
