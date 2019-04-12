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

package flow

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/operation/execution"
	"github.com/goharbor/harbor/src/replication/operation/scheduler"
)

type deletionFlow struct {
	executionID  int64
	policy       *model.Policy
	executionMgr execution.Manager
	scheduler    scheduler.Scheduler
	resources    []*model.Resource
}

// NewDeletionFlow returns an instance of the delete flow which deletes the resources
// on the destination registry
func NewDeletionFlow(executionMgr execution.Manager, scheduler scheduler.Scheduler,
	executionID int64, policy *model.Policy, resources []*model.Resource) Flow {
	return &deletionFlow{
		executionMgr: executionMgr,
		scheduler:    scheduler,
		executionID:  executionID,
		policy:       policy,
		resources:    resources,
	}
}

func (d *deletionFlow) Run(interface{}) (int, error) {
	_, dstAdapter, err := initialize(d.policy)
	if err != nil {
		return 0, err
	}
	// filling the registry information
	for _, resource := range d.resources {
		resource.Registry = d.policy.SrcRegistry
	}
	srcResources, err := filterResources(d.resources, d.policy.Filters)
	if err != nil {
		return 0, err
	}
	if len(srcResources) == 0 {
		markExecutionSuccess(d.executionMgr, d.executionID, "no resources need to be replicated")
		log.Infof("no resources need to be replicated for the execution %d, skip", d.executionID)
		return 0, nil
	}
	dstResources, err := assembleDestinationResources(dstAdapter, srcResources, d.policy)
	if err != nil {
		return 0, err
	}

	items, err := preprocess(d.scheduler, srcResources, dstResources)
	if err != nil {
		return 0, err
	}
	if err = createTasks(d.executionMgr, d.executionID, items); err != nil {
		return 0, err
	}

	return schedule(d.scheduler, d.executionMgr, items)
}
