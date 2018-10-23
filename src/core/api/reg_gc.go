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
	"net/http"
	"os"
	"strconv"

	"encoding/json"
	"github.com/goharbor/harbor/src/common/dao"
	common_http "github.com/goharbor/harbor/src/common/http"
	common_job "github.com/goharbor/harbor/src/common/job"
	common_models "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api/models"
	utils_core "github.com/goharbor/harbor/src/core/utils"
)

// GCAPI handles request of harbor admin...
type GCAPI struct {
	BaseController
}

// Prepare validates the URL and parms, it needs the system admin permission.
func (gc *GCAPI) Prepare() {
	gc.BaseController.Prepare()
	if !gc.SecurityCtx.IsAuthenticated() {
		gc.HandleUnauthorized()
		return
	}
	if !gc.SecurityCtx.IsSysAdmin() {
		gc.HandleForbidden(gc.SecurityCtx.GetUsername())
		return
	}
}

// Post ...
func (gc *GCAPI) Post() {
	gr := models.GCReq{}
	gc.DecodeJSONReqAndValidate(&gr)
	gc.submitJob(&gr)
	gc.Redirect(http.StatusCreated, strconv.FormatInt(gr.ID, 10))
}

// Put ...
func (gc *GCAPI) Put() {
	gr := models.GCReq{}
	gc.DecodeJSONReqAndValidate(&gr)

	if gr.Schedule.Type == models.ScheduleManual {
		gc.HandleInternalServerError(fmt.Sprintf("Fail to update GC schedule as wrong schedule type: %s.", gr.Schedule.Type))
		return
	}

	query := &common_models.AdminJobQuery{
		Name: common_job.ImageGC,
		Kind: common_job.JobKindPeriodic,
	}
	jobs, err := dao.GetAdminJobs(query)
	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
	if len(jobs) != 1 {
		gc.HandleInternalServerError("Fail to update GC schedule, only one schedule is accepted.")
		return
	}

	// stop the scheduled job and remove it.
	if err = utils_core.GetJobServiceClient().PostAction(jobs[0].UUID, common_job.JobActionStop); err != nil {
		if e, ok := err.(*common_http.Error); !ok || e.Code != http.StatusNotFound {
			gc.HandleInternalServerError(fmt.Sprintf("%v", err))
			return
		}
	}

	if err = dao.DeleteAdminJob(jobs[0].ID); err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}

	// Set schedule to None means to cancel the schedule, won't add new job.
	if gr.Schedule.Type != models.ScheduleNone {
		gc.submitJob(&gr)
	}
}

// GetGC ...
func (gc *GCAPI) GetGC() {
	id, err := gc.GetInt64FromPath(":id")
	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("need to specify gc id"))
		return
	}

	jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
		ID: id,
	})

	gcreps := []*models.GCRep{}
	for _, job := range jobs {
		gcrep, err := convertToGCRep(job)
		if err != nil {
			gc.HandleInternalServerError(fmt.Sprintf("failed to convert gc response: %v", err))
			return
		}
		gcreps = append(gcreps, &gcrep)
	}

	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("failed to get admin jobs: %v", err))
		return
	}
	gc.Data["json"] = gcreps
	gc.ServeJSON()
}

// List ...
func (gc *GCAPI) List() {
	jobs, err := dao.GetTop10AdminJobsOfName(common_job.ImageGC)
	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("failed to get admin jobs: %v", err))
		return
	}
	gcreps := []*models.GCRep{}
	for _, job := range jobs {
		gcrep, err := convertToGCRep(job)
		if err != nil {
			gc.HandleInternalServerError(fmt.Sprintf("failed to convert gc response: %v", err))
			return
		}
		gcreps = append(gcreps, &gcrep)
	}
	gc.Data["json"] = gcreps
	gc.ServeJSON()
}

// Get gets GC schedule ...
func (gc *GCAPI) Get() {
	jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
		Name: common_job.ImageGC,
		Kind: common_job.JobKindPeriodic,
	})
	if err != nil {
		gc.HandleNotFound(fmt.Sprintf("failed to get admin jobs: %v", err))
		return
	}
	if len(jobs) > 1 {
		gc.HandleInternalServerError("Get more than one GC scheduled job, make sure there has only one.")
		return
	}
	gcreps := []*models.GCRep{}
	for _, job := range jobs {
		gcrep, err := convertToGCRep(job)
		if err != nil {
			gc.HandleInternalServerError(fmt.Sprintf("failed to convert gc response: %v", err))
			return
		}
		gcreps = append(gcreps, &gcrep)
	}
	gc.Data["json"] = gcreps
	gc.ServeJSON()
}

// GetLog ...
func (gc *GCAPI) GetLog() {
	id, err := gc.GetInt64FromPath(":id")
	if err != nil {
		gc.HandleBadRequest("invalid ID")
		return
	}
	job, err := dao.GetAdminJob(id)
	if err != nil {
		log.Errorf("Failed to load job data for job: %d, error: %v", id, err)
		gc.CustomAbort(http.StatusInternalServerError, "Failed to get Job data")
	}
	if job == nil {
		log.Errorf("Failed to get admin job: %d", id)
		gc.CustomAbort(http.StatusNotFound, "Failed to get Job")
	}

	logBytes, err := utils_core.GetJobServiceClient().GetJobLog(job.UUID)
	if err != nil {
		if httpErr, ok := err.(*common_http.Error); ok {
			gc.RenderError(httpErr.Code, "")
			log.Errorf(fmt.Sprintf("failed to get log of job %d: %d %s",
				id, httpErr.Code, httpErr.Message))
			return
		}
		gc.HandleInternalServerError(fmt.Sprintf("Failed to get job logs, uuid: %s, error: %v", job.UUID, err))
		return
	}
	gc.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	gc.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = gc.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("Failed to write job logs, uuid: %s, error: %v", job.UUID, err))
	}
}

// submitJob submits a job to job service per request
func (gc *GCAPI) submitJob(gr *models.GCReq) {
	// cannot post multiple schdule for GC job.
	if gr.IsPeriodic() {
		jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
			Name: common_job.ImageGC,
			Kind: common_job.JobKindPeriodic,
		})
		if err != nil {
			gc.HandleInternalServerError(fmt.Sprintf("failed to get admin jobs: %v", err))
			return
		}
		if len(jobs) != 0 {
			gc.HandleStatusPreconditionFailed("Fail to set schedule for GC as always had one, please delete it firstly then to re-schedule.")
			return
		}
	}

	id, err := dao.AddAdminJob(&common_models.AdminJob{
		Name: common_job.ImageGC,
		Kind: gr.JobKind(),
		Cron: gr.CronString(),
	})
	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
	gr.ID = id
	gr.Parameters = map[string]interface{}{
		"redis_url_reg": os.Getenv("_REDIS_URL_REG"),
	}
	job, err := gr.ToJob()
	if err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}

	// submit job to jobservice
	log.Debugf("submiting GC admin job to jobservice")
	uuid, err := utils_core.GetJobServiceClient().SubmitJob(job)
	if err != nil {
		if err := dao.DeleteAdminJob(id); err != nil {
			log.Debugf("Failed to delete admin job, err: %v", err)
		}
		gc.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
	if err := dao.SetAdminJobUUID(id, uuid); err != nil {
		gc.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
}

func convertToGCRep(job *common_models.AdminJob) (models.GCRep, error) {
	if job == nil {
		return models.GCRep{}, nil
	}

	gcrep := models.GCRep{
		ID:           job.ID,
		Name:         job.Name,
		Kind:         job.Kind,
		Status:       job.Status,
		Deleted:      job.Deleted,
		CreationTime: job.CreationTime,
		UpdateTime:   job.UpdateTime,
	}
	if len(job.Cron) > 0 {
		schedule := &models.ScheduleParam{}
		if err := json.Unmarshal([]byte(job.Cron), &schedule); err != nil {
			return models.GCRep{}, err
		}
		gcrep.Schedule = schedule
	}
	return gcrep, nil
}
