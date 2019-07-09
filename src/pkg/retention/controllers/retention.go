package controllers

import (
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/pkg/retention"
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
	r.manager.CreatePolicy(nil)
}

func (r *RetentionAPI) UpdateRetention() {
	_, err := r.GetIDFromURL()
	if err != nil {
		r.SendBadRequestError(err)
		return
	}
	r.manager.UpdatePolicy(nil)
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
	//r.manager.CreateExecution(nil)
}

func (r *RetentionAPI) OperateRetentionExec() {
	r.manager.UpdateExecution(nil)
}

func (r *RetentionAPI) GetRetentionExec() {
	//r.manager.GetExecution(eid)
}

func (r *RetentionAPI) ListRetentionExec() {
	r.manager.ListExecutions(nil)
}

func (r *RetentionAPI) ListRetentionExecHistory() {
	//r.manager.ListHistories(eid, nil)
}
