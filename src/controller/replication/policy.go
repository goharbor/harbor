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
	"encoding/json"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	regmodel "github.com/goharbor/harbor/src/pkg/reg/model"
	pkgmodel "github.com/goharbor/harbor/src/pkg/replication/model"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const callbackFuncName = "REPLICATION_CALLBACK"

func init() {
	callbackFunc := func(ctx context.Context, param string) error {
		params := make(map[string]any)
		if err := json.Unmarshal([]byte(param), &params); err != nil {
			return err
		}

		var policyID int64
		if id, ok := params["policy_id"].(float64); ok {
			policyID = int64(id)
		}

		if op, ok := params["operator"].(string); ok {
			ctx = context.WithValue(ctx, operator.ContextKey{}, op)
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
	err := policy.From(p)
	if err != nil {
		return nil, err
	}
	var srcRegistryID, destRegistryID int64
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
		cbParams := map[string]any{
			"policy_id": id,
			// the operator of schedule job is harbor-jobservice
			"operator": secret.JobserviceUser,
		}
		if _, err = c.scheduler.Schedule(ctx, job.ReplicationVendorType, id, "", policy.Trigger.Settings.Cron,
			callbackFuncName, cbParams, map[string]any{}); err != nil {
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
	if err := c.scheduler.UnScheduleByVendor(ctx, job.ReplicationVendorType, policy.ID); err != nil {
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
		cbParams := map[string]any{
			"policy_id": policy.ID,
			// the operator of schedule job is harbor-jobservice
			"operator": secret.JobserviceUser,
		}
		if _, err := c.scheduler.Schedule(ctx, job.ReplicationVendorType, policy.ID, "", policy.Trigger.Settings.Cron,
			callbackFuncName, cbParams, map[string]any{}); err != nil {
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
		src, err := c.regMgr.Get(ctx, policy.SrcRegistry.ID)
		if err != nil {
			return err
		}
		if err := validateSourceOptions(policy, src); err != nil {
			return err
		}
	} else if opts, err := policy.SourceOptions(); err != nil {
		return err
	} else if !opts.IsEmpty() {
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("source options need a remote source registry")
	}
	if policy.DestRegistry != nil {
		dest, err := c.regMgr.Get(ctx, policy.DestRegistry.ID)
		if err != nil {
			return err
		}
		if dest != nil && dest.Type == regmodel.RegistryTypeHuggingFace {
			return errors.New(nil).WithCode(errors.BadRequestCode).
				WithMessagef("the %s registry %s can only be a replication source", dest.Type, dest.Name)
		}
	}
	return nil
}

// registryInfo returns the capabilities of a registry without a network call for adapters
// whose factory and Info need none.
func registryInfo(registry *regmodel.Registry) (*regmodel.RegistryInfo, error) {
	factory, err := adapter.GetFactory(registry.Type)
	if err != nil {
		return nil, err
	}
	a, err := factory.Create(registry)
	if err != nil {
		return nil, err
	}
	return a.Info()
}

func supportsFilter(info *regmodel.RegistryInfo, filterType string) bool {
	for _, f := range info.SupportedResourceFilters {
		if f != nil && f.Type == filterType {
			return true
		}
	}
	return false
}

// validateSourceOptions rejects source options the source registry cannot honor.
func validateSourceOptions(policy *model.Policy, src *regmodel.Registry) error {
	opts, err := policy.SourceOptions()
	if err != nil {
		return err
	}
	if opts.IsEmpty() {
		return nil
	}
	info, err := registryInfo(src)
	if err != nil {
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("failed to get the capabilities of the source registry %s: %v", src.Name, err)
	}
	if !opts.FileFilter.IsEmpty() && !supportsFilter(info, regmodel.FilterTypeFile) {
		return errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessagef("the source registry %s does not support the %s filter", src.Name, regmodel.FilterTypeFile)
	}
	return nil
}

func (c *controller) DeletePolicy(ctx context.Context, id int64) error {
	// delete the executions
	if err := c.execMgr.DeleteByVendor(ctx, job.ReplicationVendorType, id); err != nil {
		return err
	}
	// delete the schedule
	if err := c.scheduler.UnScheduleByVendor(ctx, job.ReplicationVendorType, id); err != nil {
		return err
	}
	// delete the policy
	return c.repMgr.Delete(ctx, id)
}
