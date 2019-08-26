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
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/dao/models"
)

// AddExecution ...
func AddExecution(execution *models.Execution) (int64, error) {
	o := dao.GetOrmer()
	now := time.Now()
	execution.StartTime = now

	return o.Insert(execution)
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
	if err != nil || len(executions) == 0 {
		return executions, err
	}
	for _, e := range executions {
		fillExecution(e)
	}
	return executions, err
}

func executionQueryConditions(query ...*models.ExecutionQuery) orm.QuerySeter {
	qs := dao.GetOrmer().QueryTable(new(models.Execution))
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
	o := dao.GetOrmer()
	t := models.Execution{ID: id}
	err := o.Read(&t)
	if err == orm.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	fillExecution(&t)
	return &t, err
}

// fillExecution will fill the statistics data and status by tasks data
func fillExecution(execution *models.Execution) error {
	if executionFinished(execution.Status) {
		return nil
	}

	o := dao.GetOrmer()
	sql := `select status, count(*) as c from replication_task where execution_id = ? group by status`
	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, execution.ID)

	dt := []*models.TaskStat{}
	count, err := o.Raw(sql, queryParam).QueryRows(&dt)

	if err != nil {
		log.Errorf("Query tasks error execution %d: %v", execution.ID, err)
		return err
	}

	if count == 0 {
		return nil
	}

	total := 0
	for _, d := range dt {
		status, _ := getStatus(d.Status)
		updateStatusCount(execution, status, d.C)
		total += d.C
	}

	if execution.Total != total {
		log.Debugf("execution task count inconsistent and fixed, executionID=%d, execution.total=%d, tasks.count=%d",
			execution.ID, execution.Total, total)
		execution.Total = total
	}
	resetExecutionStatus(execution)

	return nil
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
	execution.Status = generateStatus(execution)
	if executionFinished(execution.Status) {
		o := dao.GetOrmer()
		sql := `select max(end_time) from replication_task where execution_id = ?`
		queryParam := make([]interface{}, 1)
		queryParam = append(queryParam, execution.ID)

		var et time.Time
		err := o.Raw(sql, queryParam).QueryRow(&et)
		if err != nil {
			log.Errorf("Query end_time from tasks error execution %d: %v", execution.ID, err)
			et = time.Now()
		}
		execution.EndTime = et
	}
	return nil
}

func generateStatus(execution *models.Execution) string {
	if execution.InProgress > 0 {
		return models.ExecutionStatusInProgress
	} else if execution.Failed > 0 {
		return models.ExecutionStatusFailed
	} else if execution.Stopped > 0 {
		return models.ExecutionStatusStopped
	}
	return models.ExecutionStatusSucceed
}

func executionFinished(status string) bool {
	if status == models.ExecutionStatusStopped ||
		status == models.ExecutionStatusSucceed ||
		status == models.ExecutionStatusFailed {
		return true
	}
	return false
}

// DeleteExecution ...
func DeleteExecution(id int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Execution{ID: id})
	return err
}

// DeleteAllExecutions ...
func DeleteAllExecutions(policyID int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Execution{PolicyID: policyID}, "PolicyID")
	return err
}

// UpdateExecution ...
func UpdateExecution(execution *models.Execution, props ...string) (int64, error) {
	if execution.ID == 0 {
		return 0, fmt.Errorf("execution ID is empty")
	}
	o := dao.GetOrmer()
	return o.Update(execution, props...)
}

// AddTask ...
func AddTask(task *models.Task) (int64, error) {
	o := dao.GetOrmer()
	now := time.Now()
	task.StartTime = &now

	return o.Insert(task)
}

// GetTask ...
func GetTask(id int64) (*models.Task, error) {
	o := dao.GetOrmer()
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
	qs := dao.GetOrmer().QueryTable(new(models.Task))
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
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Task{ID: id})
	return err
}

// DeleteAllTasks ...
func DeleteAllTasks(executionID int64) error {
	o := dao.GetOrmer()
	_, err := o.Delete(&models.Task{ExecutionID: executionID}, "ExecutionID")
	return err
}

// UpdateTask ...
func UpdateTask(task *models.Task, props ...string) (int64, error) {
	if task.ID == 0 {
		return 0, fmt.Errorf("task ID is empty")
	}
	o := dao.GetOrmer()
	return o.Update(task, props...)
}

// UpdateTaskStatus updates the status of task.
// The implementation uses raw sql rather than QuerySetter.Filter... as QuerySetter
// will generate sql like:
//   `UPDATE "replication_task" SET "end_time" = $1, "status" = $2
//     WHERE "id" IN ( SELECT T0."id" FROM "replication_task" T0 WHERE T0."id" = $3
//     AND T0."status" IN ($4, $5, $6))]`
// which is not a "single" sql statement, this will cause issues when running in concurrency
func UpdateTaskStatus(id int64, status string, statusRevision int64, statusCondition ...string) (int64, error) {
	params := []interface{}{}
	sql := `update replication_task set status = ?, status_revision = ?, end_time = ? `
	params = append(params, status, statusRevision)
	var t time.Time
	// when the task is in final status, update the endtime
	// when the task re-runs again, the endtime should be cleared
	// so set the endtime to null if the task isn't in final status
	if taskFinished(status) {
		t = time.Now()
	}
	params = append(params, t)

	sql += fmt.Sprintf(`where id = ? and (status_revision < ? or status_revision = ? and status in (%s)) `, dao.ParamPlaceholderForIn(len(statusCondition)))
	params = append(params, id, statusRevision, statusRevision, statusCondition)

	result, err := dao.GetOrmer().Raw(sql, params...).Exec()
	if err != nil {
		return 0, err
	}
	n, _ := result.RowsAffected()
	if n > 0 {
		log.Debugf("update task status %d: -> %s", id, status)
	}
	return n, err
}

func taskFinished(status string) bool {
	if status == models.TaskStatusFailed || status == models.TaskStatusStopped || status == models.TaskStatusSucceed {
		return true
	}
	return false
}
