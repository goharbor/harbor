package api

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
)

type RepJobAPI struct {
	BaseAPI
}

func (ja *RepJobAPI) Prepare() {
	uid := ja.ValidateUser()
	isAdmin, err := dao.IsAdminRole(uid)
	if err != nil {
		log.Errorf("Failed to Check if the user is admin, error: %v, uid: %d", err, uid)
	}
	if !isAdmin {
		ja.CustomAbort(http.StatusForbidden, "")
	}

}

func (ja *RepJobAPI) Get() {
	policyID, err := ja.GetInt64("policy_id")
	if err != nil {
		log.Errorf("Failed to get policy id, error: %v", err)
		ja.RenderError(http.StatusBadRequest, "Invalid policy id")
		return
	}
	jobs, err := dao.GetRepJobByPolicy(policyID)
	if err != nil {
		log.Errorf("Failed to query job from db, error: %v", err)
		ja.RenderError(http.StatusInternalServerError, "Failed to query job")
		return
	}
	ja.Data["json"] = jobs
	ja.ServeJSON()
}

// GetLog ...
func (ja *RepJobAPI) GetLog() {
	id := ja.Ctx.Input.Param(":id")
	if len(id) == 0 {
		ja.CustomAbort(http.StatusBadRequest, "id is nil")
	}

	resp, err := http.Get(buildJobLogURL(id))
	if err != nil {
		log.Errorf("failed to get log for job %s: %v", id, err)
		ja.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("failed to read response body for job %s: %v", id, err)
		ja.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if resp.StatusCode == http.StatusOK {
		ja.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Disposition"), "attachment; filename=replication_job.log")
		ja.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Type"), resp.Header.Get(http.CanonicalHeaderKey("Content-Type")))
		ja.Ctx.ResponseWriter.Header().Set(http.CanonicalHeaderKey("Content-Length"), strconv.Itoa(len(b)))
		if _, err = ja.Ctx.ResponseWriter.Write(b); err != nil {
			log.Errorf("failed to write log to response; %v", err)
			ja.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}
		return
	}

	ja.CustomAbort(resp.StatusCode, string(b))
}

//TODO:add Post handler to call job service API to submit jobs by policy
