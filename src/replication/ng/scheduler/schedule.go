package schedule

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

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

// Scheduler to schedule the tasks to transfer resource data
type Scheduler interface {
	// Schedule tasks
	Schedule(srcResources []*model.Resource, destResources []*model.Resource) ([]*model.Task, error)
	StopExecution(executionID string) error
}

// Schedule the tasks base on resources
func (d *DefaultReplicator) Schedule(srcResources []*model.Resource, destResources []*model.Resource) ([]*model.Task, error) {
	if len(srcResources) != len(destResources) {
		err := errors.New("srcResources has different length with destResources")
		log.Errorf(err.Error())
		return nil, err
	}
	var tasks []*model.Task
	for index, srcResource := range srcResources {
		destResource := destResources[index]
		task := &model.Task{}
		task.ResourceType = srcResource.Type
		task.StartTime = time.Now().UTC()
		src, err := json.Marshal(srcResource)
		if err != nil {
			log.Errorf("failed to marshal the srcResource of %v.err:%s!", srcResource, err.Error())
			task.Status = "Error"
			tasks = append(tasks, task)
			continue
		}
		task.SrcResource = string(src)
		dest, err := json.Marshal(destResource)
		if err != nil {
			log.Errorf("failed to marshal the destResource of %v.err:%s!", destResource, err.Error())
			task.Status = "Error"
			tasks = append(tasks, task)
			continue
		}
		task.DstResource = string(dest)
		task.Status = "Initial"
		tasks = append(tasks, task)

	}
	return tasks, nil
}

// SubmitTasks transfer the tasks to jobs,and then submit these jobs to job service.
func (d *DefaultReplicator) SubmitTasks(tasks []*model.Task) ([]*model.Task, error) {
	for _, task := range tasks {
		if task.ID == 0 {
			err := errors.New("task do not have ID")
			log.Errorf(err.Error())
			return nil, err
		}
		job := &models.JobData{
			Metadata: &models.JobMetadata{
				JobKind: job.JobKindGeneric,
			},
			StatusHook: fmt.Sprintf("%s/service/notifications/jobs/replication/%d", config.InternalCoreURL(), task.ID),
		}

		job.Name = common_job.ImageTransfer
		job.Parameters = map[string]interface{}{
			"src_resource": task.SrcResource,
			"dst_resource": task.DstResource,
		}
		uuid, err := d.client.SubmitJob(job)
		if err != nil {
			log.Errorf("failed to submit the task:%v .err:%s!", task, err.Error())
			task.Status = "Error"
			tasks = append(tasks, task)
			continue
		}
		task.JobID = uuid
		task.Status = "Pending"
		tasks = append(tasks, task)
	}
	return tasks, nil

}

// StopExecution to stop the transfer job
func (d *DefaultReplicator) StopExecution(executionID string) error {

	err := d.client.PostAction(executionID, opm.CtlCommandStop)
	if err != nil {
		return err
	}
	return nil

}
