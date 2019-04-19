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
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/operation"
)

// UpdateTask update the status of the task
func UpdateTask(ctl operation.Controller, id int64, status string) error {
	// convert the job status to task status
	s := ""
	switch status {
	case job.JobStatusPending:
		s = models.TaskStatusPending
	case job.JobStatusScheduled, job.JobStatusRunning:
		s = models.TaskStatusInProgress
	case job.JobStatusStopped, job.JobStatusCancelled:
		s = models.TaskStatusStopped
	case job.JobStatusError:
		s = models.TaskStatusFailed
	case job.JobStatusSuccess:
		s = models.TaskStatusSucceed
	}
	return ctl.UpdateTaskStatus(id, s)
}
