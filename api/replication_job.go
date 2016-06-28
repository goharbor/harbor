/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package api

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// RepJobAPI handles request to /api/replicationJobs /api/replicationJobs/:id/log
type RepJobAPI struct {
	BaseAPI
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

// List filters jobs according to the policy and repository
func (ra *RepJobAPI) List() {
	var policyID int64
	var repository, status string
	var err error

	policyIDStr := ra.GetString("policy_id")
	if len(policyIDStr) != 0 {
		policyID, err = strconv.ParseInt(policyIDStr, 10, 64)
		if err != nil || policyID <= 0 {
			ra.CustomAbort(http.StatusBadRequest, fmt.Sprintf("invalid policy ID: %s", policyIDStr))
		}
	}

	repository = ra.GetString("repository")
	status = ra.GetString("status")

	jobs, err := dao.FilterRepJobs(policyID, repository, status)
	if err != nil {
		log.Errorf("failed to filter jobs according policy ID %d, repository %s, status %s: %v", policyID, repository, status, err)
		ra.RenderError(http.StatusInternalServerError, "Failed to query job")
		return
	}
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

	resp, err := http.Get(buildJobLogURL(strconv.FormatInt(ra.jobID, 10)))
	if err != nil {
		log.Errorf("failed to get log for job %d: %v", ra.jobID, err)
		ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if resp.StatusCode == http.StatusOK {
		ra.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), resp.Header.Get(http.CanonicalHeaderKey("Content-Length")))
		ra.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), "text/plain")

		if _, err = io.Copy(ra.Ctx.ResponseWriter, resp.Body); err != nil {
			log.Errorf("failed to write log to response; %v", err)
			ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read reponse body: %v", err)
		ra.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	ra.CustomAbort(resp.StatusCode, string(b))
}

//TODO:add Post handler to call job service API to submit jobs by policy
