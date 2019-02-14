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
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/replication/ng/scheduler"

	"github.com/goharbor/harbor/src/replication/ng/execution"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/registry"
)

type flow struct {
	policy       *model.Policy
	srcRegistry  *model.Registry
	dstRegistry  *model.Registry
	srcAdapter   adapter.Adapter
	dstAdapter   adapter.Adapter
	executionID  int64
	resources    []*model.Resource
	executionMgr execution.Manager
	scheduler    scheduler.Scheduler
}

func newFlow(policy *model.Policy, registryMgr registry.Manager,
	executionMgr execution.Manager, scheduler scheduler.Scheduler) (*flow, error) {

	f := &flow{
		policy:       policy,
		executionMgr: executionMgr,
		scheduler:    scheduler,
	}

	// get source registry
	srcRegistry, err := registryMgr.Get(policy.SrcRegistryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get registry %d: %v", policy.SrcRegistryID, err)
	}
	if srcRegistry == nil {
		return nil, fmt.Errorf("registry %d not found", policy.SrcRegistryID)
	}
	f.srcRegistry = srcRegistry

	// get destination registry
	dstRegistry, err := registryMgr.Get(policy.DestRegistryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get registry %d: %v", policy.DestRegistryID, err)
	}
	if dstRegistry == nil {
		return nil, fmt.Errorf("registry %d not found", policy.DestRegistryID)
	}
	f.dstRegistry = dstRegistry

	// create the source registry adapter
	srcFactory, err := adapter.GetFactory(srcRegistry.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter factory for registry type %s: %v", srcRegistry.Type, err)
	}
	srcAdapter, err := srcFactory(srcRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter for source registry %s: %v", srcRegistry.URL, err)
	}
	f.srcAdapter = srcAdapter

	// create the destination registry adapter
	dstFactory, err := adapter.GetFactory(dstRegistry.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter factory for registry type %s: %v", dstRegistry.Type, err)
	}
	dstAdapter, err := dstFactory(dstRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter for destination registry %s: %v", dstRegistry.URL, err)
	}
	f.dstAdapter = dstAdapter

	return f, nil
}

func (f *flow) createExecution() (int64, error) {
	id, err := f.executionMgr.Create(&model.Execution{
		PolicyID:  f.policy.ID,
		Status:    model.ExecutionStatusInProgress,
		StartTime: time.Now(),
	})
	f.executionID = id
	log.Debugf("an execution record for replication based on the policy %d created: %d", f.policy.ID, id)
	return id, err
}

func (f *flow) fetchResources() error {
	resources, err := f.srcAdapter.FetchResources(f.policy.SrcNamespaces, f.policy.Filters)
	f.resources = resources
	if err != nil {
		f.markExecutionFailure(err)
		return err
	}

	log.Debugf("resources for the execution %d fetched from the source registry", f.executionID)
	return nil
}

func (f *flow) createNamespace() error {
	// merge the metadata of all source namespaces
	metadata := map[string]interface{}{}
	for _, srcNamespace := range f.policy.SrcNamespaces {
		namespace, err := f.srcAdapter.GetNamespace(srcNamespace)
		if err != nil {
			f.markExecutionFailure(err)
			return err
		}
		for key, value := range namespace.Metadata {
			metadata[namespace.Name+":"+key] = value
		}
	}

	if err := f.dstAdapter.CreateNamespace(&model.Namespace{
		Name:     f.policy.DestNamespace,
		Metadata: metadata,
	}); err != nil {
		f.markExecutionFailure(err)
		return err
	}

	log.Debugf("namespace %s for the execution %d created on the destination registry", f.policy.DestNamespace, f.executionID)
	return nil
}

func (f *flow) schedule() error {
	dstResources := []*model.Resource{}
	for _, srcResource := range f.resources {
		dstResource := &model.Resource{
			Type: srcResource.Type,
			Metadata: &model.ResourceMetadata{
				Name:      srcResource.Metadata.Name,
				Namespace: f.policy.DestNamespace,
				Vtags:     srcResource.Metadata.Vtags,
			},
			Registry:     f.dstRegistry,
			ExtendedInfo: srcResource.ExtendedInfo,
			Deleted:      srcResource.Deleted,
			Override:     f.policy.Override,
		}
		dstResources = append(dstResources, dstResource)
	}

	tasks, err := f.scheduler.Schedule(f.resources, dstResources)
	if err != nil {
		f.markExecutionFailure(err)
		return err
	}

	allFailed := true
	for _, task := range tasks {
		if task.Status != model.TaskStatusFailed {
			allFailed = false
		}
		task.ExecutionID = f.executionID
		taskID, err := f.executionMgr.CreateTask(task)
		if err != nil {
			f.markExecutionFailure(err)
			return err
		}
		log.Debugf("task record %d for execution %d created", taskID, f.executionID)
	}
	// if all the tasks are failed, mark the execution failed
	if allFailed {
		f.markExecutionFailure(err)
	}

	return nil
}

func (f *flow) markExecutionFailure(err error) {
	statusText := ""
	if err != nil {
		statusText = err.Error()
	}
	err = f.executionMgr.Update(
		&model.Execution{
			ID:         f.executionID,
			Status:     model.ExecutionStatusFailed,
			StatusText: statusText,
			EndTime:    time.Now(),
		})
	if err != nil {
		log.Errorf("failed to update the execution %d: %v", f.executionID, err)
	}
}
