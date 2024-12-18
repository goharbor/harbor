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

package hook

import (
	"context"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/goharbor/harbor/src/pkg/task"
)

// Manager send hook
type Manager interface {
	StartHook(context.Context, *model.HookEvent, *models.JobData) error
}

// DefaultManager ...
type DefaultManager struct {
	execMgr task.ExecutionManager
	taskMgr task.Manager
}

// NewHookManager ...
func NewHookManager() *DefaultManager {
	return &DefaultManager{
		execMgr: task.ExecMgr,
		taskMgr: task.Mgr,
	}
}

// StartHook create a webhook job record in database, and submit it to jobservice
func (hm *DefaultManager) StartHook(ctx context.Context, event *model.HookEvent, data *models.JobData) error {
	var vendorType string
	switch event.Target.Type {
	case model.NotifyTypeHTTP:
		vendorType = job.WebhookJobVendorType
	case model.NotifyTypeSlack:
		vendorType = job.SlackJobVendorType
	case model.NotifyTypeTeams:
		vendorType = job.TeamsJobVendorType
	}

	if len(vendorType) == 0 {
		return errors.Errorf("invalid event target type: %s", event.Target.Type)
	}

	extraAttrs := map[string]interface{}{
		"event_type": event.EventType,
		"payload":    data.Parameters["payload"],
	}
	// create execution firstly, then create task.
	execID, err := hm.execMgr.Create(ctx, vendorType, event.PolicyID, task.ExecutionTriggerEvent, extraAttrs)
	if err != nil {
		return errors.Errorf("failed to create execution for webhook based on policy %d: %v", event.PolicyID, err)
	}

	taskID, err := hm.taskMgr.Create(ctx, execID, &task.Job{
		Name: data.Name,
		Metadata: &job.Metadata{
			JobKind: data.Metadata.JobKind,
		},
		Parameters: map[string]interface{}(data.Parameters),
	})
	if err != nil {
		return errors.Errorf("failed to create task for webhook based on policy %d: %v", event.PolicyID, err)
	}

	log.Debugf("created webhook task %d for the policy %d", taskID, event.PolicyID)

	return nil
}
