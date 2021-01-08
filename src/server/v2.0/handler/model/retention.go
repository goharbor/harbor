package model

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// RetentionPolicy ...
type RetentionPolicy struct {
	*policy.Metadata
}

// ToSwagger ...
func (s *RetentionPolicy) ToSwagger() *models.RetentionPolicy {
	var result models.RetentionPolicy
	lib.JSONCopy(&result, s)
	return &result
}

// NewRetentionPolicyFromSwagger ...
func NewRetentionPolicyFromSwagger(policy *models.RetentionPolicy) *RetentionPolicy {
	data, err := json.Marshal(policy)
	if err != nil {
		return nil
	}
	var result RetentionPolicy
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil
	}
	return &result
}

// NewRetentionPolicy ...
func NewRetentionPolicy(policy *policy.Metadata) *RetentionPolicy {
	return &RetentionPolicy{policy}
}

// RetentionExec ...
type RetentionExec struct {
	*retention.Execution
}

// ToSwagger ...
func (e *RetentionExec) ToSwagger() *models.RetentionExecution {
	var result models.RetentionExecution
	lib.JSONCopy(&result, e)
	return &result
}

// NewRetentionExec ...
func NewRetentionExec(exec *retention.Execution) *RetentionExec {
	return &RetentionExec{exec}
}

// RetentionTask ...
type RetentionTask struct {
	*retention.Task
}

// ToSwagger ...
func (e *RetentionTask) ToSwagger() *models.RetentionExecutionTask {
	var result models.RetentionExecutionTask
	lib.JSONCopy(&result, e)
	return &result
}

// NewRetentionTask ...
func NewRetentionTask(task *retention.Task) *RetentionTask {
	return &RetentionTask{task}
}
