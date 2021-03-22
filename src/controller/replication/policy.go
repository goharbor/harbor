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

	commonthttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/replication"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/model"
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

func (c *controller) ListPolicies(ctx context.Context, query *q.Query) ([]*replication.Policy, error) {
	policies, err := c.repMgr.List(ctx, query)
	if err != nil {
		return nil, err
	}
	for _, policy := range policies {
		if err := c.populateRegistry(ctx, policy); err != nil {
			return nil, err
		}
	}
	return policies, nil
}

func (c *controller) populateRegistry(ctx context.Context, policy *replication.Policy) error {
	if policy.SrcRegistry != nil && policy.SrcRegistry.ID > 0 {
		registry, err := c.regMgr.Get(ctx, policy.SrcRegistry.ID)
		if err != nil {
			return err
		}
		policy.SrcRegistry = registry
		policy.DestRegistry = GetLocalRegistry()
		return nil
	}

	registry, err := c.regMgr.Get(ctx, policy.DestRegistry.ID)
	if err != nil {
		return err
	}
	policy.DestRegistry = registry
	policy.SrcRegistry = GetLocalRegistry()
	return nil
}

func (c *controller) GetPolicy(ctx context.Context, id int64) (*replication.Policy, error) {
	policy, err := c.repMgr.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if err = c.populateRegistry(ctx, policy); err != nil {
		return nil, err
	}
	return policy, nil
}

func (c *controller) CreatePolicy(ctx context.Context, policy *replication.Policy) (int64, error) {
	if err := c.validatePolicy(ctx, policy); err != nil {
		return 0, err
	}
	// create policy
	id, err := c.repMgr.Create(ctx, policy)
	if err != nil {
		return 0, err
	}
	// create schedule if needed
	if policy.IsScheduledTrigger() {
		if _, err = c.scheduler.Schedule(ctx, job.Replication, id, "", policy.Trigger.Settings.Cron,
			callbackFuncName, policy.ID, map[string]interface{}{}); err != nil {
			return 0, err
		}
	}
	return id, nil
}

func (c *controller) UpdatePolicy(ctx context.Context, policy *replication.Policy, props ...string) error {
	if err := c.validatePolicy(ctx, policy); err != nil {
		return err
	}
	// delete the schedule
	if err := c.scheduler.UnScheduleByVendor(ctx, job.Replication, policy.ID); err != nil {
		return err
	}
	// update the policy
	if err := c.repMgr.Update(ctx, policy); err != nil {
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

func (c *controller) validatePolicy(ctx context.Context, policy *replication.Policy) error {
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

// GetLocalRegistry returns the info of the local Harbor registry
// TODO move it into the registry package
func GetLocalRegistry() *model.Registry {
	return &model.Registry{
		Type:            model.RegistryTypeHarbor,
		Name:            "Local",
		URL:             config.Config.CoreURL,
		TokenServiceURL: config.Config.TokenServiceURL,
		Status:          "healthy",
		Credential: &model.Credential{
			Type: model.CredentialTypeSecret,
			// use secret to do the auth for the local Harbor
			AccessSecret: config.Config.JobserviceSecret,
		},
		Insecure: !commonthttp.InternalTLSEnabled(),
	}
}
