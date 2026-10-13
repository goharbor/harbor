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
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	localSecurity "github.com/goharbor/harbor/src/common/security/local"
	robotSecurity "github.com/goharbor/harbor/src/common/security/robot"
	projectCtl "github.com/goharbor/harbor/src/controller/project"
	robotCtl "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	projectmodels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	robotmodel "github.com/goharbor/harbor/src/pkg/robot/model"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	projectcontrollertesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	testingMeta "github.com/goharbor/harbor/src/testing/pkg/project/metadata"
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

	ctx := security.NewContext(orm.Context(), localSecurity.NewSecurityContext(&models.User{
		UserID:       1,
		Username:     "admin",
		SysAdminFlag: true,
	}))
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

func (s *ControllerTestSuite) TestDeleteRetentionByProject() {
	const projectID = int64(2)

	projectMetaMgr := &testingMeta.Manager{}
	execMgr := &testingTask.ExecutionManager{}
	execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{}, nil)
	projectMetaMgr.On("Delete", mock.Anything, projectID, "retention_id").Return(nil)

	c := defaultController{
		manager:        retention.NewManager(),
		execMgr:        execMgr,
		taskMgr:        &testingTask.Manager{},
		launcher:       &fakeLauncher{},
		projectManager: &project.Manager{},
		projectMetaMgr: projectMetaMgr,
		repositoryMgr:  &repository.Manager{},
		scheduler:      &fakeRetentionScheduler{},
	}

	ctx := security.NewContext(orm.Context(), localSecurity.NewSecurityContext(&models.User{
		UserID:       1,
		Username:     "admin",
		SysAdminFlag: true,
	}))
	id, err := c.CreateRetention(ctx, &policy.Metadata{
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
						Pattern:    "**",
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
			Reference: projectID,
		},
	})
	s.Require().Nil(err)
	s.Require().True(id > 0)

	s.Require().Nil(c.DeleteRetentionByProject(ctx, projectID))

	p, err := c.GetRetention(ctx, id)
	s.Require().NotNil(err)
	s.Require().Nil(p)

	projectMetaMgr.AssertCalled(s.T(), "Delete", mock.Anything, projectID, "retention_id")
}

func (s *ControllerTestSuite) TestExecutionPrincipalPersistence() {
	execMgr := &testingTask.ExecutionManager{}
	execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{}, nil)
	c := &defaultController{
		manager:   retention.NewManager(),
		execMgr:   execMgr,
		scheduler: &fakeRetentionScheduler{},
	}
	// A policy submitted through the API never carries a principal.
	userPolicy := func(projectID int64) *policy.Metadata {
		p := policy.WithNDaysSinceLastPull(projectID, 7)
		p.ExecutionPrincipal = nil
		return p
	}

	localCtx := security.NewContext(orm.Context(), localSecurity.NewSecurityContext(&models.User{
		UserID: 1, Username: "admin", SysAdminFlag: true,
	}))
	localID, err := c.CreateRetention(localCtx, userPolicy(2))
	s.Require().NoError(err)
	stored := s.requireExecutionPrincipal(c, localCtx, localID, `"type":"local","id":1,"sysadmin":true`)

	updaterCtx := security.NewContext(orm.Context(), localSecurity.NewSecurityContext(&models.User{
		UserID: 2, Username: "maintainer", SysAdminFlag: true,
	}))
	s.Require().NoError(c.UpdateRetention(updaterCtx, stored))
	s.requireExecutionPrincipal(c, updaterCtx, localID, `"type":"local","id":2,"sysadmin":true`)
	s.Require().NoError(c.DeleteRetention(updaterCtx, localID))

	originalProjectCtl := projectCtl.Ctl
	projects := &projectcontrollertesting.Controller{}
	projectCtl.Ctl = projects
	s.T().Cleanup(func() {
		projectCtl.Ctl = originalProjectCtl
	})
	projects.On("Get", mock.Anything, int64(3), mock.Anything).
		Return(&projectmodels.Project{ProjectID: 3, Name: "proxy"}, nil)
	artifactDelete := []*types.Policy{{Resource: rbac.ResourceArtifact, Action: rbac.ActionDelete}}
	robotCtx := security.NewContext(orm.Context(), robotSecurity.NewSecurityContext(&robotCtl.Robot{
		Robot: robotmodel.Robot{ID: 42, Name: "proxy-cleaner", ProjectID: 3},
		Level: robotCtl.LEVELPROJECT,
		Permissions: []*robotCtl.Permission{
			{Kind: "project", Namespace: "proxy", Scope: "/project/3", Access: artifactDelete},
		},
	}))
	robotID, err := c.CreateRetention(robotCtx, userPolicy(3))
	s.Require().NoError(err)
	stored = s.requireExecutionPrincipal(c, robotCtx, robotID, `"type":"robot","id":42`)

	readOnlyRobotCtx := security.NewContext(orm.Context(), robotSecurity.NewSecurityContext(&robotCtl.Robot{
		Robot: robotmodel.Robot{ID: 43, Name: "proxy-reader", ProjectID: 3},
		Level: robotCtl.LEVELPROJECT,
		Permissions: []*robotCtl.Permission{
			{Kind: "project", Namespace: "proxy", Scope: "/project/3", Access: []*types.Policy{
				{Resource: rbac.ResourceArtifact, Action: rbac.ActionRead},
			}},
		},
	}))
	err = c.UpdateRetention(readOnlyRobotCtx, stored)
	s.Require().ErrorContains(err, "requires permission to delete artifacts")
	s.requireExecutionPrincipal(c, robotCtx, robotID, `"type":"robot","id":42`)
	s.Require().NoError(c.DeleteRetention(robotCtx, robotID))

	_, err = c.CreateRetention(orm.Context(), userPolicy(4))
	s.Require().ErrorContains(err, "authenticated principal required")
	unsupportedContext := &securitytesting.Context{}
	unsupportedContext.On("IsAuthenticated").Return(true).Once()
	unsupportedContext.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
	unsupportedContext.On("Name").Return("secret").Once()
	_, err = c.CreateRetention(security.NewContext(orm.Context(), unsupportedContext), userPolicy(4))
	s.Require().ErrorContains(err, `unsupported principal type "secret"`)

	// Group membership and external admin authority are recorded with the schedule.
	externalCtx := security.NewContext(orm.Context(), localSecurity.NewSecurityContext(&models.User{
		UserID:          5,
		Username:        "external-admin",
		AdminRoleInAuth: true,
		GroupIDs:        []int{8, 9},
	}))
	externalID, err := c.CreateRetention(externalCtx, userPolicy(4))
	s.Require().NoError(err)
	s.requireExecutionPrincipal(c, externalCtx, externalID,
		`"type":"local","id":5,"admin_role_in_auth":true,"group_ids":[8,9]`)
	s.Require().NoError(c.DeleteRetention(externalCtx, externalID))

	// Policies Harbor creates itself need no caller and run as the system.
	systemID, err := c.CreateRetention(orm.Context(), policy.WithNDaysSinceLastPull(5, 7))
	s.Require().NoError(err)
	s.requireExecutionPrincipal(c, orm.Context(), systemID, `"type":"system"`)
	s.Require().NoError(c.DeleteRetention(orm.Context(), systemID))

	inactivePolicy := userPolicy(4)
	inactivePolicy.Trigger.Settings[policy.TriggerSettingsCron] = ""
	inactiveID, err := c.CreateRetention(orm.Context(), inactivePolicy)
	s.Require().NoError(err)
	storedInactivePolicy, err := c.GetRetention(orm.Context(), inactiveID)
	s.Require().NoError(err)
	serialized, err := json.Marshal(storedInactivePolicy)
	s.Require().NoError(err)
	s.NotContains(string(serialized), "execution_principal")
	s.Require().NoError(c.DeleteRetention(orm.Context(), inactiveID))
}

func (s *ControllerTestSuite) requireExecutionPrincipal(
	c *defaultController,
	ctx context.Context,
	policyID int64,
	want string,
) *policy.Metadata {
	stored, err := c.GetRetention(ctx, policyID)
	s.Require().NoError(err)
	serialized, err := json.Marshal(stored)
	s.Require().NoError(err)
	s.Contains(string(serialized), want)
	return stored
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

	ctx := security.NewContext(orm.Context(), localSecurity.NewSecurityContext(&models.User{
		UserID:       1,
		Username:     "admin",
		SysAdminFlag: true,
	}))
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
