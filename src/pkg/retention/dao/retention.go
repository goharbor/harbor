package dao

import (
	"errors"
	"fmt"
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
func UpdatePolicy(p *models.RetentionPolicy, cols ...string) error {
	o := dao.GetOrmer()
	_, err := o.Update(p, cols...)
	return err
}

// DeletePolicyAndExec Delete Policy and Exec
func DeletePolicyAndExec(id int64) error {
	o := dao.GetOrmer()
	if _, err := o.Raw("delete from retention_task where execution_id in (select id from retention_execution where policy_id = ?) ", id).Exec(); err != nil {
		return nil
	}
	if _, err := o.Raw("delete from retention_execution where policy_id = ?", id).Exec(); err != nil {
		return err
	}
	if _, err := o.Delete(&models.RetentionExecution{
		PolicyID: id,
	}); err != nil {
		return err
	}
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
func UpdateExecution(e *models.RetentionExecution, cols ...string) error {
	o := dao.GetOrmer()
	_, err := o.Update(e, cols...)
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
	if err := fillStatus(e); err != nil {
		return nil, err
	}
	return e, nil
}

// fillStatus the priority is InProgress Stopped Failed Succeed
func fillStatus(exec *models.RetentionExecution) error {
	o := dao.GetOrmer()
	if _, err := o.Raw("select status, count(*) num from retention_task where execution_id = ? group by status", exec.ID).
		RowsToStruct(exec, "status", "num"); err != nil {
		return err
	}
	exec.Total = exec.Pending + exec.InProgress + exec.Succeed + exec.Failed + exec.Stopped
	if exec.Total == 0 {
		exec.Status = models.ExecutionStatusSucceed
		exec.EndTime = exec.StartTime
		return nil
	}
	if exec.Pending+exec.InProgress > 0 {
		exec.Status = models.ExecutionStatusInProgress
	} else if exec.Stopped > 0 {
		exec.Status = models.ExecutionStatusStopped
	} else if exec.Failed > 0 {
		exec.Status = models.ExecutionStatusFailed
	} else {
		exec.Status = models.ExecutionStatusSucceed
	}
	if exec.Status != models.ExecutionStatusInProgress {
		if err := o.Raw("select max(end_time) from retention_task where execution_id = ?", exec.ID).
			QueryRow(&exec.EndTime); err != nil {
			return err
		}
	}
	return nil
}

// ListExecutions List Executions
func ListExecutions(policyID int64, query *q.Query) ([]*models.RetentionExecution, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.RetentionExecution))

	qs = qs.Filter("policy_id", policyID)
	qs = qs.OrderBy("-id")
	if query != nil {
		qs = qs.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
	}
	var execs []*models.RetentionExecution
	_, err := qs.All(&execs)
	if err != nil {
		return nil, err
	}
	for _, e := range execs {
		if err := fillStatus(e); err != nil {
			return nil, err
		}
	}
	return execs, nil
}

/*
// ListExecHistories List Execution Histories
func ListExecHistories(executionID int64, query *q.Query) ([]*models.RetentionTask, error) {
	o := dao.GetOrmer()
	qs := o.QueryTable(new(models.RetentionTask))
	qs = qs.Filter("Execution_ID", executionID)
	if query != nil {
		qs = qs.Limit(query.PageSize, (query.PageNumber-1)*query.PageSize)
	}
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
*/

// CreateTask creates task record in database
func CreateTask(task *models.RetentionTask) (int64, error) {
	if task == nil {
		return 0, errors.New("nil task")
	}
	return dao.GetOrmer().Insert(task)
}

// UpdateTask updates the task record in database
func UpdateTask(task *models.RetentionTask, cols ...string) error {
	if task == nil {
		return errors.New("nil task")
	}
	if task.ID <= 0 {
		return fmt.Errorf("invalid task ID: %d", task.ID)
	}
	_, err := dao.GetOrmer().Update(task, cols...)
	return err
}

// DeleteTask deletes the task record specified by ID in database
func DeleteTask(id int64) error {
	_, err := dao.GetOrmer().Delete(&models.RetentionTask{
		ID: id,
	})
	return err
}

// GetTask get the task record specified by ID in database
func GetTask(id int64) (*models.RetentionTask, error) {
	task := &models.RetentionTask{
		ID: id,
	}
	if err := dao.GetOrmer().Read(task); err != nil {
		return nil, err
	}
	return task, nil
}

// ListTask lists the tasks according to the query
func ListTask(query ...*q.TaskQuery) ([]*models.RetentionTask, error) {
	qs := dao.GetOrmer().QueryTable(&models.RetentionTask{})
	if len(query) > 0 && query[0] != nil {
		q := query[0]
		if q.ExecutionID > 0 {
			qs = qs.Filter("ExecutionID", q.ExecutionID)
		}
		if len(q.Status) > 0 {
			qs = qs.Filter("Status", q.Status)
		}
		if q.PageSize > 0 {
			qs = qs.Limit(q.PageSize)
			if q.PageNumber > 0 {
				qs = qs.Offset((q.PageNumber - 1) * q.PageSize)
			}
		}
	}
	tasks := []*models.RetentionTask{}
	_, err := qs.All(&tasks)
	return tasks, err
}
