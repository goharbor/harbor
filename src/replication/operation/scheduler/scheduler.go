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
	"encoding/json"
	"errors"
	"fmt"

	cjob "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/model"
)

type defaultScheduler struct {
	client cjob.Client
}

// NewScheduler returns an instance of Scheduler
func NewScheduler(js cjob.Client) Scheduler {
	return &defaultScheduler{
		client: js,
	}
}

// ScheduleItem is an item that can be scheduled
type ScheduleItem struct {
	TaskID      int64 // used as the param in the hook
	SrcResource *model.Resource
	DstResource *model.Resource
}

// ScheduleResult is the result of the schedule for one item
type ScheduleResult struct {
	TaskID int64
	JobID  string
	Error  error
}

// Scheduler schedules
type Scheduler interface {
	// Preprocess the resources and returns the item list that can be scheduled
	Preprocess([]*model.Resource, []*model.Resource) ([]*ScheduleItem, error)
	// Schedule the items. If got error when scheduling one of the items,
	// the error should be put in the corresponding ScheduleResult and the
	// returning error of this function should be nil
	Schedule([]*ScheduleItem) ([]*ScheduleResult, error)
	// Stop the job specified by ID
	Stop(id string) error
}

// Preprocess the resources and returns the item list that can be scheduled
func (d *defaultScheduler) Preprocess(srcResources []*model.Resource, destResources []*model.Resource) ([]*ScheduleItem, error) {
	if len(srcResources) != len(destResources) {
		err := errors.New("srcResources has different length with destResources")
		return nil, err
	}
	var items []*ScheduleItem
	for index, srcResource := range srcResources {
		destResource := destResources[index]
		item := &ScheduleItem{
			SrcResource: srcResource,
			DstResource: destResource,
		}
		items = append(items, item)

	}
	return items, nil
}

// Schedule transfer the tasks to jobs,and then submit these jobs to job service.
func (d *defaultScheduler) Schedule(items []*ScheduleItem) ([]*ScheduleResult, error) {
	var results []*ScheduleResult
	for _, item := range items {
		result := &ScheduleResult{
			TaskID: item.TaskID,
		}
		if item.TaskID == 0 {
			result.Error = errors.New("some tasks do not have a ID")
			results = append(results, result)
			continue
		}
		j := &models.JobData{
			Metadata: &models.JobMetadata{
				JobKind: job.KindGeneric,
			},
			StatusHook: fmt.Sprintf("%s/service/notifications/jobs/replication/task/%d", config.Config.CoreURL, item.TaskID),
		}

		j.Name = job.Replication
		src, err := json.Marshal(item.SrcResource)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}
		dest, err := json.Marshal(item.DstResource)
		if err != nil {
			result.Error = err
			results = append(results, result)
			continue
		}
		j.Parameters = map[string]interface{}{
			"src_resource": string(src),
			"dst_resource": string(dest),
		}
		id, joberr := d.client.SubmitJob(j)
		if joberr != nil {
			result.Error = joberr
			results = append(results, result)
			continue
		}
		result.JobID = id
		results = append(results, result)
	}
	return results, nil
}

// Stop the transfer job
func (d *defaultScheduler) Stop(id string) error {
	err := d.client.PostAction(id, string(job.StopCommand))
	if err != nil {
		return err
	}
	return nil

}
