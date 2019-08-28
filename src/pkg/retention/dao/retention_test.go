package dao

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/retention/dao/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/retention/q"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

func TestPolicy(t *testing.T) {
	p := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Action:   "retain",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
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
	p1 := &models.RetentionPolicy{
		ScopeLevel:  p.Scope.Level,
		TriggerKind: p.Trigger.Kind,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	data, _ := json.Marshal(p)
	p1.Data = string(data)

	id, err := CreatePolicy(p1)
	assert.Nil(t, err)
	assert.True(t, id > 0)

	p1, err = GetPolicy(id)
	assert.Nil(t, err)
	assert.EqualValues(t, "project", p1.ScopeLevel)
	assert.True(t, p1.ID > 0)

	p1.ScopeLevel = "test"
	err = UpdatePolicy(p1)
	assert.Nil(t, err)
	p1, err = GetPolicy(id)
	assert.Nil(t, err)
	assert.EqualValues(t, "test", p1.ScopeLevel)

	err = DeletePolicyAndExec(id)
	assert.Nil(t, err)

	p1, err = GetPolicy(id)
	assert.NotNil(t, err)
	assert.True(t, strings.Contains(err.Error(), "no row found"))
}

func TestExecution(t *testing.T) {
	p := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
				Action:   "retain",
				Parameters: rule.Parameters{
					"latestPushedK": 10,
				},
				TagSelectors: []*rule.Selector{
					{
						Kind:       "doublestar",
						Decoration: "matches",
						Pattern:    "release-[\\d\\.]+",
					},
				},
				ScopeSelectors: map[string][]*rule.Selector{
					"repository": {
						{
							Kind:       "doublestar",
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
	p1 := &models.RetentionPolicy{
		ScopeLevel:  p.Scope.Level,
		TriggerKind: p.Trigger.Kind,
		CreateTime:  time.Now(),
		UpdateTime:  time.Now(),
	}
	data, _ := json.Marshal(p)
	p1.Data = string(data)

	policyID, err := CreatePolicy(p1)
	assert.Nil(t, err)
	assert.True(t, policyID > 0)

	e := &models.RetentionExecution{
		PolicyID:  policyID,
		DryRun:    false,
		Trigger:   "manual",
		StartTime: time.Now(),
	}
	id, err := CreateExecution(e)
	assert.Nil(t, err)
	assert.True(t, id > 0)

	e1, err := GetExecution(id)
	assert.Nil(t, err)
	assert.NotNil(t, e1)
	assert.EqualValues(t, id, e1.ID)

	es, err := ListExecutions(policyID, nil)
	assert.Nil(t, err)
	assert.EqualValues(t, 1, len(es))
}

func TestTask(t *testing.T) {
	task := &models.RetentionTask{
		ExecutionID: 1,
		Status:      "pending",
	}
	// create
	id, err := CreateTask(task)
	require.Nil(t, err)

	// get
	tk, err := GetTask(id)
	require.Nil(t, err)
	require.Equal(t, id, tk.ID)
	require.Equal(t, "pending", tk.Status)

	// update
	task.ID = id
	task.Total = 1
	err = UpdateTask(task, "Total")
	require.Nil(t, err)

	// update status
	err = UpdateTaskStatus(id, "running", 1, 1)
	require.Nil(t, err)

	// list
	tasks, err := ListTask(&q.TaskQuery{
		ExecutionID: 1,
		Status:      "running",
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(tasks))
	assert.Equal(t, 1, tasks[0].Total)
	assert.Equal(t, int64(1), tasks[0].ExecutionID)
	assert.Equal(t, "running", tasks[0].Status)
	assert.Equal(t, 1, tasks[0].StatusCode)
	assert.Equal(t, int64(1), tasks[0].StatusRevision)

	// delete
	err = DeleteTask(id)
	require.Nil(t, err)
}
