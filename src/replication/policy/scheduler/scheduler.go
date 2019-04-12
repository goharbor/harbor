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

package scheduler

import (
	"fmt"
	"net/http"
	"time"

	common_http "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/job"
	job_models "github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/dao"
	"github.com/goharbor/harbor/src/replication/dao/models"
)

// Scheduler can be used to schedule or unschedule a scheduled policy
// Currently, the default scheduler implements its capabilities by delegating
// the scheduled job of jobservice
type Scheduler interface {
	Schedule(policyID int64, cron string) error
	Unschedule(policyID int64) error
}

// NewScheduler returns an instance of scheduler
func NewScheduler(js job.Client) Scheduler {
	return &scheduler{
		jobservice: js,
	}
}

type scheduler struct {
	jobservice job.Client
}

func (s *scheduler) Schedule(policyID int64, cron string) error {
	now := time.Now()
	id, err := dao.ScheduleJob.Add(&models.ScheduleJob{
		PolicyID:     policyID,
		Status:       job.JobServiceStatusPending,
		CreationTime: now,
		UpdateTime:   now,
	})
	if err != nil {
		return err
	}
	log.Debugf("the schedule job record %d added", id)

	statusHookURL := fmt.Sprintf("%s/service/notifications/jobs/replication/%d", config.Config.CoreURL, id)
	jobID, err := s.jobservice.SubmitJob(&job_models.JobData{
		Name: job.ReplicationScheduler,
		Parameters: map[string]interface{}{
			"url":       config.Config.CoreURL,
			"policy_id": policyID,
		},
		Metadata: &job_models.JobMetadata{
			JobKind: job.JobKindPeriodic,
			Cron:    cron,
		},
		StatusHook: statusHookURL,
	})
	if err != nil {
		// clean up the record in database
		if e := dao.ScheduleJob.Delete(id); e != nil {
			log.Errorf("failed to delete the schedule job %d: %v", id, e)
		} else {
			log.Debugf("the schedule job record %d deleted", id)
		}
		return err
	}
	log.Debugf("the schedule job for policy %d submitted to the jobservice", policyID)

	err = dao.ScheduleJob.Update(&models.ScheduleJob{
		ID:    id,
		JobID: jobID,
	}, "JobID")
	log.Debugf("the policy %d scheduled", policyID)
	return err
}

func (s *scheduler) Unschedule(policyID int64) error {
	sjs, err := dao.ScheduleJob.List(&models.ScheduleJobQuery{
		PolicyID: policyID,
	})
	if err != nil {
		return err
	}
	for _, sj := range sjs {
		if err = s.jobservice.PostAction(sj.JobID, job.JobActionStop); err != nil {
			// if the job specified by jobID is not found in jobservice, just delete
			// the record from database
			if e, ok := err.(*common_http.Error); !ok || e.Code != http.StatusNotFound {
				return err
			}
			log.Debugf("the stop action for schedule job %s submitted to the jobservice", sj.JobID)
		}
		if err = dao.ScheduleJob.Delete(sj.ID); err != nil {
			return err
		}
		log.Debugf("the schedule job record %d deleted", sj.ID)
	}
	log.Debugf("the policy %d unscheduled", policyID)
	return nil
}
