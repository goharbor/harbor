// Copyright 2018 Project Harbor Authors
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

package jobs

import (
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	jjob "github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/retention"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/operation/hook"
	"github.com/goharbor/harbor/src/replication/policy/scheduler"
)

var statusMap = map[string]string{
	job.JobServiceStatusPending:   models.JobPending,
	job.JobServiceStatusScheduled: models.JobScheduled,
	job.JobServiceStatusRunning:   models.JobRunning,
	job.JobServiceStatusStopped:   models.JobStopped,
	job.JobServiceStatusError:     models.JobError,
	job.JobServiceStatusSuccess:   models.JobFinished,
}

// Handler handles reqeust on /service/notifications/jobs/*, which listens to the webhook of jobservice.
type Handler struct {
	api.BaseController
	id        int64
	status    string
	rawStatus string
	checkIn   string
	revision  int64
}

// Prepare ...
func (h *Handler) Prepare() {
	id, err := h.GetInt64FromPath(":id")
	if err != nil {
		log.Errorf("Failed to get job ID, error: %v", err)
		// Avoid job service from resending...
		h.Abort("200")
		return
	}
	h.id = id
	var data jjob.StatusChange
	err = json.Unmarshal(h.Ctx.Input.CopyBody(1<<32), &data)
	if err != nil {
		log.Errorf("Failed to decode job status change, job ID: %d, error: %v", id, err)
		h.Abort("200")
		return
	}
	h.rawStatus = data.Status
	status, ok := statusMap[data.Status]
	if !ok {
		log.Debugf("drop the job status update event: job id-%d, status-%s", id, status)
		h.Abort("200")
		return
	}
	h.status = status
	h.checkIn = data.CheckIn
	if data.Metadata != nil {
		h.revision = data.Metadata.Revision
	}
}

// HandleScan handles the webhook of scan job
func (h *Handler) HandleScan() {
	log.Debugf("received san job status update event: job-%d, status-%s", h.id, h.status)
	if err := dao.UpdateScanJobStatus(h.id, h.status); err != nil {
		log.Errorf("Failed to update job status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}
}

// HandleReplicationScheduleJob handles the webhook of replication schedule job
func (h *Handler) HandleReplicationScheduleJob() {
	log.Debugf("received replication schedule job status update event: schedule-job-%d, status-%s", h.id, h.status)
	if err := scheduler.UpdateStatus(h.id, h.status); err != nil {
		log.Errorf("Failed to update job status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}
}

// HandleReplicationTask handles the webhook of replication task
func (h *Handler) HandleReplicationTask() {
	log.Debugf("received replication task status update event: task-%d, status-%s", h.id, h.status)
	if err := hook.UpdateTask(replication.OperationCtl, h.id, h.rawStatus); err != nil {
		log.Errorf("failed to update the status of the replication task %d: %v", h.id, err)
		h.SendInternalServerError(err)
		return
	}
}

// HandleRetentionTask handles the webhook of retention task
func (h *Handler) HandleRetentionTask() {
	taskID := h.id
	status := h.rawStatus
	log.Debugf("received retention task status update event: task-%d, status-%s", taskID, status)
	mgr := &retention.DefaultManager{}
	// handle checkin
	if h.checkIn != "" {
		var retainObj struct {
			Total    int `json:"total"`
			Retained int `json:"retained"`
		}
		if err := json.Unmarshal([]byte(h.checkIn), &retainObj); err != nil {
			log.Errorf("failed to resolve checkin of retention task %d: %v", taskID, err)
			return
		}
		task := &retention.Task{
			ID:       taskID,
			Total:    retainObj.Total,
			Retained: retainObj.Retained,
		}
		if err := mgr.UpdateTask(task, "Total", "Retained"); err != nil {
			log.Errorf("failed to update of retention task %d: %v", taskID, err)
			h.SendInternalServerError(err)
			return
		}
		return
	}

	// handle status updating
	if err := mgr.UpdateTaskStatus(taskID, status, h.revision); err != nil {
		log.Errorf("failed to update the status of retention task %d: %v", taskID, err)
		h.SendInternalServerError(err)
		return
	}
}

// HandleNotificationJob handles the hook of notification job
func (h *Handler) HandleNotificationJob() {
	log.Debugf("received notification job status update event: job-%d, status-%s", h.id, h.status)
	if err := notification.JobMgr.Update(&models.NotificationJob{
		ID:         h.id,
		Status:     h.status,
		UpdateTime: time.Now(),
	}, "Status", "UpdateTime"); err != nil {
		log.Errorf("Failed to update notification job status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}
}
