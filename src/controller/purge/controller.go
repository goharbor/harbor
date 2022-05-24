//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package purge

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
)

const (
	// SchedulerCallback ...
	SchedulerCallback = "PURGE_AUDIT_LOG_CALLBACK"
	// VendorType ...
	VendorType = "PURGE_AUDIT_LOG"
)

// Ctrl a global purge controller instance
var Ctrl = NewController()

func init() {
	err := scheduler.RegisterCallbackFunc(SchedulerCallback, purgeCallback)
	if err != nil {
		log.Fatalf("failed to registry purge job call back, %v", err)
	}
}

func purgeCallback(ctx context.Context, p string) error {
	param := &JobPolicy{}
	if err := json.Unmarshal([]byte(p), param); err != nil {
		return fmt.Errorf("failed to unmashal the param: %v", err)
	}
	_, err := Ctrl.Start(ctx, *param, task.ExecutionTriggerSchedule)
	return err
}

// Controller defines the interface with the purge job
type Controller interface {
	// Start kick off a purge schedule
	Start(ctx context.Context, policy JobPolicy, trigger string) (int64, error)
}

type controller struct {
	taskMgr task.Manager
	exeMgr  task.ExecutionManager
}

func (c *controller) Start(ctx context.Context, policy JobPolicy, trigger string) (int64, error) {
	para := make(map[string]interface{})

	para[common.PurgeAuditDryRun] = policy.DryRun
	para[common.PurgeAuditRetentionHour] = policy.RetentionHour
	para[common.PurgeAuditIncludeOperations] = policy.IncludeOperations

	execID, err := c.exeMgr.Create(ctx, VendorType, -1, trigger, para)
	if err != nil {
		return -1, err
	}
	_, err = c.taskMgr.Create(ctx, execID, &task.Job{
		Name: job.PurgeAudit,
		Metadata: &job.Metadata{
			JobKind: job.KindGeneric,
		},
		Parameters: para,
	})
	if err != nil {
		return -1, err
	}
	return execID, nil
}

// NewController ...
func NewController() Controller {
	return &controller{
		taskMgr: task.NewManager(),
		exeMgr:  task.NewExecutionManager(),
	}
}
