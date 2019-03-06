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

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/ng/model"

	"github.com/goharbor/harbor/src/replication/ng/execution"
	"github.com/goharbor/harbor/src/replication/ng/registry"
	"github.com/goharbor/harbor/src/replication/ng/scheduler"
)

// Controller controls the replication flow
type Controller interface {
	// Start a replication according to the policy and returns the
	// execution ID and error
	StartReplication(policy *model.Policy) (int64, error)
	// Stop the replication specified by the execution ID
	StopReplication(int64) error
}

// NewController returns an instance of a Controller
func NewController(registryMgr registry.Manager,
	executionMgr execution.Manager, scheduler scheduler.Scheduler) (Controller, error) {
	if registryMgr == nil || executionMgr == nil || scheduler == nil {
		// TODO(ChenDe): Uncomment it when execution manager is ready
		// return nil, errors.New("invalid params")
	}
	return &defaultController{
		registryMgr:  registryMgr,
		executionMgr: executionMgr,
		scheduler:    scheduler,
	}, nil
}

// defaultController is the default implement for the Controller
type defaultController struct {
	registryMgr  registry.Manager
	executionMgr execution.Manager
	scheduler    scheduler.Scheduler
}

// Start a replication according to the policy
func (d *defaultController) StartReplication(policy *model.Policy) (int64, error) {
	log.Infof("starting the replication based on the policy %d ...", policy.ID)

	flow, err := newFlow(policy, d.registryMgr, d.executionMgr, d.scheduler)
	if err != nil {
		return 0, fmt.Errorf("failed to create the flow object based on policy %d: %v", policy.ID, err)
	}

	// create the execution record
	id, err := flow.createExecution()
	if err != nil {
		return 0, fmt.Errorf("failed to create the execution record for replication based on policy %d: %v", policy.ID, err)
	}

	// fetch resources from the source registry
	if err := flow.fetchResources(); err != nil {
		// just log the error message and return the execution ID
		log.Errorf("failed to fetch resources for the execution %d: %v", id, err)
		return id, nil
	}

	// create the namespace on the destination registry
	if err = flow.createNamespace(); err != nil {
		log.Errorf("failed to create the namespace %s for the execution %d on the destination registry: %v", policy.DestNamespace, id, err)
		return id, nil
	}

	// preprocess the resources
	if err = flow.preprocess(); err != nil {
		log.Errorf("failed to preprocess the resources for the execution %d: %v", id, err)
		return id, nil
	}

	// create task records in database
	if err = flow.createTasks(); err != nil {
		log.Errorf("failed to create task records for the execution %d: %v", id, err)
		return id, nil
	}

	// schedule the tasks
	if err = flow.schedule(); err != nil {
		log.Errorf("failed to schedule the execution %d: %v", id, err)
		return id, nil
	}

	log.Infof("the execution %d scheduled", id)
	return id, nil
}

func (d *defaultController) StopReplication(id int64) error {
	// TODO
	return nil
}
