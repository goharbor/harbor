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
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
    "github.com/vmware/harbor/src/common/api"
)

// RepJobAPI handles request to /api/replicationJobs /api/replicationJobs/:id/log
type RepJobAPI struct {
	api.BaseAPI
	jobID int64
}

// Prepare validates that whether user has system admin role
func (ra *RepJobAPI) Prepare() {
	uid := ra.ValidateUser()
	isAdmin, err := dao.IsAdminRole(uid)
	if err != nil {
		log.Errorf("Failed to Check if the user is admin, error: %v, uid: %d", err, uid)
	}
	if !isAdmin {
		ra.CustomAbort(http.StatusForbidden, "")
	}

	idStr := ra.Ctx.Input.Param(":id")
	if len(idStr) != 0 {
		id, err := strconv.ParseInt(idStr, 10, 64)
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

	req, err := http.NewRequest("GET", buildJobLogURL(strconv.FormatInt(ra.jobID, 10)), nil)
	if err != nil {
		log.Errorf("failed to create a request: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, "")
	}
	addAuthentication(req)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("failed to get log for job %d: %v", ra.jobID, err)
		ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		ra.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), resp.Header.Get(http.CanonicalHeaderKey("Content-Length")))
		ra.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")

		if _, err = io.Copy(ra.Ctx.ResponseWriter, resp.Body); err != nil {
			log.Errorf("failed to write log to response; %v", err)
			ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read reponse body: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	ra.CustomAbort(resp.StatusCode, string(b))
}

//TODO:add Post handler to call job service API to submit jobs by policy
