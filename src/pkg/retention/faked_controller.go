package retention

import (
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// FakedRetentionController ...
type FakedRetentionController struct {
}

// GetRetention ...
func (f *FakedRetentionController) GetRetention(id int64) (*policy.Metadata, error) {
	return &policy.Metadata{
		ID:        1,
		Algorithm: "",
		Rules:     nil,
		Trigger:   nil,
		Scope:     nil,
	}, nil
}

// CreateRetention ...
func (f *FakedRetentionController) CreateRetention(p *policy.Metadata) (int64, error) {
	return 0, nil
}

// UpdateRetention ...
func (f *FakedRetentionController) UpdateRetention(p *policy.Metadata) error {
	return nil
}

// DeleteRetention ...
func (f *FakedRetentionController) DeleteRetention(id int64) error {
	return nil
}

// TriggerRetentionExec ...
func (f *FakedRetentionController) TriggerRetentionExec(policyID int64, trigger string, dryRun bool) (int64, error) {

	return 0, nil
}

// OperateRetentionExec ...
func (f *FakedRetentionController) OperateRetentionExec(eid int64, action string) error {
	return nil
}

// GetRetentionExec ...
func (f *FakedRetentionController) GetRetentionExec(eid int64) (*Execution, error) {
	return &Execution{
		DryRun:   false,
		PolicyID: 1,
	}, nil
}

// ListRetentionExecs ...
func (f *FakedRetentionController) ListRetentionExecs(policyID int64, query *q.Query) ([]*Execution, error) {
	return nil, nil
}

// GetTotalOfRetentionExecs ...
func (f *FakedRetentionController) GetTotalOfRetentionExecs(policyID int64) (int64, error) {
	return 0, nil
}

// ListRetentionExecTasks ...
func (f *FakedRetentionController) ListRetentionExecTasks(executionID int64, query *q.Query) ([]*Task, error) {
	return nil, nil
}

// GetTotalOfRetentionExecTasks ...
func (f *FakedRetentionController) GetTotalOfRetentionExecTasks(executionID int64) (int64, error) {
	return 0, nil
}

// GetRetentionExecTaskLog ...
func (f *FakedRetentionController) GetRetentionExecTaskLog(taskID int64) ([]byte, error) {
	return nil, nil
}

// GetRetentionExecTask ...
func (f *FakedRetentionController) GetRetentionExecTask(taskID int64) (*Task, error) {
	return &Task{
		ID:          1,
		ExecutionID: 1,
	}, nil
}
