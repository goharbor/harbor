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

package replication

import (
	"context"
	"strconv"

	"github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	pkgmodel "github.com/goharbor/harbor/src/pkg/replication/model"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const callbackFuncName = "REPLICATION_CALLBACK"

func init() {
	callbackFunc := func(ctx context.Context, param string) error {
		policyID, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return err
		}
		policy, err := Ctl.GetPolicy(ctx, policyID)
		if err != nil {
			return err
		}
		_, err = Ctl.Start(ctx, policy, nil, task.ExecutionTriggerSchedule)
		return err
	}
	err := scheduler.RegisterCallbackFunc(callbackFuncName, callbackFunc)
	if err != nil {
		log.Errorf("failed to register the callback function for replication: %v", err)
	}
}

func (c *controller) PolicyCount(ctx context.Context, query *q.Query) (int64, error) {
	return c.repMgr.Count(ctx, query)
}

func (c *controller) ListPolicies(ctx context.Context, query *q.Query) ([]*model.Policy, error) {
	policies, err := c.repMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var ps []*model.Policy
	for _, policy := range policies {
		p, err := c.populateRegistry(ctx, policy)
		if err != nil {
			return nil, err
		}
		ps = append(ps, p)
	}
	return ps, nil
}

func (c *controller) populateRegistry(ctx context.Context, p *pkgmodel.Policy) (*model.Policy, error) {
	policy := &model.Policy{}
	policy.From(p)
	var srcRegistryID, destRegistryID int64 = 0, 0
	if policy.SrcRegistry != nil && policy.SrcRegistry.ID != 0 {
		srcRegistryID = policy.SrcRegistry.ID
		destRegistryID = 0
	} else {
		srcRegistryID = 0
		destRegistryID = policy.DestRegistry.ID
	}
	srcRegistry, err := c.regMgr.Get(ctx, srcRegistryID)
	if err != nil {
		return nil, err
	}
	policy.SrcRegistry = srcRegistry

	destRegistry, err := c.regMgr.Get(ctx, destRegistryID)
	if err != nil {
		return nil, err
	}
	policy.DestRegistry = destRegistry
	return policy, nil
}

func (c *controller) GetPolicy(ctx context.Context, id int64) (*model.Policy, error) {
	policy, err := c.repMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	return c.populateRegistry(ctx, policy)
}

func (c *controller) CreatePolicy(ctx context.Context, policy *model.Policy) (int64, error) {
	if err := c.validatePolicy(ctx, policy); err != nil {
		return 0, err
	}

	p, err := policy.To()
	if err != nil {
		return 0, err
	}

	// create policy
	id, err := c.repMgr.Create(ctx, p)
	if err != nil {
		return 0, err
	}
	// create schedule if needed
	if policy.IsScheduledTrigger() {
		if _, err = c.scheduler.Schedule(ctx, job.Replication, id, "", policy.Trigger.Settings.Cron,
			callbackFuncName, id, map[string]interface{}{}); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (c *controller) UpdatePolicy(ctx context.Context, policy *model.Policy, props ...string) error {
	if err := c.validatePolicy(ctx, policy); err != nil {
		return err
	}
	// delete the schedule
	if err := c.scheduler.UnScheduleByVendor(ctx, job.Replication, policy.ID); err != nil {
		return err
	}

	p, err := policy.To()
	if err != nil {
		return err
	}
	// update the policy
	if err := c.repMgr.Update(ctx, p, props...); err != nil {
		return err
	}
	// create schedule if needed
	if policy.IsScheduledTrigger() {
		if _, err := c.scheduler.Schedule(ctx, job.Replication, policy.ID, "", policy.Trigger.Settings.Cron,
			callbackFuncName, policy.ID, map[string]interface{}{}); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) validatePolicy(ctx context.Context, policy *model.Policy) error {
	if err := policy.Validate(); err != nil {
		return err
	}
	if policy.SrcRegistry != nil {
		if _, err := c.regMgr.Get(ctx, policy.SrcRegistry.ID); err != nil {
			return err
		}
	}
	if policy.DestRegistry != nil {
		if _, err := c.regMgr.Get(ctx, policy.DestRegistry.ID); err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) DeletePolicy(ctx context.Context, id int64) error {
	// delete the executions
	if err := c.execMgr.DeleteByVendor(ctx, job.Replication, id); err != nil {
		return err
	}
	// delete the schedule
	if err := c.scheduler.UnScheduleByVendor(ctx, job.Replication, id); err != nil {
		return err
	}
	// delete the policy
	return c.repMgr.Delete(ctx, id)
}
