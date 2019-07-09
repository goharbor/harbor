package controllers

import (
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	"time"
)

type RetentionAPI struct {
	api.BaseController
	manager retention.Manager
}

// Prepare validates the user
func (t *RetentionAPI) Prepare() {
	t.BaseController.Prepare()
	t.manager = retention.NewManager()
}

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

func (r *RetentionAPI) CreateRetention() {
	p := &policy.Metadata{}
	isValid, err := r.DecodeJSONReqAndValidate(p)
	if !isValid {
		r.SendBadRequestError(err)
		return
	}

	r.manager.CreatePolicy(p)
}

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

func (r *RetentionAPI) DeleteRetention() {
	id, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.manager.DeletePolicy(id)
}

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
	query := &q.Query{
		PageNumber: page,
		PageSize:   size,
	}
	his, err := r.manager.ListHistories(eid, query)
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.Data["json"] = his
	r.ServeJSON()
}
