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
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/utils"
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
		ra.HandleUnauthorized()
		return
	}

	if !ra.SecurityCtx.IsSysAdmin() {
		ra.HandleForbidden(ra.SecurityCtx.GetUsername())
		return
	}

	if len(ra.GetStringFromPath(":id")) != 0 {
		id, err := ra.GetInt64FromPath(":id")
		if err != nil {
			ra.CustomAbort(http.StatusBadRequest, "ID is invalid")
		}
		ra.jobID = id
	}

}

// List filters jobs according to the parameters
func (ra *RepJobAPI) List() {

	policyID, err := ra.GetInt64("policy_id")
	if err != nil || policyID <= 0 {
		ra.CustomAbort(http.StatusBadRequest, "invalid policy_id")
	}

	policy, err := dao.GetRepPolicy(policyID)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", policyID, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	if policy == nil {
		ra.CustomAbort(http.StatusNotFound, fmt.Sprintf("policy %d not found", policyID))
	}

	repository := ra.GetString("repository")
	status := ra.GetString("status")

	var startTime *time.Time
	startTimeStr := ra.GetString("start_time")
	if len(startTimeStr) != 0 {
		i, err := strconv.ParseInt(startTimeStr, 10, 64)
		if err != nil {
			ra.CustomAbort(http.StatusBadRequest, "invalid start_time")
		}
		t := time.Unix(i, 0)
		startTime = &t
	}

	var endTime *time.Time
	endTimeStr := ra.GetString("end_time")
	if len(endTimeStr) != 0 {
		i, err := strconv.ParseInt(endTimeStr, 10, 64)
		if err != nil {
			ra.CustomAbort(http.StatusBadRequest, "invalid end_time")
		}
		t := time.Unix(i, 0)
		endTime = &t
	}

	page, pageSize := ra.GetPaginationParams()

	jobs, total, err := dao.FilterRepJobs(policyID, repository, status,
		startTime, endTime, pageSize, pageSize*(page-1))
	if err != nil {
		log.Errorf("failed to filter jobs according policy ID %d, repository %s, status %s, start time %v, end time %v: %v",
			policyID, repository, status, startTime, endTime, err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}

	ra.SetPaginationHeader(total, page, pageSize)

	ra.Data["json"] = jobs
	ra.ServeJSON()
}

// Delete ...
func (ra *RepJobAPI) Delete() {
	if ra.jobID == 0 {
		ra.CustomAbort(http.StatusBadRequest, "id is nil")
	}

	job, err := dao.GetRepJob(ra.jobID)
	if err != nil {
		log.Errorf("failed to get job %d: %v", ra.jobID, err)
		ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if job == nil {
		ra.CustomAbort(http.StatusNotFound, fmt.Sprintf("job %d not found", ra.jobID))
	}

	if job.Status == models.JobPending || job.Status == models.JobRunning {
		ra.CustomAbort(http.StatusBadRequest, fmt.Sprintf("job is %s, can not be deleted", job.Status))
	}

	if err = dao.DeleteRepJob(ra.jobID); err != nil {
		log.Errorf("failed to deleted job %d: %v", ra.jobID, err)
		ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// GetLog ...
func (ra *RepJobAPI) GetLog() {
	if ra.jobID == 0 {
		ra.CustomAbort(http.StatusBadRequest, "id is nil")
	}
	url := buildJobLogURL(strconv.FormatInt(ra.jobID, 10), ReplicationJobType)
	err := utils.RequestAsUI(http.MethodGet, url, nil, utils.NewJobLogRespHandler(&ra.BaseAPI))
	if err != nil {
		ra.RenderError(http.StatusInternalServerError, err.Error())
		return
	}
}

//TODO:add Post handler to call job service API to submit jobs by policy
