package manager

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/dao/notification"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/pkg/notification/job"
)

// DefaultManager ..
type DefaultManager struct {
}

// NewDefaultManager ...
func NewDefaultManager() job.Manager {
	return &DefaultManager{}
}

// Create ...
func (d *DefaultManager) Create(job *models.NotificationJob) (int64, error) {
	return notification.AddNotificationJob(job)
}

// List ...
func (d *DefaultManager) List(query ...*models.NotificationJobQuery) (int64, []*models.NotificationJob, error) {
	total, err := notification.GetTotalCountOfNotificationJobs(query...)
	if err != nil {
		return 0, nil, err
	}

	executions, err := notification.GetNotificationJobs(query...)
	if err != nil {
		return 0, nil, err
	}
	return total, executions, nil
}

// Update ...
func (d *DefaultManager) Update(job *models.NotificationJob, props ...string) error {
	n, err := notification.UpdateNotificationJob(job, props...)
	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("execution %d not found", job.ID)
	}
	return nil
}

// UpdateJobStatus ...
func (d *DefaultManager) UpdateJobStatus(jobID int64, status string, statusCondition ...string) error {
	n, err := notification.UpdateNotificationJobStatus(jobID, status, statusCondition...)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("Update notification job %d status -> %s failed ", jobID, status)
	}
	return nil
}

// ListJobsGroupByEventType lists last triggered jobs group by event type
func (d *DefaultManager) ListJobsGroupByEventType(policyID int64) ([]*models.NotificationJob, error) {
	return notification.GetLastTriggerJobsGroupByEventType(policyID)
}
