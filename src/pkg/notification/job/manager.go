package job

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Manager manages notification jobs recorded in database
type Manager interface {
	// Create create a notification job
	Create(job *models.NotificationJob) (int64, error)

	// List list notification jobs
	List(...*models.NotificationJobQuery) (int64, []*models.NotificationJob, error)

	// Update update notification job
	Update(job *models.NotificationJob, props ...string) error

	// UpdateJobStatus update notification job status
	UpdateJobStatus(jobID int64, status string, statusCondition ...string) error

	// ListJobsGroupByEventType lists last triggered jobs group by event type
	ListJobsGroupByEventType(policyID int64) ([]*models.NotificationJob, error)
}
