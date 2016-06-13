package api

import (
	"fmt"

	"net/http"
	"strconv"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

// RepPolicyAPI handles /api/replicationPolicies /api/replicationPolicies/:id/enablement
type RepPolicyAPI struct {
	BaseAPI
	policyID int64
	policy   *models.RepPolicy
}

// Prepare validates whether the user has system admin role
// and parsed the policy ID if it exists
func (pa *RepPolicyAPI) Prepare() {
	uid := pa.ValidateUser()
	var err error
	isAdmin, err := dao.IsAdminRole(uid)
	if err != nil {
		log.Errorf("Failed to Check if the user is admin, error: %v, uid: %d", err, uid)
	}
	if !isAdmin {
		pa.CustomAbort(http.StatusForbidden, "")
	}
}

// Get ...
func (pa *RepPolicyAPI) Get() {
	id := pa.GetIDFromURL()
	policy, err := dao.GetRepPolicy(id)
	if err != nil {
		log.Errorf("failed to get policy %d: %v", id, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if policy == nil {
		pa.CustomAbort(http.StatusNotFound, http.StatusText(http.StatusNotFound))
	}

	pa.Data["json"] = policy
	pa.ServeJSON()
}

// List filters policies by name and project_id, if name and project_id
// are nil, List returns all policies
func (pa *RepPolicyAPI) List() {
	name := pa.GetString("name")
	projectIDStr := pa.GetString("project_id")

	var projectID int64
	var err error

	if len(projectIDStr) != 0 {
		projectID, err = strconv.ParseInt(projectIDStr, 10, 64)
		if err != nil || projectID <= 0 {
			pa.CustomAbort(http.StatusBadRequest, "invalid project ID")
		}
	}

	policies, err := dao.FilterRepPolicies(name, projectID)
	if err != nil {
		log.Errorf("failed to filter policies %s project ID %d: %v", name, projectID, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	pa.Data["json"] = policies
	pa.ServeJSON()
}

// Post creates a policy, and if it is enbled, the replication will be triggered right now.
func (pa *RepPolicyAPI) Post() {
	policy := &models.RepPolicy{}
	pa.DecodeJSONReqAndValidate(policy)

	po, err := dao.GetRepPolicyByName(policy.Name)
	if err != nil {
		log.Errorf("failed to get policy %s: %v", policy.Name, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if po != nil {
		pa.CustomAbort(http.StatusConflict, "name is already used")
	}

	project, err := dao.GetProjectByID(policy.ProjectID)
	if err != nil {
		log.Errorf("failed to get project %d: %v", policy.ProjectID, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if project == nil {
		pa.CustomAbort(http.StatusBadRequest, fmt.Sprintf("project %d does not exist", policy.ProjectID))
	}

	target, err := dao.GetRepTarget(policy.TargetID)
	if err != nil {
		log.Errorf("failed to get target %d: %v", policy.TargetID, err)
		pa.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if target == nil {
		pa.CustomAbort(http.StatusBadRequest, fmt.Sprintf("target %d does not exist", policy.TargetID))
	}

	pid, err := dao.AddRepPolicy(*policy)
	if err != nil {
		log.Errorf("Failed to add policy to DB, error: %v", err)
		pa.RenderError(http.StatusInternalServerError, "Internal Error")
		return
	}

	if policy.Enabled == 1 {
		go func() {
			if err := TriggerReplication(pid, "", nil, models.RepOpTransfer); err != nil {
				log.Errorf("failed to trigger replication of %d: %v", pid, err)
			} else {
				log.Infof("replication of %d triggered", pid)
			}
		}()
	}

	pa.Redirect(http.StatusCreated, strconv.FormatInt(pid, 10))
}

type enablementReq struct {
	Enabled int `json:"enabled"`
}

// UpdateEnablement changes the enablement of the policy
func (pa *RepPolicyAPI) UpdateEnablement() {
	e := enablementReq{}
	pa.DecodeJSONReq(&e)
	if e.Enabled != 0 && e.Enabled != 1 {
		pa.RenderError(http.StatusBadRequest, "invalid enabled value")
		return
	}

	if pa.policy.Enabled == e.Enabled {
		return
	}

	if err := dao.UpdateRepPolicyEnablement(pa.policyID, e.Enabled); err != nil {
		log.Errorf("Failed to update policy enablement in DB, error: %v", err)
		pa.RenderError(http.StatusInternalServerError, "Internal Error")
		return
	}

	if e.Enabled == 1 {
		go func() {
			if err := TriggerReplication(pa.policyID, "", nil, models.RepOpTransfer); err != nil {
				log.Errorf("failed to trigger replication of %d: %v", pa.policyID, err)
			} else {
				log.Infof("replication of %d triggered", pa.policyID)
			}
		}()
	} else {
		go func() {
			if err := postReplicationAction(pa.policyID, "stop"); err != nil {
				log.Errorf("failed to stop replication of %d: %v", pa.policyID, err)
			} else {
				log.Infof("try to stop replication of %d", pa.policyID)
			}
		}()
	}
}
