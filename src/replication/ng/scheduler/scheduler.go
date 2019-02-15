package scheduler

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/goharbor/harbor/src/common/job"
	common_job "github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/jobservice/opm"
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// DefaultReplicator provides a default implement for Replicator
type DefaultReplicator struct {
	client job.Client
}

// NewDefaultReplicator returns an instance of DefaultReplicator
func NewDefaultReplicator(client job.Client) *DefaultReplicator {
	return &DefaultReplicator{
		client: client,
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
func (d *DefaultReplicator) Preprocess(srcResources []*model.Resource, destResources []*model.Resource) ([]*ScheduleItem, error) {
	if len(srcResources) != len(destResources) {
		err := errors.New("srcResources has different length with destResources")
		log.Errorf(err.Error())
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
func (d *DefaultReplicator) Schedule(items []*ScheduleItem) ([]*ScheduleResult, error) {
	var results []*ScheduleResult
	for _, item := range items {
		if item.TaskID == 0 {
			err := errors.New("some tasks do not have a ID")
			log.Errorf(err.Error())
			return nil, err
		}
		result := &ScheduleResult{
			TaskID: item.TaskID,
		}
		job := &models.JobData{
			Metadata: &models.JobMetadata{
				JobKind: job.JobKindGeneric,
			},
			StatusHook: fmt.Sprintf("%s/service/notifications/jobs/replication/%d", config.InternalCoreURL(), item.TaskID),
		}

		job.Name = common_job.ImageTransfer
		src, err := json.Marshal(item.SrcResource)
		if err != nil {
			log.Errorf("failed to marshal the srcResource of %v.err:%s!", item.SrcResource, err.Error())
			result.Error = err
			results = append(results, result)
			continue
		}
		dest, err := json.Marshal(item.DstResource)
		if err != nil {
			log.Errorf("failed to marshal the dstResource of %v.err:%s!", item.DstResource, err.Error())
			result.Error = err
			results = append(results, result)
			continue
		}
		job.Parameters = map[string]interface{}{
			"src_resource": string(src),
			"dst_resource": string(dest),
		}
		_, joberr := d.client.SubmitJob(job)
		if joberr != nil {
			log.Errorf("failed to submit the task:%v .err:%s!", item, joberr.Error())
			result.Error = joberr
			results = append(results, result)
			continue
		}
		results = append(results, result)
	}
	return results, nil

}

// Stop the transfer job
func (d *DefaultReplicator) Stop(id string) error {

	err := d.client.PostAction(id, opm.CtlCommandStop)
	if err != nil {
		return err
	}
	return nil

}
