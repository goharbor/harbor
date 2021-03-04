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

package controller

import (
	"fmt"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/policy"
	"github.com/goharbor/harbor/src/replication/policy/manager"
)

// const definitions
const (
	CallbackFuncName = "REPLICATION_CALLBACK"
)

// NewController returns a policy controller which can CURD and schedule policies
func NewController() policy.Controller {
	mgr := manager.NewDefaultManager()
	ctl := &controller{
		scheduler: scheduler.Sched,
	}
	ctl.Controller = mgr
	return ctl
}

type controller struct {
	policy.Controller
	scheduler scheduler.Scheduler
}

func (c *controller) Create(policy *model.Policy) (int64, error) {
	id, err := c.Controller.Create(policy)
	if err != nil {
		return 0, err
	}
	if isScheduledTrigger(policy) {
		extras := make(map[string]interface{})
		if _, err = c.scheduler.Schedule(orm.Context(), job.Replication, id, "", policy.Trigger.Settings.Cron, CallbackFuncName, id, extras); err != nil {
			log.Errorf("failed to schedule the policy %d: %v", id, err)
		}
	}
	return id, nil
}

func (c *controller) Update(policy *model.Policy) error {
	origin, err := c.Controller.Get(policy.ID)
	if err != nil {
		return err
	}
	if origin == nil {
		return fmt.Errorf("policy %d not found", policy.ID)
	}
	// if no need to reschedule the policy, just update it
	if !isScheduleTriggerChanged(origin, policy) {
		return c.Controller.Update(policy)
	}
	// need to reschedule the policy
	// unschedule first if needed
	if isScheduledTrigger(origin) {
		if err = c.scheduler.UnScheduleByVendor(orm.Context(), job.Replication, origin.ID); err != nil {
			return fmt.Errorf("failed to unschedule the policy %d: %v", origin.ID, err)
		}
	}
	// update the policy
	if err = c.Controller.Update(policy); err != nil {
		return err
	}
	// schedule again if needed
	if isScheduledTrigger(policy) {
		extras := make(map[string]interface{})
		if _, err = c.scheduler.Schedule(orm.Context(), job.Replication, policy.ID, "", policy.Trigger.Settings.Cron, CallbackFuncName, policy.ID, extras); err != nil {
			return fmt.Errorf("failed to schedule the policy %d: %v", policy.ID, err)
		}
	}
	return nil
}

func (c *controller) Remove(policyID int64) error {
	policy, err := c.Controller.Get(policyID)
	if err != nil {
		return err
	}
	if policy == nil {
		return fmt.Errorf("policy %d not found", policyID)
	}
	if isScheduledTrigger(policy) {
		if err = c.scheduler.UnScheduleByVendor(orm.Context(), job.Replication, policyID); err != nil {
			return err
		}
	}
	return c.Controller.Remove(policyID)
}

func isScheduledTrigger(policy *model.Policy) bool {
	if policy == nil {
		return false
	}
	if !policy.Enabled {
		return false
	}
	if policy.Trigger == nil {
		return false
	}
	return policy.Trigger.Type == model.TriggerTypeScheduled
}

func isScheduleTriggerChanged(origin, current *model.Policy) bool {
	o := isScheduledTrigger(origin)
	c := isScheduledTrigger(current)
	// both triggers are not scheduled
	if !o && !c {
		return false
	}
	// both triggers are scheduled
	if o && c {
		return origin.Trigger.Settings.Cron != current.Trigger.Settings.Cron
	}
	// one is scheduled but the other one isn't
	return true
}
