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

package task

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/task/dao"
)

var (
	// HkHandler is a global instance of the HookHandler
	HkHandler = NewHookHandler()
)

// NewHookHandler creates a hook handler instance
func NewHookHandler() *HookHandler {
	return &HookHandler{
		taskDAO:      dao.NewTaskDAO(),
		executionDAO: dao.NewExecutionDAO(),
	}
}

// HookHandler handles the job status changing webhook
type HookHandler struct {
	taskDAO      dao.TaskDAO
	executionDAO dao.ExecutionDAO
}

// Handle the job status changing webhook
func (h *HookHandler) Handle(ctx context.Context, sc *job.StatusChange) error {
	logger := log.GetLogger(ctx)

	jobID := sc.JobID
	// the "JobID" field of some kinds of jobs are set as "87bbdee19bed5ce09c48a149@1605104520" which contains "@".
	// In this case, read the parent periodical job ID from "sc.Metadata.UpstreamJobID"
	if len(sc.Metadata.UpstreamJobID) > 0 {
		jobID = sc.Metadata.UpstreamJobID
	}
	tasks, err := h.taskDAO.List(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"JobID": jobID,
		},
	})
	if err != nil {
		return err
	}
	if len(tasks) == 0 {
		return errors.New(nil).WithCode(errors.NotFoundCode).
			WithMessage("task with job ID %s not found", sc.JobID)
	}
	task := tasks[0]
	execution, err := h.executionDAO.Get(ctx, task.ExecutionID)
	if err != nil {
		return err
	}

	// process check in data
	if len(sc.CheckIn) > 0 {
		processor, exist := checkInProcessorRegistry[execution.VendorType]
		if !exist {
			return fmt.Errorf("the check in processor for task %d not found", task.ID)
		}
		t := &Task{}
		t.From(task)
		return processor(ctx, t, sc)
	}

	// update task status
	if err = h.taskDAO.UpdateStatus(ctx, task.ID, sc.Status, sc.Metadata.Revision); err != nil {
		return err
	}
	// run the status change post function
	if fc, exist := statusChangePostFuncRegistry[execution.VendorType]; exist {
		if err = fc(ctx, task.ID, sc.Status); err != nil {
			logger.Errorf("failed to run the task status change post function for task %d: %v", task.ID, err)
		}
	}

	// update execution status
	statusChanged, currentStatus, err := h.executionDAO.RefreshStatus(ctx, task.ExecutionID)
	if err != nil {
		return err
	}
	// run the status change post function
	if fc, exist := executionStatusChangePostFuncRegistry[execution.VendorType]; exist && statusChanged {
		if err = fc(ctx, task.ExecutionID, currentStatus); err != nil {
			logger.Errorf("failed to run the execution status change post function for execution %d: %v", task.ExecutionID, err)
		}
	}
	return nil
}
