package dao

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/q"
)

func CreatePolicy(p *models.RetentionPolicy) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(p)
}

func UpdatePolicy(p *models.RetentionPolicy) error {
	o := dao.GetOrmer()
	_, err := o.Update(p)
	return err
}

func DeletePolicy(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.RetentionPolicy{
		ID: id,
	})
	return err
}

func GetPolicy(id int64) (*models.RetentionPolicy, error) {
	o := dao.GetOrmer()
	p := &models.RetentionPolicy{
		ID: id,
	}
	if err := o.Read(p); err != nil {
		return nil, err
	} else {
		return p, nil
	}
}

func CreateExecution(e *models.RetentionExecution) (int64, error) {
	o := dao.GetOrmer()
	return o.Insert(e)
}

func UpdateExecution(e *models.RetentionExecution) error {
	o := dao.GetOrmer()
	_, err := o.Update(e)
	return err
}

func DeleteExecution(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.RetentionExecution{
		ID: id,
	})
	return err
}

func GetExecution(id int64) (*models.RetentionExecution, error) {
	o := dao.GetOrmer()
	e := &models.RetentionExecution{
		ID: id,
	}
	if err := o.Read(e); err != nil {
		return nil, err
	} else {
		return e, nil
	}
}

func ListExecutions(query *q.Query) ([]*models.RetentionExecution, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.RetentionExecution))
	qs.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
	var execs []*models.RetentionExecution
	_, err := qs.All(&execs)
	if err != nil {
		return nil, err
	} else {
		return execs, nil
	}
}

func ListExecHistories(executionID int64, query *q.Query) ([]*models.RetentionTask, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.RetentionTask))
	qs.Filter("Execution_ID", executionID)
	qs.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
	var tasks []*models.RetentionTask
	_, err := qs.All(&tasks)
	if err != nil {
		return nil, err
	} else {
		return tasks, nil
	}
}
