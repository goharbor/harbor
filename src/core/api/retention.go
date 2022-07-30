package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/filter"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// RetentionAPI ...
type RetentionAPI struct {
	BaseController
	pm promgr.ProjectManager
}

// Prepare validates the user
func (r *RetentionAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsAuthenticated() {
		r.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	pm, e := filter.GetProjectManager(r.Ctx.Request)
	if e != nil {
		r.SendInternalServerError(e)
		return
	}
	r.pm = pm

}

// GetMetadatas Get Metadatas
func (r *RetentionAPI) GetMetadatas() {
	data := `
{
    "templates": [
        {
            "rule_template": "latestPushedK",
            "display_text": "the most recently pushed # images",
            "action": "retain",
            "params": [
                {
                    "type": "int",
                    "unit": "COUNT",
                    "required": true
                }
            ]
        },
        {
            "rule_template": "latestPulledN",
            "display_text": "the most recently pulled # images",
            "action": "retain",
            "params": [
                {
                    "type": "int",
                    "unit": "COUNT",
                    "required": true
                }
            ]
        },
		{
			"rule_template": "nDaysSinceLastPush",
			"display_text": "pushed within the last # days",
			"action": "retain",
			"params": [
				{
					"type": "int",
					"unit": "DAYS",
					"required": true
				}
			]
		},
		{
			"rule_template": "nDaysSinceLastPull",
			"display_text": "pulled within the last # days",
			"action": "retain",
			"params": [
				{
					"type": "int",
					"unit": "DAYS",
					"required": true
				}
			]
		},
		{
            "rule_template": "always",
            "display_text": "always",
            "action": "retain",
            "params": []
        }
    ],
    "scope_selectors": [
        {
            "display_text": "Repositories",
            "kind": "doublestar",
            "decorations": [
                "repoMatches",
                "repoExcludes"
            ]
        }
    ],
    "tag_selectors": [
        {
            "display_text": "Tags",
            "kind": "doublestar",
            "decorations": [
                "matches",
                "excludes"
            ]
        }
    ]
}
`
	w := r.Ctx.ResponseWriter
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(data))
}

// GetRetention Get Retention
func (r *RetentionAPI) GetRetention() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	p, err := retentionController.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionRead) {
		return
	}
	r.WriteJSONData(p)
}

// CreateRetention Create Retention
func (r *RetentionAPI) CreateRetention() {
	p := &policy.Metadata{}
	isValid, err := r.DecodeJSONReqAndValidate(p)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}
	if len(p.Rules) > 15 {
		r.SendBadRequestError(errors.New("only 15 rules are allowed at most"))
		return
	}
	if err = r.checkRuleConflict(p); err != nil {
		r.SendConflictError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionCreate) {
		return
	}
	switch p.Scope.Level {
	case policy.ScopeLevelProject:
		if p.Scope.Reference <= 0 {
			r.SendBadRequestError(fmt.Errorf("invalid Project id %d", p.Scope.Reference))
			return
		}

		proj, err := r.pm.Get(p.Scope.Reference)
		if err != nil {
			r.SendBadRequestError(err)
		}
		if proj == nil {
			r.SendBadRequestError(fmt.Errorf("invalid Project id %d", p.Scope.Reference))
		}
	default:
		r.SendBadRequestError(fmt.Errorf("scope %s is not support", p.Scope.Level))
		return
	}
	old, err := r.pm.GetMetadataManager().Get(p.Scope.Reference, "retention_id")
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	if old != nil && len(old) > 0 {
		r.SendBadRequestError(fmt.Errorf("project %v already has retention policy %v", p.Scope.Reference, old["retention_id"]))
		return
	}
	id, err := retentionController.CreateRetention(p)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	if err := r.pm.GetMetadataManager().Add(p.Scope.Reference,
		map[string]string{"retention_id": strconv.FormatInt(id, 10)}); err != nil {
		r.SendInternalServerError(err)
	}
	r.Redirect(http.StatusCreated, strconv.FormatInt(id, 10))
}

// UpdateRetention Update Retention
func (r *RetentionAPI) UpdateRetention() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	p := &policy.Metadata{}
	isValid, err := r.DecodeJSONReqAndValidate(p)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}
	p.ID = id
	if len(p.Rules) > 15 {
		r.SendBadRequestError(errors.New("only 15 rules are allowed at most"))
		return
	}
	if err = r.checkRuleConflict(p); err != nil {
		r.SendConflictError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionUpdate) {
		return
	}
	if err := r.requirePolicyAccess(p); err != nil {
		r.SendPermissionError()
		return
	}
	if err = retentionController.UpdateRetention(p); err != nil {
		r.SendInternalServerError(err)
		return
	}
}

func (r *RetentionAPI) checkRuleConflict(p *policy.Metadata) error {
	temp := make(map[string]int)
	for n, rule := range p.Rules {
		rule.ID = 0
		bs, _ := json.Marshal(rule)
		if old, exists := temp[string(bs)]; exists {
			return fmt.Errorf("rule %d is conflict with rule %d", n, old)
		}
		temp[string(bs)] = n
		rule.ID = n
	}
	return nil
}

// TriggerRetentionExec Trigger Retention Execution
func (r *RetentionAPI) TriggerRetentionExec() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	d := &struct {
		DryRun bool `json:"dry_run"`
	}{
		DryRun: false,
	}
	isValid, err := r.DecodeJSONReqAndValidate(d)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}
	p, err := retentionController.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionUpdate) {
		return
	}
	eid, err := retentionController.TriggerRetentionExec(id, retention.ExecutionTriggerManual, d.DryRun)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	r.Redirect(http.StatusCreated, strconv.FormatInt(eid, 10))
}

// OperateRetentionExec Operate Retention Execution
func (r *RetentionAPI) OperateRetentionExec() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	eid, err := r.GetInt64FromPath(":eid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	a := &struct {
		Action string `json:"action" valid:"Required;Match(stop)"`
	}{}
	isValid, err := r.DecodeJSONReqAndValidate(a)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}
	p, err := retentionController.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionUpdate) {
		return
	}
	if err := r.requirePolicyAccess(p); err != nil {
		r.SendPermissionError()
		return
	}
	if err := r.requireExecutionInProject(p, eid); err != nil {
		r.SendPermissionError()
		return
	}
	if err = retentionController.OperateRetentionExec(eid, a.Action); err != nil {
		r.SendInternalServerError(err)
		return
	}
}

// ListRetentionExecs List Retention Execution
func (r *RetentionAPI) ListRetentionExecs() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	query := &q.Query{
		PageNumber: page,
		PageSize:   size,
	}
	p, err := retentionController.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionList) {
		return
	}
	execs, err := retentionController.ListRetentionExecs(id, query)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	total, err := retentionController.GetTotalOfRetentionExecs(id)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	r.SetPaginationHeader(total, query.PageNumber, query.PageSize)
	r.WriteJSONData(execs)
}

// ListRetentionExecTasks List Retention Execution Tasks
func (r *RetentionAPI) ListRetentionExecTasks() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	eid, err := r.GetInt64FromPath(":eid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	query := &q.Query{
		PageNumber: page,
		PageSize:   size,
	}
	p, err := retentionController.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if p == nil {
		r.SendNotFoundError(errors.New("retention policy is not found"))
		return
	}
	if !r.requireAccess(p, rbac.ActionList) {
		return
	}
	if err := r.requirePolicyAccess(p); err != nil {
		r.SendNotFoundError(err)
		return
	}
	if err := r.requireExecutionInProject(p, eid); err != nil {
		r.SendNotFoundError(err)
		return
	}
	his, err := retentionController.ListRetentionExecTasks(eid, query)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	total, err := retentionController.GetTotalOfRetentionExecTasks(eid)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	r.SetPaginationHeader(total, query.PageNumber, query.PageSize)
	r.WriteJSONData(his)
}

// GetRetentionExecTaskLog Get Retention Execution Task log
func (r *RetentionAPI) GetRetentionExecTaskLog() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	eid, err := r.GetInt64FromPath(":eid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	tid, err := r.GetInt64FromPath(":tid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	p, err := retentionController.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if p == nil {
		r.SendNotFoundError(errors.New("retention policy is not found"))
		return
	}
	if !r.requireAccess(p, rbac.ActionRead) {
		return
	}
	if err := r.requirePolicyAccess(p); err != nil {
		r.SendForbiddenError(err)
		return
	}
	if err := r.requireTaskInProject(p, eid, tid); err != nil {
		r.SendForbiddenError(err)
		return
	}
	log, err := retentionController.GetRetentionExecTaskLog(tid)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	w := r.Ctx.ResponseWriter
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write(log)
}

func (r *RetentionAPI) requireAccess(p *policy.Metadata, action rbac.Action, subresources ...rbac.Resource) bool {
	var hasPermission bool

	switch p.Scope.Level {
	case "project":
		if len(subresources) == 0 {
			subresources = append(subresources, rbac.ResourceTagRetention)
		}
		hasPermission, _ = r.HasProjectPermission(p.Scope.Reference, action, subresources...)
	default:
		hasPermission = r.SecurityCtx.IsSysAdmin()
	}

	if !hasPermission {
		if !r.SecurityCtx.IsAuthenticated() {
			r.SendUnAuthorizedError(errors.New("UnAuthorized"))
		} else {
			r.SendForbiddenError(errors.New(r.SecurityCtx.GetUsername()))
		}
		return false
	}

	return true
}

// requirePolicyAccess checks the scope reference whether has the permission to
// the retention policy.
func (r *RetentionAPI) requirePolicyAccess(p *policy.Metadata) error {
	// the id of policy should be consistent with project metadata
	meta, err := r.pm.GetMetadataManager().Get(p.Scope.Reference, "retention_id")
	if err != nil {
		return err
	}
	// validate
	if len(meta["retention_id"]) > 0 {
		// return err if retention id does not match
		if meta["retention_id"] == fmt.Sprintf("%d", p.ID) {
			return nil
		}
	}
	return fmt.Errorf("the project id:%d, its retention policy is not policy: %d", p.Scope.Reference, p.ID)
}

func (r *RetentionAPI) requireExecutionInProject(p *policy.Metadata, executionID int64) error {
	exec, err := retentionMgr.GetExecution(executionID)
	if err != nil {
		return err
	}
	if exec == nil {
		return fmt.Errorf("project: %d, execution id %d not found", p.Scope.Reference, executionID)
	}
	if exec.PolicyID != p.ID {
		return fmt.Errorf("project: %d, execution id %d not found", p.Scope.Reference, executionID)
	}
	return nil
}

func (r *RetentionAPI) requireTaskInProject(p *policy.Metadata, executionID, taskID int64) error {
	notFoundError := fmt.Errorf("task %d not found", taskID)
	if err := r.requireExecutionInProject(p, executionID); err != nil {
		return fmt.Errorf("execution %d not found", executionID)
	}
	tasks, err := retentionController.ListRetentionExecTasks(executionID, nil)
	if err != nil {
		return notFoundError
	}
	for _, t := range tasks {
		if t.ID == taskID {
			if t.ExecutionID != executionID {
				return notFoundError
			}
			return nil
		}
	}
	return notFoundError
}
