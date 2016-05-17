package api

import (
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/utils/log"
	"net/http"
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

//TODO:add Post handler to call job service API to submit jobs by policy
