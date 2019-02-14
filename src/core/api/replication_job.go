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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	api_models "github.com/goharbor/harbor/src/core/api/models"
	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/replication/core"
)

// RepJobAPI handles request to /api/replicationJobs /api/replicationJobs/:id/log
type RepJobAPI struct {
	BaseController
	jobID int64
}

// Prepare validates that whether user has system admin role
func (ra *RepJobAPI) Prepare() {
	ra.BaseController.Prepare()
	if !ra.SecurityCtx.IsAuthenticated() {
		ra.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	if !(ra.Ctx.Request.Method == http.MethodGet || ra.SecurityCtx.IsSysAdmin()) {
		ra.SendForbiddenError(errors.New(ra.SecurityCtx.GetUsername()))
		return
	}

	if len(ra.GetStringFromPath(":id")) != 0 {
		id, err := ra.GetInt64FromPath(":id")
		if err != nil {
			ra.SendBadRequestError(fmt.Errorf("invalid ID: %s", ra.GetStringFromPath(":id")))
			return
		}
		ra.jobID = id
	}

}

// List filters jobs according to the parameters
func (ra *RepJobAPI) List() {

	policyID, err := ra.GetInt64("policy_id")
	if err != nil || policyID <= 0 {
		ra.SendBadRequestError(fmt.Errorf("invalid policy_id: %s", ra.GetString("policy_id")))
		return
	}

	policy, err := core.GlobalController.GetPolicy(policyID)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", policyID, err)
		ra.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", policyID, err))
		return
	}

	if policy.ID == 0 {
		ra.SendNotFoundError(fmt.Errorf("policy %d not found", policyID))
		return
	}

	resource := rbac.NewProjectNamespace(policy.ProjectIDs[0]).Resource(rbac.ResourceReplicationJob)
	if !ra.SecurityCtx.Can(rbac.ActionList, resource) {
		ra.SendForbiddenError(errors.New(ra.SecurityCtx.GetUsername()))
		return
	}

	query := &models.RepJobQuery{
		PolicyID: policyID,
		// hide the schedule job, the schedule job is used to trigger replication
		// for scheduled policy
		Operations: []string{models.RepOpTransfer, models.RepOpDelete},
	}

	query.Repository = ra.GetString("repository")
	query.Statuses = ra.GetStrings("status")
	query.OpUUID = ra.GetString("op_uuid")

	startTimeStr := ra.GetString("start_time")
	if len(startTimeStr) != 0 {
		i, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			ra.SendBadRequestError(fmt.Errorf("invalid start_time: %s", startTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.StartTime = &t
	}

	endTimeStr := ra.GetString("end_time")
	if len(endTimeStr) != 0 {
		i, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			ra.SendBadRequestError(fmt.Errorf("invalid end_time: %s", endTimeStr))
			return
		}
		t := time.Unix(i, 0)
		query.EndTime = &t
	}

	query.Page, query.Size, err = ra.GetPaginationParams()
	if err != nil {
		ra.SendBadRequestError(err)
		return
	}

	total, err := dao.GetTotalCountOfRepJobs(query)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get total count of repository jobs of policy %d: %v", policyID, err))
		return
	}
	jobs, err := dao.GetRepJobs(query)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get repository jobs, query: %v :%v", query, err))
		return
	}

	ra.SetPaginationHeader(total, query.Page, query.Size)

	ra.Data["json"] = jobs
	ra.ServeJSON()
}

// Delete ...
func (ra *RepJobAPI) Delete() {
	if ra.jobID == 0 {
		ra.SendBadRequestError(errors.New("ID is nil"))
		return
	}

	job, err := dao.GetRepJob(ra.jobID)
	if err != nil {
		log.Errorf("failed to get job %d: %v", ra.jobID, err)
		ra.SendInternalServerError(fmt.Errorf("failed to get job %d: %v", ra.jobID, err))
		return
	}

	if job == nil {
		ra.SendNotFoundError(fmt.Errorf("job %d not found", ra.jobID))
		return
	}

	if job.Status == models.JobPending || job.Status == models.JobRunning {
		ra.SendBadRequestError(fmt.Errorf("job is %s, can not be deleted", job.Status))
		return
	}

	if err = dao.DeleteRepJob(ra.jobID); err != nil {
		log.Errorf("failed to deleted job %d: %v", ra.jobID, err)
		ra.SendInternalServerError(fmt.Errorf("failed to deleted job %d: %v", ra.jobID, err))
		return
	}
}

// GetLog ...
func (ra *RepJobAPI) GetLog() {
	if ra.jobID == 0 {
		ra.SendBadRequestError(errors.New("ID is nil"))
		return
	}

	job, err := dao.GetRepJob(ra.jobID)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get replication job %d: %v", ra.jobID, err))
		return
	}

	if job == nil {
		ra.SendNotFoundError(fmt.Errorf("replication job %d not found", ra.jobID))
		return
	}

	policy, err := core.GlobalController.GetPolicy(job.PolicyID)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", job.PolicyID, err))
		return
	}

	resource := rbac.NewProjectNamespace(policy.ProjectIDs[0]).Resource(rbac.ResourceReplicationJob)
	if !ra.SecurityCtx.Can(rbac.ActionRead, resource) {
		ra.SendForbiddenError(errors.New(ra.SecurityCtx.GetUsername()))
		return
	}

	logBytes, err := utils.GetJobServiceClient().GetJobLog(job.UUID)
	if err != nil {
		ra.ParseAndHandleError(fmt.Sprintf("failed to get log of job %s",
			job.UUID), err)
		return
	}
	ra.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(logBytes)))
	ra.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")
	_, err = ra.Ctx.ResponseWriter.Write(logBytes)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to write log of job %s: %v", job.UUID, err))
		return
	}
}

// StopJobs stop replication jobs for the policy
func (ra *RepJobAPI) StopJobs() {
	req := &api_models.StopJobsReq{}
	isValid, err := ra.DecodeJSONReqAndValidate(req)
	if !isValid {
		ra.SendBadRequestError(err)
		return
	}

	policy, err := core.GlobalController.GetPolicy(req.PolicyID)
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to get policy %d: %v", req.PolicyID, err))
		return
	}

	if policy.ID == 0 {
		ra.SendNotFoundError(fmt.Errorf("policy %d not found", req.PolicyID))
		return
	}

	jobs, err := dao.GetRepJobs(&models.RepJobQuery{
		PolicyID:   policy.ID,
		Operations: []string{models.RepOpTransfer, models.RepOpDelete},
	})
	if err != nil {
		ra.SendInternalServerError(fmt.Errorf("failed to list jobs of policy %d: %v", policy.ID, err))
		return
	}
	for _, job := range jobs {
		if err = utils.GetJobServiceClient().PostAction(job.UUID, common_job.JobActionStop); err != nil {
			log.Errorf("failed to stop job id-%d uuid-%s: %v", job.ID, job.UUID, err)
			continue
		}
	}
}

// TODO:add Post handler to call job service API to submit jobs by policy
