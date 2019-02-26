// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"fmt"
)

// AddExecution ...
func AddExecution(execution *models.Execution) (int64, error) {
	o := GetOrmer()

	sql := "insert into replication_execution (policy_id, status, status_text, total, failed, succeed, in_progress, stopped, trigger) " +
		"values (?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id"

	var id int64
	err := o.Raw(sql, execution.PolicyID, execution.Status, execution.StatusText,execution.Total, execution.Failed,
		execution.Succeed, execution.InProgress, execution.Stopped, execution.Trigger).QueryRow(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetTotalOfExecutions returns the total count of replication execution
func GetTotalOfExecutions(query ...*models.ExecutionQuery) (int64, error) {
	qs := executionQueryConditions(query...)
	return qs.Count()
}

// GetExecutions ...
func GetExecutions(query ...*models.ExecutionQuery) ([]*models.Execution, error) {
	executions := []*models.Execution{}

	qs := executionQueryConditions(query...)
	if len(query) > 0 && query[0] != nil {
		qs = paginateForQuerySetter(qs, query[0].Page, query[0].Size)
	}

	qs = qs.OrderBy("-StartTime")

	_, err := qs.All(&executions)
	return executions, err
}

func executionQueryConditions(query ...*models.ExecutionQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(new(models.Execution))
	if len(query) == 0 || query[0] == nil {
		return qs
	}

	q := query[0]
	if q.PolicyID != 0 {
		qs = qs.Filter("PolicyID", q.PolicyID)
	}
	if len(q.Trigger) > 0 {
		qs = qs.Filter("Trigger", q.Trigger)
	}
	if len(q.Statuses) > 0 {
		qs = qs.Filter("Status__in", q.Statuses)
	}
	return qs
}

// GetExecution ...
func GetExecution(id int64) (*models.Execution, error) {
	o := GetOrmer()
	t := models.Execution{ID: id}
	err := o.Read(&t)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	return &t, err
}

// DeleteExecution ...
func DeleteExecution(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.Execution{ID: id})
	return err
}

// DeleteAllExecutions ...
func DeleteAllExecutions(policyID int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.Execution{PolicyID: policyID}, "PolicyID")
	return err
}

// UpdateExecution ...
func UpdateExecution(execution *models.Execution, props ...string) (int64, error) {
	if execution.ID == 0 {
		return 0, fmt.Errorf("execution ID is empty")
	}
	o := GetOrmer()
	return o.Update(execution, props...)
}

// AddTask ...
func AddTask(task *models.Task) (int64, error) {
	o := GetOrmer()
	sql := `insert into replication_task (execution_id, resource_type, src_resource, dst_resource, job_id, status) 
				values (?, ?, ?, ?, ?, ?) RETURNING id`

	args := []interface{}{}
	args = append(args, task.ExecutionID, task.ResourceType, task.SrcResource, task.DstResource, task.JobID, task.Status)

	var taskID int64
	err := o.Raw(sql, args).QueryRow(&taskID)
	if err != nil {
		return 0, err
	}

	return taskID, nil
}

// GetTask ...
func GetTask(id int64) (*models.Task, error) {
	o := GetOrmer()
	sql := `select * from replication_task where id = ?`

	var task models.Task

	if err := o.Raw(sql, id).QueryRow(&task); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &task, nil
}

// GetTotalOfTasks ...
func GetTotalOfTasks(query ...*models.TaskQuery) (int64, error) {
	qs := taskQueryConditions(query...)
	return qs.Count()
}

// GetTasks ...
func GetTasks(query ...*models.TaskQuery) ([]*models.Task, error) {
	tasks := []*models.Task{}

	qs := taskQueryConditions(query...)
	if len(query) > 0 && query[0] != nil {
		qs = paginateForQuerySetter(qs, query[0].Page, query[0].Size)
	}

	qs = qs.OrderBy("-StartTime")

	_, err := qs.All(&tasks)
	return tasks, err
}

func taskQueryConditions(query ...*models.TaskQuery) orm.QuerySeter {
	qs := GetOrmer().QueryTable(new(models.Task))
	if len(query) == 0 || query[0] == nil {
		return qs
	}

	q := query[0]
	if q.ExecutionID != 0 {
		qs = qs.Filter("ExecutionID", q.ExecutionID)
	}
	if len(q.JobID) > 0 {
		qs = qs.Filter("JobID", q.JobID)
	}
	if len(q.ResourceType) > 0 {
		qs = qs.Filter("ResourceType", q.ResourceType)
	}
	if len(q.Statuses) > 0 {
		qs = qs.Filter("Status__in", q.Statuses)
	}
	return qs
}

// DeleteTask ...
func DeleteTask(id int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.Task{ID: id})
	return err
}

// DeleteAllTasks ...
func DeleteAllTasks(executionID int64) error {
	o := GetOrmer()
	_, err := o.Delete(&models.Task{ExecutionID: executionID}, "ExecutionID")
	return err
}

// UpdateTask ...
func UpdateTask(task *models.Task, props ...string) (int64, error) {
	if task.ID == 0 {
		return 0, fmt.Errorf("task ID is empty")
	}
	o := GetOrmer()
	return o.Update(task, props...)
}

// UpdateTaskStatus ...
func UpdateTaskStatus(id int64, status string, statusCondition ...string) (int64, error) {
	// can not use the globalOrm
	o := orm.NewOrm()
	o.Begin()

	// query the task status
	var task models.Task
	sql := `select * from replication_task where id = ?`
	if err := o.Raw(sql, id).QueryRow(&task); err != nil {
		if err == orm.ErrNoRows {
			o.Rollback()
			return 0, err
		}
	}

	// check status
	satisfy := false
	if len(statusCondition) == 0 {
		satisfy = true
	} else {
		for _, stCondition := range statusCondition {
			if task.Status == stCondition {
				satisfy = true
				break
			}
		}
	}
	if !satisfy {
		o.Rollback()
		return 0, fmt.Errorf("Status condition not match ")
	}

	// update status
	params := []interface{}{}
	sql = `update replication_task set status = ?`
	params = append(params, status)
	if taskFinished(status) { // should update endTime
		sql += ` ,end_time = ?`
		params = append(params, time.Now())
	}
	sql += ` where id = ?`
	params = append(params, id)
	_, err := o.Raw(sql, params).Exec()
	log.Infof("Update task %d: %s -> %s", id, task.Status, status)
	if err != nil{
		log.Errorf("Update task failed %d: %s -> %s", id, task.Status, status)
		o.Rollback()
		return 0, err
	}

	// query the execution
	var execution models.Execution
	sql = `select * from replication_execution where id = ?`
	if err := o.Raw(sql, task.ExecutionID).QueryRow(&execution); err != nil {
		if err == orm.ErrNoRows {
			log.Errorf("Execution not found id: %d", task.ExecutionID)
			o.Rollback()
			return 0, err
		}
	}
	// check execution data
	execuStatus, _ := getStatus(task.Status)
	count := getStatusCount(&execution, execuStatus)
	if count <= 0 {
		log.Errorf("Task statistics in execution inconsistent")
		o.Commit()
		return 1, nil
	}

	// update execution data
	updateStatusCount(&execution, execuStatus, -1)
	execuStatusUp, _ := getStatus(status)
	updateStatusCount(&execution, execuStatusUp, 1)

	resetExecutionStatus(&execution)
	_, err = o.Update(&execution, models.ExecutionPropsName.Status, models.ExecutionPropsName.Total, models.ExecutionPropsName.InProgress,
		models.ExecutionPropsName.Failed, models.ExecutionPropsName.Succeed, models.ExecutionPropsName.Stopped,
		models.ExecutionPropsName.EndTime)
	if err != nil {
		log.Errorf("Update execution status failed %d: %v", execution.ID, err)
		o.Rollback()
		return 0, err
	}
	o.Commit()
	return 1, nil
}

func taskFinished(status string) bool {
	if status == models.TaskStatusFailed || status == models.TaskStatusStopped || status == models.TaskStatusSucceed {
		return true
	}
	return false
}

func getStatus(status string) (string, error) {
	switch status {
	case models.TaskStatusInitialized, models.TaskStatusPending, models.TaskStatusInProgress:
		return models.ExecutionStatusInProgress, nil
	case models.TaskStatusSucceed:
		return models.ExecutionStatusSucceed, nil
	case models.TaskStatusStopped:
		return models.ExecutionStatusStopped, nil
	case models.TaskStatusFailed:
		return models.ExecutionStatusFailed, nil
	}
	return "", fmt.Errorf("Not support task status ")
}

func getStatusCount(execution *models.Execution, status string) int {
	switch status {
	case models.ExecutionStatusInProgress:
		return execution.InProgress
	case models.ExecutionStatusSucceed:
		return execution.Succeed
	case models.ExecutionStatusStopped:
		return execution.Stopped
	case models.ExecutionStatusFailed:
		return execution.Failed
	}
	return 0
}

func updateStatusCount(execution *models.Execution, status string, delta int) error {
	switch status {
	case models.ExecutionStatusInProgress:
		execution.InProgress += delta
	case models.ExecutionStatusSucceed:
		execution.Succeed += delta
	case models.ExecutionStatusStopped:
		execution.Stopped += delta
	case models.ExecutionStatusFailed:
		execution.Failed += delta
	}
	return nil
}

func resetExecutionStatus(execution *models.Execution) error {
	status := generateStatus(execution)
	if status != execution.Status {
		execution.Status = status
		log.Debugf("Execution status changed %d: %s -> %s", execution.ID, execution.Status, status)
	}
	if n := getStatusCount(execution, models.ExecutionStatusInProgress); n == 0 {
		// execution finished in this time
		execution.EndTime = time.Now()
	}
	return nil
}

func generateStatus(execution *models.Execution) string {
	if execution.InProgress > 0 {
		return models.ExecutionStatusInProgress
	}else if execution.Failed > 0 {
		return models.ExecutionStatusFailed
	} else if execution.Stopped > 0 {
		return models.ExecutionStatusStopped
	} else {
		return models.ExecutionStatusSucceed
	}
}
