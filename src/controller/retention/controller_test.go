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

package retention

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
)

type ControllerTestSuite struct {
	suite.Suite

	oldClient dep.Client
}

// SetupSuite ...
func (s *ControllerTestSuite) SetupSuite() {

}

func TestMain(m *testing.M) {
	dao.PrepareTestForPostgresSQL()
	os.Exit(m.Run())
}

// TestController ...
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (s *ControllerTestSuite) TestPolicy() {
	projectMgr := &project.Manager{}
	repositoryMgr := &repository.Manager{}
	retentionScheduler := &fakeRetentionScheduler{}
	retentionLauncher := &fakeLauncher{}
	execMgr := &testingTask.ExecutionManager{}
	taskMgr := &testingTask.Manager{}
	retentionMgr := retention.NewManager()
	execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	execMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{
		ID:     1,
		Status: job.RunningStatus.String(),
		ExtraAttrs: map[string]any{
			"dry_run": true,
		},
	}, nil)
	execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{{
		ID:     1,
		Status: job.RunningStatus.String(),
		ExtraAttrs: map[string]any{
			"dry_run": true,
		},
	}}, nil)
	taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{{
		ID:     1,
		Status: job.RunningStatus.String(),
		ExtraAttrs: map[string]any{
			"total":    1,
			"retained": 1,
		},
	}}, nil)
	taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)

	c := defaultController{
		manager:        retentionMgr,
		execMgr:        execMgr,
		taskMgr:        taskMgr,
		launcher:       retentionLauncher,
		projectManager: projectMgr,
		repositoryMgr:  repositoryMgr,
		scheduler:      retentionScheduler,
	}

	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
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
			{
				ID:       2,
				Priority: 1,
				Template: "latestPushedK",
				Disabled: true,
				Parameters: rule.Parameters{
					"latestPushedK": 3,
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
			Settings: map[string]any{
				"cron": "0 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	ctx := orm.Context()
	id, err := c.CreateRetention(ctx, p1)
	s.Require().Nil(err)
	s.Require().True(id > 0)

	p1, err = c.GetRetention(ctx, id)
	s.Require().Nil(err)
	s.Require().EqualValues("project", p1.Scope.Level)
	s.Require().True(p1.ID > 0)

	p1.Scope.Level = "test"
	err = c.UpdateRetention(ctx, p1)
	s.Require().Nil(err)
	p1, err = c.GetRetention(ctx, id)
	s.Require().Nil(err)
	s.Require().EqualValues("test", p1.Scope.Level)

	err = c.DeleteRetention(ctx, id)
	s.Require().Nil(err)

	p1, err = c.GetRetention(ctx, id)
	s.Require().NotNil(err)
	s.Require().True(strings.Contains(err.Error(), "no such Retention policy"))
	s.Require().Nil(p1)
}

func (s *ControllerTestSuite) TestExecution() {
	projectMgr := &project.Manager{}
	repositoryMgr := &repository.Manager{}
	retentionScheduler := &fakeRetentionScheduler{}
	retentionLauncher := &fakeLauncher{}
	execMgr := &testingTask.ExecutionManager{}
	taskMgr := &testingTask.Manager{}
	retentionMgr := retention.NewManager()
	execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	execMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Execution{
		ID:     1,
		Status: job.RunningStatus.String(),
		ExtraAttrs: map[string]any{
			"dry_run": true,
		},
	}, nil)
	execMgr.On("MarkDone", mock.Anything, mock.Anything, mock.Anything).Return(nil).Once()
	execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{{
		ID:     1,
		Status: job.RunningStatus.String(),
		ExtraAttrs: map[string]any{
			"dry_run": true,
		},
	}}, nil)
	taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Task{{
		ID:     1,
		Status: job.RunningStatus.String(),
		ExtraAttrs: map[string]any{
			"total":    1,
			"retained": 1,
		},
	}}, nil)
	taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)

	m := defaultController{
		manager:        retentionMgr,
		execMgr:        execMgr,
		taskMgr:        taskMgr,
		launcher:       retentionLauncher,
		projectManager: projectMgr,
		repositoryMgr:  repositoryMgr,
		scheduler:      retentionScheduler,
		wp:             lib.NewWorkerPool(10),
	}

	p1 := &policy.Metadata{
		Algorithm: "or",
		Rules: []rule.Metadata{
			{
				ID:       1,
				Priority: 1,
				Template: "latestPushedK",
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
			Settings: map[string]any{
				"cron": "0 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	ctx := orm.Context()
	policyID, err := m.CreateRetention(ctx, p1)
	s.Require().Nil(err)
	s.Require().True(policyID > 0)

	id, err := m.TriggerRetentionExec(ctx, policyID, retention.ExecutionTriggerManual, false)
	s.Require().Nil(err)
	s.Require().True(id > 0)

	e1, err := m.GetRetentionExec(ctx, id)
	s.Require().Nil(err)
	s.Require().NotNil(e1)
	s.Require().EqualValues(id, e1.ID)

	err = m.OperateRetentionExec(ctx, id, "stop")
	s.Require().Nil(err)

	es, err := m.ListRetentionExecs(ctx, policyID, nil)
	s.Require().Nil(err)
	s.Require().EqualValues(1, len(es))

	ts, err := m.ListRetentionExecTasks(nil, id, nil)
	s.Require().Nil(err)
	s.Require().EqualValues(1, len(ts))

}

type fakeRetentionScheduler struct {
}

func (f *fakeRetentionScheduler) CountSchedules(ctx context.Context, query *q.Query) (int64, error) {
	panic("implement me")
}

func (f *fakeRetentionScheduler) Schedule(ctx context.Context, vendorType string, vendorID int64, cronType string, cron string, callbackFuncName string, params any, extras map[string]any) (int64, error) {
	return 111, nil
}

func (f *fakeRetentionScheduler) UnScheduleByID(ctx context.Context, id int64) error {
	return nil
}
func (f *fakeRetentionScheduler) UnScheduleByVendor(ctx context.Context, vendorType string, vendorID int64) error {
	return nil
}

func (f *fakeRetentionScheduler) GetSchedule(ctx context.Context, id int64) (*scheduler.Schedule, error) {
	return nil, nil
}
func (f *fakeRetentionScheduler) ListSchedules(ctx context.Context, q *q.Query) ([]*scheduler.Schedule, error) {
	return nil, nil
}

type fakeLauncher struct {
}

func (f *fakeLauncher) Stop(ctx context.Context, executionID int64) error {
	return nil
}

func (f *fakeLauncher) Launch(ctx context.Context, policy *policy.Metadata, executionID int64, isDryRun bool) (int64, error) {
	return 0, nil
}

func (s *ControllerTestSuite) TestConvertExecution() {
	// Case 1: ExtraAttrs is nil
	e1 := &task.Execution{
		ID:         1,
		VendorID:   10,
		Status:     "Success",
		VendorType: "RETENTION",
		ExtraAttrs: nil,
	}
	r1 := convertExecution(e1)
	s.Equal(int64(1), r1.ID)
	s.Equal(int64(10), r1.PolicyID)
	s.Equal("Success", r1.Status)
	s.False(r1.DryRun)
	s.Empty(r1.Operator)

	// Case 2: ExtraAttrs is empty
	e2 := &task.Execution{
		ID:         2,
		VendorID:   10,
		Status:     "Success",
		VendorType: "RETENTION",
		ExtraAttrs: map[string]any{},
	}
	r2 := convertExecution(e2)
	s.Equal(int64(2), r2.ID)
	s.False(r2.DryRun)
	s.Empty(r2.Operator)

	// Case 3: ExtraAttrs has invalid types
	e3 := &task.Execution{
		ID:         3,
		VendorID:   10,
		Status:     "Success",
		VendorType: "RETENTION",
		ExtraAttrs: map[string]any{
			"dry_run":  "not-a-bool",
			"operator": 12345,
		},
	}
	r3 := convertExecution(e3)
	s.Equal(int64(3), r3.ID)
	s.False(r3.DryRun)
	s.Empty(r3.Operator)

	// Case 4: ExtraAttrs has valid types
	e4 := &task.Execution{
		ID:         4,
		VendorID:   10,
		Status:     "Success",
		VendorType: "RETENTION",
		ExtraAttrs: map[string]any{
			"dry_run":  true,
			"operator": "admin",
		},
	}
	r4 := convertExecution(e4)
	s.Equal(int64(4), r4.ID)
	s.True(r4.DryRun)
	s.Equal("admin", r4.Operator)
}

