package dao

import (
	"github.com/stretchr/testify/require"
	"testing"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"time"
)

func TestMethodOfExecution(t *testing.T) {
	execution1 := &models.Execution{
		PolicyID: 11209,
		Status: "InProgress",
		StatusText: "None",
		Total: 12,
		Failed: 0,
		Succeed: 7,
		InProgress: 5,
		Stopped: 0,
		Trigger: "Event",
		StartTime: time.Now(),
	}
	execution2 := &models.Execution{
		PolicyID: 11209,
		Status: "Failed",
		StatusText: "Network error",
		Total: 9,
		Failed: 1,
		Succeed: 8,
		InProgress: 0,
		Stopped: 0,
		Trigger: "Manual",
		StartTime: time.Now(),
	}

	// test add
	id1, err := AddExecution(execution1)
	require.Nil(t, err)

	_, err = AddExecution(execution2)
	require.Nil(t, err)

	// test list
	query := &models.ExecutionQuery{
		Statuses: []string{"InProgress", "Failed"},
		Pagination: models.Pagination{
			Page: 1,
			Size: 10,
		},
	}
	executions, err := GetExecutions(query)
	require.Nil(t, err)
	assert.Equal(t, 2, len(executions))

	total, err := GetTotalOfExecutions(query)
	require.Nil(t, err)
	assert.Equal(t, int64(2), total)

	// test get
	execution, err := GetExecution(id1)
	require.Nil(t, err)
	assert.Equal(t, execution1.Status, execution.Status)

	// test update
	executionNew := &models.Execution{
		ID: id1,
		Status: "Succeed",
		Succeed: 12,
		InProgress: 0,
		EndTime: time.Now(),
	}
	n, err := UpdateExecution(executionNew, models.ExecutionPropsName.Status, models.ExecutionPropsName.Succeed, models.ExecutionPropsName.InProgress,
		models.ExecutionPropsName.EndTime)
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)

	// test delete
	require.Nil(t, DeleteExecution(execution1.ID))
	execution, err = GetExecution(execution1.ID)
	require.Nil(t, err)
	require.Nil(t, execution)

	// test delete all
	require.Nil(t, DeleteAllExecutions(execution1.PolicyID))
	query = &models.ExecutionQuery{
	}
	n, err = GetTotalOfExecutions(query)
	require.Nil(t, err)
	assert.Equal(t, int64(0), n)
}

func TestMethodOfTask(t *testing.T) {
	task1 := &models.Task{
		ExecutionID: 112200,
		ResourceType: "resourceType1",
		SrcResource: "srcResource1",
		DstResource: "dstResource1",
		JobID: "jobID1",
		Status: "Initialized",
		StartTime: time.Now(),
	}
	task2 := &models.Task{
		ExecutionID: 112200,
		ResourceType: "resourceType2",
		SrcResource: "srcResource2",
		DstResource: "dstResource2",
		JobID: "jobID2",
		Status: "Stopped",
		StartTime: time.Now(),
		EndTime: time.Now(),
	}

	// test add
	id1, err := AddTask(task1)
	require.Nil(t, err)

	_, err = AddTask(task2)
	require.Nil(t, err)

	// test list
	query := &models.TaskQuery{
		ResourceType: "resourceType1",
		Pagination: models.Pagination{
			Page: 1,
			Size: 10,
		},
	}
	tasks, err := GetTasks(query)
	require.Nil(t, err)
	assert.Equal(t, 1, len(tasks))

	total, err := GetTotalOfTasks(query)
	require.Nil(t, err)
	assert.Equal(t, int64(1), total)

	// test get
	task, err := GetTask(id1)
	require.Nil(t, err)
	assert.Equal(t, task1.Status, task.Status)

	// test update
	taskNew := &models.Task{
		ID: id1,
		Status: "Failed",
		EndTime: time.Now(),
	}
	n, err := UpdateTask(taskNew, models.TaskPropsName.Status, models.TaskPropsName.EndTime)
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)

	// test delete
	require.Nil(t, DeleteTask(id1))
	task, err = GetTask(id1)
	require.Nil(t, err)
	require.Nil(t, task)

	// test delete all
	require.Nil(t, DeleteAllTasks(task1.ExecutionID))
	query = &models.TaskQuery{
	}
	n, err = GetTotalOfTasks(query)
	require.Nil(t, err)
	assert.Equal(t, int64(0), n)
}

func TestUpdateJobStatus(t *testing.T) {
	execution := &models.Execution{
		PolicyID: 11209,
		Status: "InProgress",
		StatusText: "None",
		Total: 12,
		Failed: 0,
		Succeed: 10,
		InProgress: 1,
		Stopped: 1,
		Trigger: "Event",
		StartTime: time.Now(),
	}
	executionID, _ := AddExecution(execution)
	task1 := &models.Task{
		ID: 20191,
		ExecutionID: executionID,
		ResourceType: "resourceType1",
		SrcResource: "srcResource1",
		DstResource: "dstResource1",
		JobID: "jobID1",
		Status: "Pending",
		StartTime: time.Now(),
	}
	task2 := &models.Task{
		ID: 20192,
		ExecutionID: executionID,
		ResourceType: "resourceType2",
		SrcResource: "srcResource2",
		DstResource: "dstResource2",
		JobID: "jobID2",
		Status: "Stopped",
		StartTime: time.Now(),
		EndTime: time.Now(),
	}
	taskID1, _ := AddTask(task1)
	taskID2, _ := AddTask(task2)

	defer func() {
		DeleteAllTasks(executionID)
		DeleteAllExecutions(11209)
	}()

	// update Pending->InProgress
	n, err := UpdateTaskStatus(taskID1, "InProgress", "Pending")
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)

	execu, err := GetExecution(executionID)
	require.Nil(t, err)
	assert.Equal(t, execution.InProgress, execu.InProgress)
	assert.Equal(t, execution.Status, execu.Status)

	// update InProgress->Failed: Execution.InProgress-1, Failed+1
	n, err = UpdateTaskStatus(taskID1, "Failed")
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)

	execu, err = GetExecution(executionID)
	require.Nil(t, err)
	assert.Equal(t, 1, execu.Failed)
	assert.Equal(t, "Failed", execu.Status)

	// update Stopped->Pending: Execution.Stopped-1, InProgress+1
	n, err = UpdateTaskStatus(taskID2, "Pending")
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)

	execu, err = GetExecution(executionID)
	require.Nil(t, err)
	assert.Equal(t, 1, execu.InProgress)
	assert.Equal(t, "InProgress", execu.Status)
}
