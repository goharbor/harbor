package job

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Controller handles the webhook job related operations
type Controller interface {
	// ListWebhookJobs list webhook jobs
	ListWebhookJobs(...*models.WebhookJobQuery) (int64, []*models.WebhookJob, error)

	// WebhookJob update webhook job
	UpdateWebhookJob(*models.WebhookJob, ...string) error

	// ListLastTriggerInfos list jobs info including hook type and last trigger time
	ListLastTriggerInfos(policyID int64) ([]*models.LastTriggerInfo, error)
}
