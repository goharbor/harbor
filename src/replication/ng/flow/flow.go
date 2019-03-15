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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/replication/ng/scheduler"

	"github.com/goharbor/harbor/src/replication/ng/execution"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/replication/ng/adapter"
	"github.com/goharbor/harbor/src/replication/ng/dao/models"
	"github.com/goharbor/harbor/src/replication/ng/model"
	"github.com/goharbor/harbor/src/replication/ng/registry"
)

type flow struct {
	policy        *model.Policy
	srcRegistry   *model.Registry
	dstRegistry   *model.Registry
	srcAdapter    adapter.Adapter
	dstAdapter    adapter.Adapter
	executionID   int64
	srcResources  []*model.Resource
	dstResources  []*model.Resource
	executionMgr  execution.Manager
	scheduler     scheduler.Scheduler
	scheduleItems []*scheduler.ScheduleItem
}

func newFlow(policy *model.Policy, registryMgr registry.Manager,
	executionMgr execution.Manager, scheduler scheduler.Scheduler) (*flow, error) {

	f := &flow{
		policy:       policy,
		executionMgr: executionMgr,
		scheduler:    scheduler,
	}

	// TODO consider to put registry model in the policy directly rather than just the registry ID?
	url, err := config.RegistryURL()
	if err != nil {
		return nil, fmt.Errorf("failed to get the registry URL: %v", err)
	}
	registry := &model.Registry{
		Type: model.RegistryTypeHarbor,
		Name: "Local",
		URL:  url,
		// TODO use the service account
		Credential: &model.Credential{
			Type:         model.CredentialTypeBasic,
			AccessKey:    "admin",
			AccessSecret: "Harbor12345",
		},
		Insecure: true,
	}

	// get source registry
	if policy.SrcRegistryID != 0 {
		srcRegistry, err := registryMgr.Get(policy.SrcRegistryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get registry %d: %v", policy.SrcRegistryID, err)
		}
		if srcRegistry == nil {
			return nil, fmt.Errorf("registry %d not found", policy.SrcRegistryID)
		}
		f.srcRegistry = srcRegistry
	} else {
		f.srcRegistry = registry
	}

	// get destination registry
	if policy.DestRegistryID != 0 {
		dstRegistry, err := registryMgr.Get(policy.DestRegistryID)
		if err != nil {
			return nil, fmt.Errorf("failed to get registry %d: %v", policy.DestRegistryID, err)
		}
		if dstRegistry == nil {
			return nil, fmt.Errorf("registry %d not found", policy.DestRegistryID)
		}
		f.dstRegistry = dstRegistry
	} else {
		f.dstRegistry = registry
	}

	// create the source registry adapter
	srcFactory, err := adapter.GetFactory(f.srcRegistry.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter factory for registry type %s: %v", f.srcRegistry.Type, err)
	}
	srcAdapter, err := srcFactory(f.srcRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter for source registry %s: %v", f.srcRegistry.URL, err)
	}
	f.srcAdapter = srcAdapter

	// create the destination registry adapter
	dstFactory, err := adapter.GetFactory(f.dstRegistry.Type)
	if err != nil {
		return nil, fmt.Errorf("failed to get adapter factory for registry type %s: %v", f.dstRegistry.Type, err)
	}
	dstAdapter, err := dstFactory(f.dstRegistry)
	if err != nil {
		return nil, fmt.Errorf("failed to create adapter for destination registry %s: %v", f.dstRegistry.URL, err)
	}
	f.dstAdapter = dstAdapter

	return f, nil
}

func (f *flow) createExecution() (int64, error) {
	id, err := f.executionMgr.Create(&models.Execution{
		PolicyID:  f.policy.ID,
		Status:    models.ExecutionStatusInProgress,
		StartTime: time.Now(),
	})
	f.executionID = id
	log.Debugf("an execution record for replication based on the policy %d created: %d", f.policy.ID, id)
	return id, err
}

func (f *flow) fetchResources() error {
	resTypes := []model.ResourceType{}
	filters := []*model.Filter{}
	for _, filter := range f.policy.Filters {
		if filter.Type != model.FilterTypeResource {
			filters = append(filters, filter)
			continue
		}
		resTypes = append(resTypes, filter.Value.(model.ResourceType))
	}
	if len(resTypes) == 0 {
		resTypes = append(resTypes, adapter.GetAdapterInfo(f.srcRegistry.Type).SupportedResourceTypes...)
	}

	// TODO consider whether the logic can be refactored by using reflect
	srcResources := []*model.Resource{}
	for _, typ := range resTypes {
		if typ == model.ResourceTypeRepository {
			reg, ok := f.srcAdapter.(adapter.ImageRegistry)
			if !ok {
				err := fmt.Errorf("the adapter doesn't implement the ImageRegistry interface")
				f.markExecutionFailure(err)
				return err
			}
			res, err := reg.FetchImages(f.policy.SrcNamespaces, filters)
			if err != nil {
				f.markExecutionFailure(err)
				return err
			}
			srcResources = append(srcResources, res...)
			continue
		}
		// TODO add support for chart
	}

	dstResources := []*model.Resource{}
	for _, srcResource := range srcResources {
		dstResource := &model.Resource{
			Type: srcResource.Type,
			Metadata: &model.ResourceMetadata{
				Name:      srcResource.Metadata.Name,
				Namespace: srcResource.Metadata.Namespace,
				Vtags:     srcResource.Metadata.Vtags,
			},
			Registry:     f.dstRegistry,
			ExtendedInfo: srcResource.ExtendedInfo,
			Deleted:      srcResource.Deleted,
			Override:     f.policy.Override,
		}
		// TODO check whether the logic is applied to chart
		// if the destination namespace is specified, use the specified one
		if len(f.policy.DestNamespace) > 0 {
			dstResource.Metadata.Name = strings.Replace(srcResource.Metadata.Name,
				srcResource.Metadata.Namespace, f.policy.DestNamespace, 1)
			dstResource.Metadata.Namespace = f.policy.DestNamespace
		}
		dstResources = append(dstResources, dstResource)
	}

	f.srcResources = srcResources
	f.dstResources = dstResources

	log.Debugf("resources for the execution %d fetched from the source registry", f.executionID)
	return nil
}

func (f *flow) createNamespace() error {
	// Merge the metadata of all source namespaces
	// eg:
	// We have two source namespaces:
	// {
	//	Name: "source01",
	//  Metadata: {"public": true}
	// }
	// and
	// {
	//	Name: "source02",
	//  Metadata: {"public": false}
	// }
	// The name of the destination namespace is "destination",
	// after merging the metadata, the destination namespace
	// looks like this:
	// {
	//	 Name: "destination",
	//   Metadata: {
	//		"public": {
	//			"source01": true,
	//			"source02": false,
	//		},
	//	 },
	// }
	// TODO merge the metadata of different namespaces
	namespaces := []*model.Namespace{}
	for i, resource := range f.dstResources {
		namespace := &model.Namespace{
			Name: resource.Metadata.Namespace,
		}
		// get the metadata of the namespace from the source registry
		ns, err := f.srcAdapter.GetNamespace(f.srcResources[i].Metadata.Namespace)
		if err != nil {
			f.markExecutionFailure(err)
			return err
		}
		namespace.Metadata = ns.Metadata
		namespaces = append(namespaces, namespace)
	}

	for _, namespace := range namespaces {
		if err := f.dstAdapter.CreateNamespace(namespace); err != nil {
			f.markExecutionFailure(err)
			return err
		}

		log.Debugf("namespace %s for the execution %d created on the destination registry", namespace.Name, f.executionID)
	}

	return nil
}

func (f *flow) preprocess() error {
	items, err := f.scheduler.Preprocess(f.srcResources, f.dstResources)
	if err != nil {
		f.markExecutionFailure(err)
		return err
	}
	f.scheduleItems = items
	log.Debugf("the preprocess for resources of the execution %d completed",
		f.executionID)
	return nil
}

func (f *flow) createTasks() error {
	for _, item := range f.scheduleItems {
		task := &models.Task{
			ExecutionID:  f.executionID,
			Status:       models.TaskStatusInitialized,
			ResourceType: string(item.SrcResource.Type),
			SrcResource:  getResourceName(item.SrcResource),
			DstResource:  getResourceName(item.DstResource),
		}
		id, err := f.executionMgr.CreateTask(task)
		if err != nil {
			// if failed to create the task for one of the items,
			// the whole execution is marked as failure and all
			// the items will not be submitted
			f.markExecutionFailure(err)
			return err
		}

		item.TaskID = id
		log.Debugf("task record %d for the execution %d created",
			id, f.executionID)
	}
	return nil
}

func (f *flow) schedule() error {
	results, err := f.scheduler.Schedule(f.scheduleItems)
	if err != nil {
		f.markExecutionFailure(err)
		return err
	}

	allFailed := true
	for _, result := range results {
		// if the task is failed to be submitted, update the status of the
		// task as failure
		if result.Error != nil {
			log.Errorf("failed to schedule task %d: %v", result.TaskID, err)
			if err = f.executionMgr.UpdateTaskStatus(result.TaskID, models.TaskStatusFailed); err != nil {
				log.Errorf("failed to update task status %d: %v", result.TaskID, err)
			}
			continue
		}
		allFailed = false
		// if the task is submitted successfully, update the status, job ID and start time
		if err = f.executionMgr.UpdateTaskStatus(result.TaskID, models.TaskStatusPending); err != nil {
			log.Errorf("failed to update task status %d: %v", result.TaskID, err)
		}
		if err = f.executionMgr.UpdateTask(&models.Task{
			ID:        result.TaskID,
			JobID:     result.JobID,
			StartTime: time.Now(),
		}, "JobID", "StartTime"); err != nil {
			log.Errorf("failed to update task %d: %v", result.TaskID, err)
		}
		log.Debugf("the task %d scheduled", result.TaskID)
	}
	// if all the tasks are failed, mark the execution failed
	if allFailed {
		err = errors.New("all tasks are failed")
		f.markExecutionFailure(err)
		return err
	}
	return nil
}

func (f *flow) markExecutionFailure(err error) {
	statusText := ""
	if err != nil {
		statusText = err.Error()
	}
	log.Errorf("the execution %d is marked as failure because of the error: %s",
		f.executionID, statusText)
	err = f.executionMgr.Update(
		&models.Execution{
			ID:         f.executionID,
			Status:     models.ExecutionStatusFailed,
			StatusText: statusText,
			EndTime:    time.Now(),
		}, "Status", "StatusText", "EndTime")
	if err != nil {
		log.Errorf("failed to update the execution %d: %v", f.executionID, err)
	}
}

func (f *flow) markExecutionSuccess(msg string) {
	log.Debugf("the execution %d is marked as success", f.executionID)
	err := f.executionMgr.Update(
		&models.Execution{
			ID:         f.executionID,
			Status:     models.ExecutionStatusSucceed,
			StatusText: msg,
			EndTime:    time.Now(),
		}, "Status", "StatusText", "EndTime")
	if err != nil {
		log.Errorf("failed to update the execution %d: %v", f.executionID, err)
	}
}

// return the name with format "res_name" or "res_name:[vtag1,vtag2,vtag3]"
// if the resource has vtags
func getResourceName(res *model.Resource) string {
	if res == nil {
		return ""
	}
	meta := res.Metadata
	if meta == nil {
		return ""
	}
	if len(meta.Vtags) == 0 {
		return meta.Name
	}
	return meta.Name + ":[" + strings.Join(meta.Vtags, ",") + "]"
}
