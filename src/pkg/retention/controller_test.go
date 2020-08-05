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
	"strings"
	"testing"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/retention/dep"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	"github.com/goharbor/harbor/src/testing/pkg/repository"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	suite.Suite

	oldClient dep.Client
}

// SetupSuite ...
func (s *ControllerTestSuite) SetupSuite() {

}

// TestController ...
func TestController(t *testing.T) {
	suite.Run(t, new(ControllerTestSuite))
}

func (s *ControllerTestSuite) TestPolicy() {
	projectMgr := &project.Manager{}
	repositoryMgr := &repository.FakeManager{}
	retentionScheduler := &fakeRetentionScheduler{}
	retentionLauncher := &fakeLauncher{}
	retentionMgr := NewManager()
	c := NewAPIController(retentionMgr, projectMgr, repositoryMgr, retentionScheduler, retentionLauncher)

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
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	id, err := c.CreateRetention(p1)
	s.Require().Nil(err)
	s.Require().True(id > 0)

	p1, err = c.GetRetention(id)
	s.Require().Nil(err)
	s.Require().EqualValues("project", p1.Scope.Level)
	s.Require().True(p1.ID > 0)

	p1.Scope.Level = "test"
	err = c.UpdateRetention(p1)
	s.Require().Nil(err)
	p1, err = c.GetRetention(id)
	s.Require().Nil(err)
	s.Require().EqualValues("test", p1.Scope.Level)

	err = c.DeleteRetention(id)
	s.Require().Nil(err)

	p1, err = c.GetRetention(id)
	s.Require().NotNil(err)
	s.Require().True(strings.Contains(err.Error(), "no such Retention policy"))
	s.Require().Nil(p1)
}

func (s *ControllerTestSuite) TestExecution() {
	projectMgr := &project.Manager{}
	repositoryMgr := &repository.FakeManager{}
	retentionScheduler := &fakeRetentionScheduler{}
	retentionLauncher := &fakeLauncher{}
	retentionMgr := NewManager()
	m := NewAPIController(retentionMgr, projectMgr, repositoryMgr, retentionScheduler, retentionLauncher)

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
			Settings: map[string]interface{}{
				"cron": "* 22 11 * * *",
			},
		},
		Scope: &policy.Scope{
			Level:     "project",
			Reference: 1,
		},
	}

	policyID, err := m.CreateRetention(p1)
	s.Require().Nil(err)
	s.Require().True(policyID > 0)

	id, err := m.TriggerRetentionExec(policyID, ExecutionTriggerManual, false)
	s.Require().Nil(err)
	s.Require().True(id > 0)

	e1, err := m.GetRetentionExec(id)
	s.Require().Nil(err)
	s.Require().NotNil(e1)
	s.Require().EqualValues(id, e1.ID)

	err = m.OperateRetentionExec(id, "stop")
	s.Require().Nil(err)

	es, err := m.ListRetentionExecs(policyID, nil)
	s.Require().Nil(err)
	s.Require().EqualValues(1, len(es))

	ts, err := m.ListRetentionExecTasks(id, nil)
	s.Require().Nil(err)
	s.Require().EqualValues(0, len(ts))

}

type fakeRetentionScheduler struct {
}

func (f *fakeRetentionScheduler) Schedule(ctx context.Context, vendorType string, vendorID int64, cron string, callbackFuncName string, params interface{}) (int64, error) {
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

func (f *fakeLauncher) Stop(executionID int64) error {
	return nil
}

func (f *fakeLauncher) Launch(policy *policy.Metadata, executionID int64, isDryRun bool) (int64, error) {
	return 0, nil
}
