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
	"github.com/goharbor/harbor/src/lib/orm"
	"time"

	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/service/notifications"
	jjob "github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
)

var statusMap = map[string]string{
	job.JobServiceStatusPending:   models.JobPending,
	job.JobServiceStatusScheduled: models.JobScheduled,
	job.JobServiceStatusRunning:   models.JobRunning,
	job.JobServiceStatusStopped:   models.JobStopped,
	job.JobServiceStatusError:     models.JobError,
	job.JobServiceStatusSuccess:   models.JobFinished,
}

// Handler handles request on /service/notifications/jobs/*, which listens to the webhook of jobservice.
type Handler struct {
	notifications.BaseHandler
	id        int64
	status    string
	rawStatus string
	checkIn   string
	revision  int64
	trackID   string
	change    *jjob.StatusChange
}

// Prepare ...
func (h *Handler) Prepare() {
	h.BaseHandler.Prepare()
	h.trackID = h.GetStringFromPath(":uuid")
	if len(h.trackID) == 0 {
		id, err := h.GetInt64FromPath(":id")
		if err != nil {
			log.Errorf("Failed to get job ID, error: %v", err)
			// Avoid job service from resending...
			h.Abort("200")
			return
		}
		h.id = id
	}

	var data jjob.StatusChange
	err := json.Unmarshal(h.Ctx.Input.CopyBody(1<<32), &data)
	if err != nil {
		log.Errorf("Failed to decode job status change with error: %v", err)
		h.Abort("200")
		return
	}
	h.change = &data
	h.rawStatus = data.Status
	status, ok := statusMap[data.Status]
	if !ok {
		log.Debugf("drop the job status update event: job id-%d/track id-%s, status-%s", h.id, h.trackID, status)
		h.Abort("200")
		return
	}
	h.status = status
	h.checkIn = data.CheckIn
	if data.Metadata != nil {
		h.revision = data.Metadata.Revision
	}
}

// HandleNotificationJob handles the hook of notification job
func (h *Handler) HandleNotificationJob() {
	log.Debugf("received notification job status update event: job-%d, status-%s", h.id, h.status)
	if err := notification.JobMgr.Update(orm.Context(), &model.Job{
		ID:         h.id,
		Status:     h.status,
		UpdateTime: time.Now(),
	}, "Status", "UpdateTime"); err != nil {
		log.Errorf("Failed to update notification job status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}
}
