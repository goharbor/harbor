package api

import (
	"errors"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/core/filter"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// RetentionAPI ...
type RetentionAPI struct {
	BaseController
	api retention.APIController
}

// Prepare validates the user
func (r *RetentionAPI) Prepare() {
	r.BaseController.Prepare()
	if !r.SecurityCtx.IsAuthenticated() {
		r.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}
	p, e := filter.GetProjectManager(r.Ctx.Request)
	if e != nil {
		r.SendInternalServerError(e)
		return
	}
	r.api = retention.NewAPIController(p, projectMgr, repositoryMgr, retentionScheduler, retentionLauncher)

}

// GetMetadatas Get Metadatas
func (r *RetentionAPI) GetMetadatas() {
	data := `
{
    "templates": [
        {
            "rule_template": "lastXDays",
            "display_text": "the images from the last # days",
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
            "rule_template": "latestActiveK",
            "display_text": "the most recent active # images",
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
            "rule_template": "latestK",
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
            "rule_template": "latestPulledK",
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
            "rule_template": "always",
            "display_text": "always",
            "action": "retain",
            "params": [
                {
                    "type": "int",
                    "unit": "COUNT",
                    "required": true
                }
            ]
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
            "display_text": "Labels",
            "kind": "label",
            "decorations": [
                "withLabels",
                "withoutLabels"
            ]
        },
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
	r.WriteJSONData(data)
}

// GetRetention Get Retention
func (r *RetentionAPI) GetRetention() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	p, err := r.api.GetRetention(id)
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
	if !r.requireAccess(p, rbac.ActionCreate) {
		return
	}
	if err = r.api.CreateRetention(p); err != nil {
		r.SendInternalServerError(err)
		return
	}
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
	if !r.requireAccess(p, rbac.ActionUpdate) {
		return
	}
	if err = r.api.UpdateRetention(p); err != nil {
		r.SendInternalServerError(err)
		return
	}
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
	p, err := r.api.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionUpdate) {
		return
	}
	if err = r.api.TriggerRetentionExec(id, retention.ExecutionTriggerManual, d.DryRun); err != nil {
		r.SendInternalServerError(err)
		return
	}
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
		Action string `json:"action" valid:"Required"`
	}{}
	isValid, err := r.DecodeJSONReqAndValidate(a)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}
	p, err := r.api.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionUpdate) {
		return
	}
	if err = r.api.OperateRetentionExec(eid, a.Action); err != nil {
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
	p, err := r.api.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionList) {
		return
	}
	execs, err := r.api.ListRetentionExecs(id, query)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
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
	p, err := r.api.GetRetention(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	if !r.requireAccess(p, rbac.ActionList) {
		return
	}
	his, err := r.api.ListRetentionExecTasks(eid, query)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	r.WriteJSONData(his)
}

// GetRetentionExecTaskLog Get Retention Execution Task log
func (r *RetentionAPI) GetRetentionExecTaskLog() {
	tid, err := r.GetInt64FromPath(":tid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	log, err := r.api.GetRetentionExecTaskLog(tid)
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	w := r.Ctx.ResponseWriter
	w.Write(log)
}

func (r *RetentionAPI) requireAccess(p *policy.Metadata, action rbac.Action, subresources ...rbac.Resource) bool {
	var hasPermission bool

	switch p.Scope.Level {
	case "project":
		if len(subresources) == 0 {
			subresources = append(subresources, rbac.ResourceTagRetention)
		}
		resource := rbac.NewProjectNamespace(p.Scope.Reference).Resource(subresources...)
		hasPermission = r.SecurityCtx.Can(action, resource)
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
