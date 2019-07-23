package controllers

import (
	"time"

	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// RetentionAPI ...
type RetentionAPI struct {
	api.BaseController
	manager retention.Manager
}

// Prepare validates the user
func (r *RetentionAPI) Prepare() {
	r.BaseController.Prepare()
	r.manager = retention.NewManager()
}

// GetRetention Get Retention
func (r *RetentionAPI) GetRetention() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	p, err := r.manager.GetPolicy(id)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.Data["json"] = p
	r.ServeJSON()
}

// CreateRetention Create Retention
func (r *RetentionAPI) CreateRetention() {
	p := &policy.Metadata{}
	isValid, err := r.DecodeJSONReqAndValidate(p)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}

	r.manager.CreatePolicy(p)
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
	r.manager.UpdatePolicy(p)
}

// DeleteRetention Delete Retention
func (r *RetentionAPI) DeleteRetention() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.manager.DeletePolicy(id)
}

// TriggerRetentionExec Trigger Retention Execution
func (r *RetentionAPI) TriggerRetentionExec() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	exec := &retention.Execution{
		PolicyID:  id,
		StartTime: time.Now(),
		Status:    "Running",
	}
	r.manager.CreateExecution(exec)
}

// OperateRetentionExec Operate Retention Execution
func (r *RetentionAPI) OperateRetentionExec() {
	eid, err := r.GetSpecialIDFromURL("eid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	exec := &retention.Execution{}
	isValid, err := r.DecodeJSONReqAndValidate(exec)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}
	exec.ID = eid
	r.manager.UpdateExecution(nil)
}

// GetRetentionExec Get Retention Execution
func (r *RetentionAPI) GetRetentionExec() {
	eid, err := r.GetSpecialIDFromURL("eid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	exec, err := r.manager.GetExecution(eid)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.Data["json"] = exec
	r.ServeJSON()
}

// ListRetentionExec List Retention Execution
func (r *RetentionAPI) ListRetentionExec() {
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	query := &q.Query{
		PageNumber: page,
		PageSize:   size,
	}
	execs, err := r.manager.ListExecutions(query)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.Data["json"] = execs
	r.ServeJSON()
}

// ListRetentionExecHistory List Retention Execution Histories
func (r *RetentionAPI) ListRetentionExecHistory() {
	eid, err := r.GetSpecialIDFromURL("eid")
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	page, size, err := r.GetPaginationParams()
	if err != nil {
		r.SendInternalServerError(err)
		return
	}
	tasks, err := r.manager.ListTasks(&q.TaskQuery{
		ExecutionID: eid,
		PageNumber:  page,
		PageSize:    size,
	})
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.Data["json"] = tasks
	r.ServeJSON()
}
