package job

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Manager manages webhook jobs
type Manager interface {
	// Create create a webhook job
	Create(job *models.WebhookJob) (int64, error)

	// List list webhook jobs
	List(...*models.WebhookJobQuery) (int64, []*models.WebhookJob, error)

	// Update update webhook job
	Update(job *models.WebhookJob, props ...string) error

	// UpdateJobStatus update webhook job status
	UpdateJobStatus(jobID int64, status string, statusCondition ...string) error

	// ListLastTriggerInfos list jobs info including hook type and last trigger time
	ListLastTriggerInfos(policyID int64) ([]*models.LastTriggerInfo, error)
}
