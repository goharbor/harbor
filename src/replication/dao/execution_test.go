package dao

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMethodOfExecution(t *testing.T) {
	execution1 := &models.Execution{
		PolicyID:   11209,
		Status:     "InProgress",
		StatusText: "None",
		Total:      12,
		Failed:     0,
		Succeed:    7,
		InProgress: 5,
		Stopped:    0,
		Trigger:    "Event",
		StartTime:  time.Now(),
	}
	execution2 := &models.Execution{
		PolicyID:   11209,
		Status:     "Failed",
		StatusText: "Network error",
		Total:      9,
		Failed:     1,
		Succeed:    8,
		InProgress: 0,
		Stopped:    0,
		Trigger:    "Manual",
		StartTime:  time.Now(),
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
		ID:         id1,
		Status:     "Succeed",
		Succeed:    12,
		InProgress: 0,
		EndTime:    time.Now(),
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
	query = &models.ExecutionQuery{}
	n, err = GetTotalOfExecutions(query)
	require.Nil(t, err)
	assert.Equal(t, int64(0), n)
}

func TestMethodOfTask(t *testing.T) {
	now := time.Now()
	task1 := &models.Task{
		ExecutionID:    112200,
		ResourceType:   "resourceType1",
		SrcResource:    "srcResource1",
		DstResource:    "dstResource1",
		JobID:          "jobID1",
		Status:         "Initialized",
		StatusRevision: 1,
		StartTime:      now,
	}
	task2 := &models.Task{
		ExecutionID:    112200,
		ResourceType:   "resourceType2",
		SrcResource:    "srcResource2",
		DstResource:    "dstResource2",
		JobID:          "jobID2",
		Status:         "Stopped",
		StatusRevision: 1,
		StartTime:      now,
		EndTime:        now,
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
		ID:      id1,
		Status:  "Failed",
		EndTime: now,
	}
	n, err := UpdateTask(taskNew, models.TaskPropsName.Status, models.TaskPropsName.EndTime)
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)

	// test update status
	n, err = UpdateTaskStatus(id1, "Succeed", 2, "Initialized")
	require.Nil(t, err)
	assert.Equal(t, int64(1), n)
	task, _ = GetTask(id1)
	assert.Equal(t, "Succeed", task.Status)
	assert.Equal(t, int64(2), task.StatusRevision)

	// test delete
	require.Nil(t, DeleteTask(id1))
	task, err = GetTask(id1)
	require.Nil(t, err)
	require.Nil(t, task)

	// test delete all
	require.Nil(t, DeleteAllTasks(task1.ExecutionID))
	query = &models.TaskQuery{}
	n, err = GetTotalOfTasks(query)
	require.Nil(t, err)
	assert.Equal(t, int64(0), n)
}

func TestExecutionFill(t *testing.T) {
	now := time.Now()
	execution := &models.Execution{
		PolicyID:   11209,
		Status:     "InProgress",
		StatusText: "None",
		Total:      2,
		Trigger:    "Event",
		StartTime:  time.Now(),
	}
	executionID, _ := AddExecution(execution)
	et1, _ := time.Parse("2006-01-02 15:04:05", "2019-03-21 08:01:01")
	et2, _ := time.Parse("2006-01-02 15:04:05", "2019-04-01 10:11:53")
	task1 := &models.Task{
		ID:           20191,
		ExecutionID:  executionID,
		ResourceType: "resourceType1",
		SrcResource:  "srcResource1",
		DstResource:  "dstResource1",
		JobID:        "jobID1",
		Status:       "Succeed",
		StartTime:    now,
		EndTime:      et1,
	}
	task2 := &models.Task{
		ID:           20192,
		ExecutionID:  executionID,
		ResourceType: "resourceType2",
		SrcResource:  "srcResource2",
		DstResource:  "dstResource2",
		JobID:        "jobID2",
		Status:       "Stopped",
		StartTime:    now,
		EndTime:      et2,
	}
	AddTask(task1)
	AddTask(task2)

	defer func() {
		DeleteAllTasks(executionID)
		DeleteAllExecutions(11209)
	}()

	// query and fill
	exe, err := GetExecution(executionID)
	require.Nil(t, err)
	assert.Equal(t, "Stopped", exe.Status)
	assert.Equal(t, 0, exe.InProgress)
	assert.Equal(t, 1, exe.Stopped)
	assert.Equal(t, 0, exe.Failed)
	assert.Equal(t, 1, exe.Succeed)
	assert.Equal(t, et2.Second(), exe.EndTime.Second())
}

func TestExecutionFill2(t *testing.T) {
	now := time.Now()
	execution := &models.Execution{
		PolicyID:   11209,
		Status:     "InProgress",
		StatusText: "None",
		Total:      2,
		Trigger:    "Event",
		StartTime:  time.Now(),
	}
	executionID, _ := AddExecution(execution)
	task1 := &models.Task{
		ID:             20191,
		ExecutionID:    executionID,
		ResourceType:   "resourceType1",
		SrcResource:    "srcResource1",
		DstResource:    "dstResource1",
		JobID:          "jobID1",
		Status:         models.TaskStatusInProgress,
		StatusRevision: 1,
		StartTime:      now,
	}
	task2 := &models.Task{
		ID:             20192,
		ExecutionID:    executionID,
		ResourceType:   "resourceType2",
		SrcResource:    "srcResource2",
		DstResource:    "dstResource2",
		JobID:          "jobID2",
		Status:         "Stopped",
		StatusRevision: 1,
		StartTime:      now,
		EndTime:        now,
	}
	taskID1, _ := AddTask(task1)
	AddTask(task2)

	defer func() {
		DeleteAllTasks(executionID)
		DeleteAllExecutions(11209)
	}()

	// query and fill
	exe, err := GetExecution(executionID)
	require.Nil(t, err)
	assert.Equal(t, models.ExecutionStatusInProgress, exe.Status)
	assert.Equal(t, 1, exe.InProgress)
	assert.Equal(t, 1, exe.Stopped)
	assert.Equal(t, 0, exe.Failed)
	assert.Equal(t, 0, exe.Succeed)

	// update task status and query and fill
	UpdateTaskStatus(taskID1, models.TaskStatusFailed, 2, models.TaskStatusInProgress)
	exes, err := GetExecutions(&models.ExecutionQuery{
		PolicyID: 11209,
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(exes))
	assert.Equal(t, models.ExecutionStatusFailed, exes[0].Status)
	assert.Equal(t, 0, exes[0].InProgress)
	assert.Equal(t, 1, exes[0].Stopped)
	assert.Equal(t, 1, exes[0].Failed)
	assert.Equal(t, 0, exes[0].Succeed)
}
