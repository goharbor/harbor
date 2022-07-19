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
	"context"
	"encoding/json"

	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/task"
)

type copyFlow struct {
	executionID  int64
	resources    []*model.Resource
	policy       *repctlmodel.Policy
	executionMgr task.ExecutionManager
	taskMgr      task.Manager
}

// NewCopyFlow returns an instance of the copy flow which replicates the resources from
// the source registry to the destination registry. If the parameter "resources" isn't provided,
// will fetch the resources first
func NewCopyFlow(executionID int64, policy *repctlmodel.Policy, resources ...*model.Resource) Flow {
	return &copyFlow{
		executionMgr: task.ExecMgr,
		taskMgr:      task.Mgr,
		executionID:  executionID,
		policy:       policy,
		resources:    resources,
	}
}

func (c *copyFlow) Run(ctx context.Context) error {
	logger := log.GetLogger(ctx)
	srcAdapter, dstAdapter, err := initialize(c.policy)
	if err != nil {
		return err
	}
	srcResources := c.resources
	if len(srcResources) == 0 {
		srcResources, err = fetchResources(srcAdapter, c.policy)
		if err != nil {
			return err
		}
	}

	isStopped, err := c.isExecutionStopped(ctx)
	if err != nil {
		return err
	}
	if isStopped {
		logger.Debugf("the execution %d is stopped, stop the flow", c.executionID)
		return nil
	}

	if len(srcResources) == 0 {
		// no candidates, mark the execution as done directly
		if err := c.executionMgr.MarkDone(ctx, c.executionID, "no resources need to be replicated"); err != nil {
			logger.Errorf("failed to mark done for the execution %d: %v", c.executionID, err)
		}
		return nil
	}

	srcResources = assembleSourceResources(srcResources, c.policy)
	info, err := dstAdapter.Info()
	if err != nil {
		return err
	}
	dstResources, err := assembleDestinationResources(srcResources, c.policy, info.SupportedRepositoryPathComponentType)
	if err != nil {
		return err
	}

	if err = prepareForPush(dstAdapter, dstResources); err != nil {
		return err
	}

	return c.createTasks(ctx, srcResources, dstResources, c.policy.Speed)
}

func (c *copyFlow) isExecutionStopped(ctx context.Context) (bool, error) {
	execution, err := c.executionMgr.Get(ctx, c.executionID)
	if err != nil {
		return false, err
	}
	return execution.Status == job.StoppedStatus.String(), nil
}

func (c *copyFlow) createTasks(ctx context.Context, srcResources, dstResources []*model.Resource, speed int32) error {
	var taskCnt int
	defer func() {
		// if no task be created, mark execution done.
		if taskCnt == 0 {
			if err := c.executionMgr.MarkDone(ctx, c.executionID, "no resources need to be replicated"); err != nil {
				logger.Errorf("failed to mark done for the execution %d: %v", c.executionID, err)
			}
		}
	}()

	for i, srcResource := range srcResources {
		dstResource := dstResources[i]
		// if dest resource should be skipped, ignore replicate.
		if dstResource.Skip {
			log.Warningf("skip create replication task because of dest limitation, src: %s, dst: %s", srcResource.Metadata, dstResource.Metadata)
			continue
		}

		src, err := json.Marshal(srcResource)
		if err != nil {
			return err
		}
		dest, err := json.Marshal(dstResource)
		if err != nil {
			return err
		}

		job := &task.Job{
			Name: job.Replication,
			Metadata: &job.Metadata{
				JobKind: job.KindGeneric,
			},
			Parameters: map[string]interface{}{
				"src_resource": string(src),
				"dst_resource": string(dest),
				"speed":        speed,
			},
		}

		if _, err = c.taskMgr.Create(ctx, c.executionID, job, map[string]interface{}{
			"operation":            "copy",
			"resource_type":        string(srcResource.Type),
			"source_resource":      getResourceName(srcResource),
			"destination_resource": getResourceName(dstResource)}); err != nil {
			return err
		}

		taskCnt++
	}
	return nil
}
