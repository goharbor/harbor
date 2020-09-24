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

package admin

import (
	"context"
	"encoding/json"

	o "github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	job_model "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/quota"
	"github.com/goharbor/harbor/src/controller/scan"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/service/notifications"
	j "github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
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

// Handler handles request on /service/notifications/jobs/adminjob/*, which listens to the webhook of jobservice.
type Handler struct {
	notifications.BaseHandler
	id            int64
	UUID          string
	status        string
	UpstreamJobID string
	revision      int64
	jobName       string
	checkIn       string
	statusCode    uint16
}

// Prepare ...
func (h *Handler) Prepare() {
	h.BaseHandler.Prepare()
	var data job_model.JobStatusChange
	err := json.Unmarshal(h.Ctx.Input.CopyBody(1<<32), &data)
	if err != nil {
		log.Errorf("Failed to decode job status change, error: %v", err)
		h.Abort("200")
		return
	}
	id, err := h.GetInt64FromPath(":id")
	if err != nil {
		log.Errorf("Failed to get job ID, error: %v", err)
		// Avoid job service from resending...
		h.Abort("200")
		return
	}
	h.id = id
	// UpstreamJobID is the periodic job id
	if data.Metadata.UpstreamJobID != "" {
		h.UUID = data.Metadata.UpstreamJobID
	} else {
		h.UUID = data.JobID
	}

	status, ok := statusMap[data.Status]
	if !ok {
		log.Infof("drop the job status update event: job id-%d, status-%s", h.id, status)
		h.Abort("200")
		return
	}
	h.statusCode = (uint16)(j.Status(data.Status).Code())
	h.status = status
	h.revision = data.Metadata.Revision
	h.jobName = data.Metadata.JobName
	h.checkIn = data.CheckIn
}

// HandleAdminJob handles the webhook of admin jobs
func (h *Handler) HandleAdminJob() {
	log.Infof("received admin job status update event: job-%d, job_uuid-%s, status-%s, revision-%d", h.id, h.UUID, h.status, h.revision)

	// create the mapping relationship between the jobs in database and jobservice
	if err := dao.SetAdminJobUUID(h.id, h.UUID); err != nil {
		h.SendInternalServerError(err)
		return
	}
	if err := dao.UpdateAdminJobStatus(h.id, h.status, h.statusCode, h.revision); err != nil {
		log.Errorf("Failed to update job status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}

	// For scan all job
	if h.jobName == job.ImageScanAllJob && h.checkIn != "" {
		go scan.HandleCheckIn(orm.NewContext(context.TODO(), o.NewOrm()), h.checkIn)
	} else if h.jobName == job.ImageGC && h.status == models.JobFinished {
		go func() {
			if config.QuotaPerProjectEnable() {
				quota.RefreshForProjects(orm.NewContext(context.TODO(), o.NewOrm()))
			}
		}()
	}
}
