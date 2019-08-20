package retention

import (
	"os"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	jjob "github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	tjob "github.com/goharbor/harbor/src/testing/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestPolicy(t *testing.T) {
	m := NewManager()
	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "recentXdays",
				Parameters: rule.Parameters{
					"num": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "label",
						Decoration: "with",
						Pattern:    "latest",
					},
					{
						Kind:       "regularExpression",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "regularExpression",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
		},
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	id, err := m.CreatePolicy(p1)
	assert.Nil(t, err)
	assert.True(t, id > 0)

	p1, err = m.GetPolicy(id)
	assert.Nil(t, err)
	assert.EqualValues(t, "project", p1.Scope.Level)
	assert.True(t, p1.ID > 0)

	p1.Scope.Level = "test"
	err = m.UpdatePolicy(p1)
	assert.Nil(t, err)
	p1, err = m.GetPolicy(id)
	assert.Nil(t, err)
	assert.EqualValues(t, "test", p1.Scope.Level)

	err = m.DeletePolicyAndExec(id)
	assert.Nil(t, err)

	p1, err = m.GetPolicy(id)
	assert.Nil(t, err)
	assert.Nil(t, p1)
}

func TestExecution(t *testing.T) {
	m := NewManager()
	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "recentXdays",
				Parameters: rule.Parameters{
					"num": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "label",
						Decoration: "with",
						Pattern:    "latest",
					},
					{
						Kind:       "regularExpression",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "regularExpression",
							Decoration: "matches",
							Pattern:    ".+",
						},
					},
				},
			},
		},
		Trigger: &policy.Trigger{
			Kind: "Schedule",
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	policyID, err := m.CreatePolicy(p1)
	assert.Nil(t, err)
	assert.True(t, policyID > 0)

	e1 := &Execution{
		PolicyID:  policyID,
		StartTime: time.Now(),
		Trigger:   ExecutionTriggerManual,
		DryRun:    false,
	}
	id, err := m.CreateExecution(e1)
	assert.Nil(t, err)
	assert.True(t, id > 0)

	e1, err = m.GetExecution(id)
	assert.Nil(t, err)
	assert.NotNil(t, e1)
	assert.EqualValues(t, id, e1.ID)

	es, err := m.ListExecutions(policyID, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(es))

	err = m.DeleteExecution(id)
	assert.Nil(t, err)
}

func TestTask(t *testing.T) {
	m := NewManager()
	task := &Task{
		ExecutionID: 1,
		JobID:       "1",
		Status:      jjob.PendingStatus.String(),
		StatusCode:  jjob.PendingStatus.Code(),
		Total:       0,
		StartTime:   time.Now(),
	}
	// create
	id, err := m.CreateTask(task)
	require.Nil(t, err)

	// get
	tk, err := m.GetTask(id)
	require.Nil(t, err)
	assert.EqualValues(t, 1, tk.ExecutionID)

	// update
	task.ID = id
	task.Total = 1
	err = m.UpdateTask(task, "Total")
	require.Nil(t, err)

	// update status to success which is a final status
	err = m.UpdateTaskStatus(id, jjob.SuccessStatus.String())
	require.Nil(t, err)

	// try to update status to running, as the status has already
	// been updated to a final status, this updating shouldn't take effect
	err = m.UpdateTaskStatus(id, jjob.RunningStatus.String())
	require.Nil(t, err)

	// list
	tasks, err := m.ListTasks(&q.TaskQuery{
		ExecutionID: 1,
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(tasks))
	assert.Equal(t, int64(1), tasks[0].ExecutionID)
	assert.Equal(t, 1, tasks[0].Total)
	assert.Equal(t, jjob.SuccessStatus.String(), tasks[0].Status)
	assert.Equal(t, jjob.SuccessStatus.Code(), tasks[0].StatusCode)

	// get task log
	job.GlobalClient = &tjob.MockJobClient{
		JobUUID: []string{"1"},
	}
	data, err := m.GetTaskLog(task.ID)
	require.Nil(t, err)
	assert.Equal(t, "some log", string(data))
}
