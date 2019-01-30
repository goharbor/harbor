package schedule

import (
	"encoding/json"
	"time"

	"github.com/goharbor/harbor/src/common/job"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/common/utils/log"
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
	StopTransfer(jobID string) error
}

// Schedule the task to transfer resouce data
func (d *DefaultReplicator) Schedule(srcResources []*model.Resource, destResources []*model.Resource) ([]*model.Task, error) {
	var tasks []*model.Task
	for _, destResource := range destResources {

		for _, srcResource := range srcResources {
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

			newjob := &models.JobData{
				Metadata: &models.JobMetadata{
					JobKind: job.JobKindGeneric,
				},
			}

			newjob.Name = job.ImageTransfer
			newjob.Parameters = map[string]interface{}{
				"src_resource": srcResource,
				"dst_resource": destResource,
			}
			uuid, err := d.client.SubmitJob(newjob)
			if err != nil {
				log.Errorf("failed to submit the job from %v to %v.err:%s!", srcResource, destResource, err.Error())
			}
			task.JobID = uuid
			task.Status = ""
			tasks = append(tasks, task)

		}
	}
	return tasks, nil
}

// StopTransfer to stop the transfer job
func (d *DefaultReplicator) StopTransfer(jobID string) error {

	err := d.client.PostAction(jobID, opm.CtlCommandStop)
	if err != nil {
		return err
	}
	return nil

}
