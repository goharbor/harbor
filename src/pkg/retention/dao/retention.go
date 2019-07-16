package dao

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

// CreatePolicy Create Policy
func CreatePolicy(p *models.RetentionPolicy) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(p)
}

// UpdatePolicy Update Policy
func UpdatePolicy(p *models.RetentionPolicy) error {
	o := dao.GetOrmer()
	_, err := o.Update(p)
	return err
}

// DeletePolicy Delete Policy
func DeletePolicy(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.RetentionPolicy{
		ID: id,
	})
	return err
}

// GetPolicy Get Policy
func GetPolicy(id int64) (*models.RetentionPolicy, error) {
	o := dao.GetOrmer()
	p := &models.RetentionPolicy{
		ID: id,
	}
	if err := o.Read(p); err != nil {
		return nil, err
	}
	return p, nil
}

// CreateExecution Create Execution
func CreateExecution(e *models.RetentionExecution) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(e)
}

// UpdateExecution Update Execution
func UpdateExecution(e *models.RetentionExecution) error {
	o := dao.GetOrmer()
	_, err := o.Update(e)
	return err
}

// DeleteExecution Delete Execution
func DeleteExecution(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.RetentionExecution{
		ID: id,
	})
	return err
}

// GetExecution Get Execution
func GetExecution(id int64) (*models.RetentionExecution, error) {
	o := dao.GetOrmer()
	e := &models.RetentionExecution{
		ID: id,
	}
	if err := o.Read(e); err != nil {
		return nil, err
	}
	return e, nil
}

// ListExecutions List Executions
func ListExecutions(query *q.Query) ([]*models.RetentionExecution, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.RetentionExecution))
	qs.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
	var execs []*models.RetentionExecution
	_, err := qs.All(&execs)
	if err != nil {
		return nil, err
	}
	return execs, nil
}

// ListExecHistories List Execution Histories
func ListExecHistories(executionID int64, query *q.Query) ([]*models.RetentionTask, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.RetentionTask))
	qs.Filter("Execution_ID", executionID)
	qs.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
	var tasks []*models.RetentionTask
	_, err := qs.All(&tasks)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// AppendExecHistory Append Execution History
func AppendExecHistory(t *models.RetentionTask) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(t)
}
