// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"encoding/json"
	"fmt"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/job"
	job_model "github.com/vmware/harbor/src/common/job/models"
	"github.com/vmware/harbor/src/common/models"
	common_models "github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/api"
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

// Handler handles reqeust on /service/notifications/jobs/adminjob/*, which listens to the webhook of jobservice.
type Handler struct {
	api.BaseController
	id     int64
	status string
}

// Prepare ...
func (h *Handler) Prepare() {
	// id, err := h.GetInt64FromPath(":id")
	// if err != nil {
	// 	log.Errorf("Failed to get job ID, error: %v", err)
	// 	//Avoid job service from resending...
	// 	h.Abort("200")
	// 	return
	// }

	var data job_model.JobStatusChange
	err := json.Unmarshal(h.Ctx.Input.CopyBody(1<<32), &data)
	if err != nil {
		log.Errorf("Failed to decode job status change, error: %v", err)
		h.Abort("200")
		return
	}
	// get the id for hook request body.
	query := &common_models.AdminJobQuery{
		UUID: data.JobID,
	}
	jobs, err := dao.GetAdminJobs(query)
	if err != nil {
		h.HandleInternalServerError(fmt.Sprintf("%v", err))
		return
	}
	if len(jobs) != 1 {
		log.Warningf("find more than one admin jobs in hook.")
		h.Abort("500")
		return
	}
	h.id = jobs[0].ID
	status, ok := statusMap[data.Status]
	if !ok {
		log.Infof("drop the job status update event: job id-%d, status-%s", id, status)
		h.Abort("200")
		return
	}
	h.status = status
}

//HandleAdminJob handles the webhook of admin jobs
func (h *Handler) HandleAdminJob() {
	log.Infof("received admin job status update event: job-%d, status-%s", h.id, h.status)
	if err := dao.UpdateAdminJobStatus(h.id, h.status); err != nil {
		log.Errorf("Failed to update job status, id: %d, status: %s", h.id, h.status)
		h.HandleInternalServerError(err.Error())
		return
	}
}
