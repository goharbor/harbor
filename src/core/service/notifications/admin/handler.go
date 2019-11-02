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
	"encoding/json"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/job"
	job_model "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/models"
	cutils "github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/core/utils"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/pkg/errors"
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
	api.BaseController
	id            int64
	UUID          string
	status        string
	UpstreamJobID string
}

// Prepare ...
func (h *Handler) Prepare() {
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
	h.status = status
}

// HandleAdminJob handles the webhook of admin jobs
func (h *Handler) HandleAdminJob() {
	log.Infof("received admin job status update event: job-%d, status-%s", h.id, h.status)
	// create the mapping relationship between the jobs in database and jobservice
	if err := dao.SetAdminJobUUID(h.id, h.UUID); err != nil {
		h.SendInternalServerError(err)
		return
	}
	if err := dao.UpdateAdminJobStatus(h.id, h.status); err != nil {
		log.Errorf("Failed to update job status, id: %d, status: %s", h.id, h.status)
		h.SendInternalServerError(err)
		return
	}
}

// scanAllArtifacts implements a post action for the scan all job,
// which now is used to send signal of requesting scan all. The purpose of
// implementing such way is avoid scan all related data migration.
// Obviously, this is not good way but it's an accept way at present.
//
// Once the status of scan all job is success, this method will be called
// in a non-blocking way.
//
// This is a try-best approach, failures will not cause the whole process exit.
func (h *Handler) scanAllArtifacts() error {
	// Ignore the input arguments as the methods called here will not depend on them.
	// This is technically wrong, but it's a simple way to leverage this manager.
	repositoryMgr := repository.New(nil, nil)

	// Get all the repositories first.
	repos, err := repositoryMgr.ListImageRepositories()
	if err != nil {
		return errors.Wrap(err, "admin job handler: scan all artifacts")
	}

	if len(repos) == 0 {
		// Treat as complete work
		return nil
	}

	// Used to collect errors in the retrieving goroutines.
	errs := make([]error, len(repos))
	// Used to control the scale of goroutines.
	tokens := make(chan interface{}, 10)

	for i, repo := range repos {
		// Get one token first for processing
		<-tokens

		go func(repoName string) {
			defer func() {
				// Return the token
				tokens <- struct{}{}
			}()

			rclient, err := utils.NewRepositoryClientForUI("harbor-core", repoName)
			if err != nil {
				errs[i] = errors.Wrap(err, "scan controller: scan all")
				return
			}

			// Retrieve all tags
			tags, err := rclient.ListTag()
			if err != nil {
				errs[i] = errors.Wrap(err, "scan controller: scan all")
				return
			}

			// Scan one by one
			for _, tag := range tags {
				_, exits, er := rclient.ManifestExist(tag)
				if er == nil && !exits {
					er = errors.Errorf("tag %s does not exists in repository %s", tag, repoName)
				}

				if er != nil {
					// Append
					if err != nil {
						err = errors.Wrap(er, err.Error())
					} else {
						err = er
					}

					continue
				}

				proName, _ := cutils.ParseRepository(repoName)
				_, err := h.ProjectMgr.Get(proName)
				if err != nil {
					errs[i] = err
				}
			}
		}(repo.Name)
	}

	return nil
}
