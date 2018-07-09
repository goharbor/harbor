// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"strconv"

	"github.com/vmware/harbor/src/common/dao"
	common_http "github.com/vmware/harbor/src/common/http"
	"github.com/vmware/harbor/src/common/job"
	common_models "github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/api/models"
	utils_ui "github.com/vmware/harbor/src/ui/utils"
)

// AdminJobAPI handles request of harbor admin...
type AdminJobAPI struct {
	BaseController
}

// Prepare validates the URL and parms
func (aj *AdminJobAPI) Prepare() {
	aj.BaseController.Prepare()
	if !aj.SecurityCtx.IsAuthenticated() {
		aj.HandleUnauthorized()
		return
	}
	if !aj.SecurityCtx.IsSysAdmin() {
		aj.HandleForbidden(aj.SecurityCtx.GetUsername())
		return
	}

}

//Post ...
func (aj *AdminJobAPI) Post() {

	ajr := models.AdminJobReq{}
	aj.DecodeJSONReqAndValidate(&ajr)

	aj.submitJob(ajr)
}

//Put ...
func (aj *AdminJobAPI) Put() {
	ajr := models.AdminJobReq{}
	aj.DecodeJSONReq(&ajr)

	query := &common_models.AdminJobQuery{
		Name: ajr.Name,
	}
	jobs, err := dao.GetAdminJobs(query)
	if err != nil {
		aj.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
	if len(jobs) != 1 {
		log.Warningf("find more than one admin jobs.")
	}

	// stop the scheduled job and remove it.
	for _, j := range jobs {
		if err = utils_ui.GetJobServiceClient().PostAction(j.UUID, job.JobActionStop); err != nil {
			if e, ok := err.(*common_http.Error); !ok || e.Code != http.StatusNotFound {
				aj.HandleInternalServerError(fmt.Sprintf("%v", err))
				return
			}
		}
		if err = dao.DeleteAdminJob(j.ID); err != nil {
			aj.HandleInternalServerError(fmt.Sprintf("%v", err))
			return
		}
	}

	aj.submitJob(ajr)
}

// Get ...
func (aj *AdminJobAPI) Get() {
	jobName := aj.GetStringFromPath(":name")
	if len(jobName) <= 0 {
		aj.HandleInternalServerError(fmt.Sprintf("need to specify job name"))
		return
	}

	jobs, err := dao.GetAdminJobs(&common_models.AdminJobQuery{
		Name: jobName,
	})
	if err != nil {
		aj.HandleInternalServerError(fmt.Sprintf("failed to get admin jobs: %v", err))
		return
	}
	aj.Data["json"] = jobs
	aj.ServeJSON()
}

//GetLog ...
func (aj *AdminJobAPI) GetLog() {
	id, err := aj.GetInt64FromPath(":id")
	if err != nil {
		aj.HandleBadRequest("invalid ID")
		return
	}
	job, err := dao.GetAdminJob(id)
	if err != nil {
		log.Errorf("Failed to load job data for job: %d, error: %v", id, err)
		aj.CustomAbort(http.StatusInternalServerError, "Failed to get Job data")
	}
	if job == nil {
		log.Errorf("Failed to get admin job: %d", id)
		aj.CustomAbort(http.StatusInternalServerError, "Failed to get Job")
	}

	logBytes, err := utils_ui.GetJobServiceClient().GetJobLog(job.UUID)
	if err != nil {
		if httpErr, ok := err.(*common_http.Error); ok {
			aj.RenderError(httpErr.Code, "")
			log.Errorf(fmt.Sprintf("failed to get log of job %d: %d %s",
				id, httpErr.Code, httpErr.Message))
			return
		}
		aj.HandleInternalServerError(fmt.Sprintf("Failed to get job logs, uuid: %s, error: %v", job.UUID, err))
		return
	}
	aj.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	aj.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = aj.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		aj.HandleInternalServerError(fmt.Sprintf("Failed to write job logs, uuid: %s, error: %v", job.UUID, err))
	}
}

// submitJob submits a job to job service per request
func (aj *AdminJobAPI) submitJob(ajr models.AdminJobReq) {
	job, err := ajr.ToJob()
	if err != nil {
		aj.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}

	// submit job to jobservice
	log.Debugf("submiting admin job to jobservice, type: %s, parameters: %s,", ajr.Name)
	uuid, err := utils_ui.GetJobServiceClient().SubmitJob(job)
	if err != nil {
		aj.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}

	id, err := dao.AddAdminJob(&common_models.AdminJob{
		Name: ajr.Name,
		Kind: ajr.Kind,
	})
	if err != nil {
		aj.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
	ajr.ID = id

	// create the mapping relationship between the jobs in database and jobservice
	if err = dao.SetAdminJobUUID(id, uuid); err != nil {
		aj.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
}
