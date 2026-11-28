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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	policymodel "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

func TestStartHookMarksExecutionErrorWhenTaskCreationFails(t *testing.T) {
	execMgr := &tasktesting.ExecutionManager{}
	taskMgr := &tasktesting.Manager{}
	hm := &DefaultManager{execMgr: execMgr, taskMgr: taskMgr}

	execMgr.On("Create", mock.Anything, job.WebhookJobVendorType, int64(1), mock.Anything, mock.Anything).Return(int64(7), nil)
	taskMgr.On("Create", mock.Anything, int64(7), mock.Anything).Return(int64(0), errors.New("jobservice unreachable"))
	execMgr.On("MarkError", mock.Anything, int64(7), "jobservice unreachable").Return(nil)

	event := &model.HookEvent{PolicyID: 1, EventType: "PUSH_ARTIFACT", Target: &policymodel.EventTarget{Type: model.NotifyTypeHTTP}}
	data := &models.JobData{Name: job.WebhookJobVendorType, Metadata: &models.JobMetadata{JobKind: job.KindGeneric}, Parameters: map[string]any{"payload": "{}"}}

	assert.Error(t, hm.StartHook(context.Background(), event, data))
	execMgr.AssertExpectations(t)
}
