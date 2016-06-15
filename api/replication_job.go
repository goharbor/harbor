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

// Get gets all the jobs according to the policy
func (ra *RepJobAPI) Get() {
	policyID, err := ra.GetInt64("policy_id")
	if err != nil {
		log.Errorf("Failed to get policy id, error: %v", err)
		ra.RenderError(http.StatusBadRequest, "Invalid policy id")
		return
	}
	jobs, err := dao.GetRepJobByPolicy(policyID)
	if err != nil {
		log.Errorf("Failed to query job from db, error: %v", err)
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
		for key, values := range resp.Header {
			for _, value := range values {
				ra.Ctx.ResponseWriter.Header().Set(key, value)
			}
		}

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
