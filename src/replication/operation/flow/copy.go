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

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/operation/execution"
	"github.com/goharbor/harbor/src/replication/operation/scheduler"
)

type copyFlow struct {
	executionID  int64
	policy       *model.Policy
	executionMgr execution.Manager
	scheduler    scheduler.Scheduler
}

// NewCopyFlow returns an instance of the copy flow which replicates the resources from
// the source registry to the destination registry
func NewCopyFlow(executionMgr execution.Manager, scheduler scheduler.Scheduler,
	executionID int64, policy *model.Policy) Flow {
	return &copyFlow{
		executionMgr: executionMgr,
		scheduler:    scheduler,
		executionID:  executionID,
		policy:       policy,
	}
}

func (c *copyFlow) Run(interface{}) (int, error) {
	srcAdapter, dstAdapter, err := initialize(c.policy)
	if err != nil {
		return 0, err
	}
	srcResources, err := fetchResources(srcAdapter, c.policy)
	if err != nil {
		return 0, err
	}
	if len(srcResources) == 0 {
		markExecutionSuccess(c.executionMgr, c.executionID, "no resources need to be replicated")
		log.Infof("no resources need to be replicated for the execution %d, skip", c.executionID)
		return 0, nil
	}
	dstResources, err := assembleDestinationResources(dstAdapter, srcResources, c.policy)
	if err != nil {
		return 0, err
	}

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
