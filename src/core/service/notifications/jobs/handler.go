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
	jobmodels "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/replication"
	"github.com/goharbor/harbor/src/replication/operation/hook"
	"github.com/goharbor/harbor/src/replication/policy/scheduler"
	"github.com/goharbor/harbor/src/webhook"
)

var statusMap = map[string]string{
	job.JobServiceStatusPending:   models.JobPending,
	job.JobServiceStatusRunning:   models.JobRunning,
	job.JobServiceStatusStopped:   models.JobStopped,
	job.JobServiceStatusCancelled: models.JobCanceled,
	job.JobServiceStatusError:     models.JobError,
	job.JobServiceStatusSuccess:   models.JobFinished,
	job.JobServiceStatusScheduled: models.JobScheduled,
}

// Handler handles reqeust on /service/notifications/jobs/*, which listens to the webhook of jobservice.
type Handler struct {
	api.BaseController
	id        int64
	status    string
	rawStatus string
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
	var data jobmodels.JobStatusChange
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
		log.Errorf("Failed to update replication task status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}
}

// HandleWebhookJob handles the hook of webhook job
func (h *Handler) HandleWebhookJob() {
	log.Debugf("received webhook task status update event: task-%d, status-%s", h.id, h.status)
	if err := webhook.JobCtl.UpdateWebhookJob(&models.WebhookJob{
		ID:         h.id,
		Status:     h.status,
		UpdateTime: time.Now(),
	}, "Status", "UpdateTime"); err != nil {
		log.Errorf("Failed to update webhook job status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}
}
