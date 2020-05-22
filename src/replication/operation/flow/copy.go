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
	"time"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/operation/execution"
	"github.com/goharbor/harbor/src/replication/operation/scheduler"
)

type copyFlow struct {
	executionID  int64
	resources    []*model.Resource
	policy       *model.Policy
	executionMgr execution.Manager
	scheduler    scheduler.Scheduler
}

// NewCopyFlow returns an instance of the copy flow which replicates the resources from
// the source registry to the destination registry. If the parameter "resources" isn't provided,
// will fetch the resources first
func NewCopyFlow(executionMgr execution.Manager, scheduler scheduler.Scheduler,
	executionID int64, policy *model.Policy, resources ...*model.Resource) Flow {
	return &copyFlow{
		executionMgr: executionMgr,
		scheduler:    scheduler,
		executionID:  executionID,
		policy:       policy,
		resources:    resources,
	}
}

func (c *copyFlow) Run(interface{}) (int, error) {
	srcAdapter, dstAdapter, err := initialize(c.policy)
	if err != nil {
		return 0, err
	}
	var srcResources []*model.Resource
	if len(c.resources) > 0 {
		srcResources, err = filterResources(c.resources, c.policy.Filters)
	} else {
		srcResources, err = fetchResources(srcAdapter, c.policy)
	}
	if err != nil {
		return 0, err
	}

	isStopped, err := isExecutionStopped(c.executionMgr, c.executionID)
	if err != nil {
		return 0, err
	}
	if isStopped {
		log.Debugf("the execution %d is stopped, stop the flow", c.executionID)
		return 0, nil
	}

	if len(srcResources) == 0 {
		markExecutionSuccess(c.executionMgr, c.executionID, "no resources need to be replicated")
		log.Infof("no resources need to be replicated for the execution %d, skip", c.executionID)
		return 0, nil
	}

	srcResources = assembleSourceResources(srcResources, c.policy)
	dstResources := assembleDestinationResources(srcResources, c.policy)

	if err = prepareForPush(dstAdapter, dstResources); err != nil {
		return 0, err
	}
	items, err := preprocess(c.scheduler, srcResources, dstResources)
	if err != nil {
		return 0, err
	}
	if err = createTasks(c.executionMgr, c.executionID, items); err != nil {
		return 0, err
	}

	return schedule(c.scheduler, c.executionMgr, items)
}

// mark the execution as success in database
func markExecutionSuccess(mgr execution.Manager, id int64, message string) {
	err := mgr.Update(
		&models.Execution{
			ID:         id,
			Status:     models.ExecutionStatusSucceed,
			StatusText: message,
			EndTime:    time.Now(),
		}, "Status", "StatusText", "EndTime")
	if err != nil {
		log.Errorf("failed to update the execution %d: %v", id, err)
		return
	}
}
