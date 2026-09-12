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
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/pkg/task"
	testingScheduler "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
)

type schedulePolicyManager struct {
	retention.Manager
	mock.Mock
}

func (m *schedulePolicyManager) CreatePolicy(ctx context.Context, p *policy.Metadata) (int64, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(int64), args.Error(1)
}

func (m *schedulePolicyManager) GetPolicy(ctx context.Context, id int64) (*policy.Metadata, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*policy.Metadata), args.Error(1)
}

func (m *schedulePolicyManager) UpdatePolicy(ctx context.Context, p *policy.Metadata) error {
	return m.Called(ctx, p).Error(0)
}

func (m *schedulePolicyManager) DeletePolicy(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func scheduledPolicy(projectID int64, cron string) *policy.Metadata {
	p := policy.WithNDaysSinceLastPull(projectID, 7)
	p.ID = 73
	p.Trigger.Settings[policy.TriggerSettingsCron] = cron
	return p
}

func expectRetentionSchedule(s *testingScheduler.Scheduler, p *policy.Metadata, cron string) *mock.Call {
	return s.On("Schedule", mock.Anything, schedulerVendorType, p.ID, "", cron, SchedulerCallback, TriggerParam{
		PolicyID: p.ID,
		Trigger:  retention.ExecutionTriggerSchedule,
		Operator: secret.JobserviceUser,
	}, map[string]any{}).Return(int64(111), nil).Once()
}

func TestCreateRetentionScheduling(t *testing.T) {
	for _, tt := range []struct {
		name         string
		policy       *policy.Metadata
		wantSchedule bool
	}{
		{name: "unattached project", policy: scheduledPolicy(0, "0 0 0 * * *")},
		{name: "attached project", policy: scheduledPolicy(42, "0 0 0 * * *"), wantSchedule: true},
		{name: "empty cron", policy: scheduledPolicy(42, "")},
		{name: "system scope", policy: func() *policy.Metadata {
			p := scheduledPolicy(0, "0 0 0 * * *")
			p.Scope.Level = "system"
			return p
		}(), wantSchedule: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manager := &schedulePolicyManager{}
			manager.Test(t)
			t.Cleanup(func() { manager.AssertExpectations(t) })
			scheduler := testingScheduler.NewScheduler(t)
			created := manager.On("CreatePolicy", mock.Anything, tt.policy).Return(tt.policy.ID, nil).Once()
			if tt.wantSchedule {
				expectRetentionSchedule(scheduler, tt.policy, "0 0 0 * * *").NotBefore(created)
			}
			controller := &defaultController{manager: manager, scheduler: scheduler}
			id, err := controller.CreateRetention(context.Background(), tt.policy)
			require.NoError(t, err)
			require.Equal(t, tt.policy.ID, id)
		})
	}
}

func TestUpdateRetentionScheduling(t *testing.T) {
	const daily = "0 0 0 * * *"
	const hourly = "0 0 * * * *"
	for _, tt := range []struct {
		name           string
		oldProjectID   int64
		newProjectID   int64
		oldCron        string
		newCron        string
		oldManual      bool
		newManual      bool
		wantUnschedule bool
		wantSchedule   bool
	}{
		{name: "attach unchanged cron", oldProjectID: 0, newProjectID: 42, oldCron: daily, newCron: daily, wantSchedule: true},
		{name: "attach changed cron", oldProjectID: 0, newProjectID: 42, oldCron: daily, newCron: hourly, wantSchedule: true},
		{name: "attach empty cron", oldProjectID: 0, newProjectID: 42},
		{name: "edit unattached cron", oldCron: daily, newCron: hourly},
		{name: "enable unattached cron", newCron: daily},
		{name: "enable unattached trigger", oldManual: true, newCron: daily},
		{name: "detach policy", oldProjectID: 42, oldCron: daily, newCron: daily, wantUnschedule: true},
		{name: "unchanged attached cron", oldProjectID: 42, newProjectID: 42, oldCron: daily, newCron: daily},
		{name: "edit attached cron", oldProjectID: 42, newProjectID: 42, oldCron: daily, newCron: hourly, wantUnschedule: true, wantSchedule: true},
		{name: "disable attached cron", oldProjectID: 42, newProjectID: 42, oldCron: daily, wantUnschedule: true},
		{name: "enable attached cron", oldProjectID: 42, newProjectID: 42, newCron: daily, wantSchedule: true},
		{name: "disable attached trigger", oldProjectID: 42, newProjectID: 42, oldCron: daily, newCron: daily, newManual: true, wantUnschedule: true},
		{name: "enable attached trigger", oldProjectID: 42, newProjectID: 42, oldCron: daily, newCron: daily, oldManual: true, wantSchedule: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			oldPolicy := scheduledPolicy(tt.oldProjectID, tt.oldCron)
			newPolicy := scheduledPolicy(tt.newProjectID, tt.newCron)
			if tt.oldManual {
				oldPolicy.Trigger.Kind = ""
			}
			if tt.newManual {
				newPolicy.Trigger.Kind = ""
			}
			manager := &schedulePolicyManager{}
			manager.Test(t)
			t.Cleanup(func() { manager.AssertExpectations(t) })
			scheduler := testingScheduler.NewScheduler(t)
			manager.On("GetPolicy", mock.Anything, oldPolicy.ID).Return(oldPolicy, nil).Once()
			updated := manager.On("UpdatePolicy", mock.Anything, newPolicy).Return(nil).Once()
			if tt.wantUnschedule {
				updated = scheduler.On("UnScheduleByVendor", mock.Anything, schedulerVendorType, newPolicy.ID).Return(nil).Once().NotBefore(updated)
			}
			if tt.wantSchedule {
				expectRetentionSchedule(scheduler, newPolicy, tt.newCron).NotBefore(updated)
			}
			controller := &defaultController{manager: manager, scheduler: scheduler}
			require.NoError(t, controller.UpdateRetention(context.Background(), newPolicy))
		})
	}
}

func TestDeleteRetentionScheduling(t *testing.T) {
	for _, tt := range []struct {
		name           string
		policy         *policy.Metadata
		wantUnschedule bool
	}{
		{name: "unattached project", policy: scheduledPolicy(0, "0 0 0 * * *")},
		{name: "attached project", policy: scheduledPolicy(42, "0 0 0 * * *"), wantUnschedule: true},
		{name: "empty cron", policy: scheduledPolicy(42, "")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			manager := &schedulePolicyManager{}
			manager.Test(t)
			t.Cleanup(func() { manager.AssertExpectations(t) })
			scheduler := testingScheduler.NewScheduler(t)
			executions := testingTask.NewExecutionManager(t)
			manager.On("GetPolicy", mock.Anything, tt.policy.ID).Return(tt.policy, nil).Once()
			manager.On("DeletePolicy", mock.Anything, tt.policy.ID).Return(nil).Once()
			executions.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{}, nil).Once()
			if tt.wantUnschedule {
				scheduler.On("UnScheduleByVendor", mock.Anything, schedulerVendorType, tt.policy.ID).Return(nil).Once()
			}
			controller := &defaultController{manager: manager, scheduler: scheduler, execMgr: executions}
			require.NoError(t, controller.DeleteRetention(context.Background(), tt.policy.ID))
		})
	}
}
