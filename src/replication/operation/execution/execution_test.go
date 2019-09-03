package execution

import (
	"os"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var executionManager = NewDefaultManager()

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestMethodOfExecutionManager(t *testing.T) {
	execution := &models.Execution{
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

	defer func() {
		executionManager.RemoveAll(execution.PolicyID)
	}()

	// Create
	id, err := executionManager.Create(execution)
	require.Nil(t, err)

	// List
	query := &models.ExecutionQuery{
		Statuses: []string{"InProgress", "Failed"},
		Pagination: models.Pagination{
			Page: 1,
			Size: 10,
		},
	}
	count, executions, err := executionManager.List(query)
	require.Nil(t, err)
	assert.Equal(t, int64(1), count)
	assert.Equal(t, 1, len(executions))

	// Get
	_, err = executionManager.Get(id)
	require.Nil(t, err)

	// Update
	executionNew := &models.Execution{
		ID:         id,
		Status:     "Failed",
		Succeed:    12,
		InProgress: 0,
		EndTime:    time.Now(),
	}
	err = executionManager.Update(executionNew, models.ExecutionPropsName.Status, models.ExecutionPropsName.Succeed, models.ExecutionPropsName.InProgress,
		models.ExecutionPropsName.EndTime)
	require.Nil(t, err)

	// Remove
	require.Nil(t, executionManager.Remove(id))
}

func TestMethodOfTaskManager(t *testing.T) {
	now := time.Now()
	task := &models.Task{
		ExecutionID:    112200,
		ResourceType:   "resourceType1",
		SrcResource:    "srcResource1",
		DstResource:    "dstResource1",
		JobID:          "jobID1",
		Status:         "Initialized",
		StatusRevision: 1,
		StartTime:      &now,
	}

	defer func() {
		executionManager.RemoveAllTasks(task.ExecutionID)
	}()

	// CreateTask
	id, err := executionManager.CreateTask(task)
	require.Nil(t, err)

	// ListTasks
	query := &models.TaskQuery{
		ResourceType: "resourceType1",
		Pagination: models.Pagination{
			Page: 1,
			Size: 10,
		},
	}
	count, tasks, err := executionManager.ListTasks(query)
	require.Nil(t, err)
	assert.Equal(t, 1, len(tasks))
	assert.Equal(t, int64(1), count)

	// GetTask
	_, err = executionManager.GetTask(id)
	require.Nil(t, err)

	// UpdateTask
	taskNew := &models.Task{
		ID:          id,
		SrcResource: "srcResourceChanged",
	}
	err = executionManager.UpdateTask(taskNew, models.TaskPropsName.SrcResource)
	require.Nil(t, err)
	taskUpdate, _ := executionManager.GetTask(id)
	assert.Equal(t, taskNew.SrcResource, taskUpdate.SrcResource)

	// UpdateTaskStatus
	err = executionManager.UpdateTaskStatus(id, models.TaskStatusSucceed, 1, models.TaskStatusInitialized)
	require.Nil(t, err)
	taskUpdate, _ = executionManager.GetTask(id)
	assert.Equal(t, models.TaskStatusSucceed, taskUpdate.Status)

	// Remove
	require.Nil(t, executionManager.RemoveTask(id))

	// RemoveAll
	require.Nil(t, executionManager.RemoveAll(id))
}
