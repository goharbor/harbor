package model

import (
	"github.com/go-openapi/strfmt"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// WebhookJob ...
type WebhookJob struct {
	*task.Execution
}

// ToSwagger ...
func (n *WebhookJob) ToSwagger() *models.WebhookJob {
	webhookJob := &models.WebhookJob{
		ID:           n.ID,
		PolicyID:     n.VendorID,
		Status:       n.Status,
		CreationTime: strfmt.DateTime(n.StartTime),
		UpdateTime:   strfmt.DateTime(n.UpdateTime),
	}

	var notifyType string
	// do the conversion for compatible with old API
	if n.VendorType == job.WebhookJobVendorType {
		notifyType = "http"
	} else if n.VendorType == job.SlackJobVendorType {
		notifyType = "slack"
	}
	webhookJob.NotifyType = notifyType

	if n.ExtraAttrs != nil {
		if eventType, ok := n.ExtraAttrs["event_type"].(string); ok {
			webhookJob.EventType = eventType
		}

		if payload, ok := n.ExtraAttrs["payload"].(string); ok {
			webhookJob.JobDetail = payload
		}
	}

	return webhookJob
}

// NewWebhookJob ...
func NewWebhookJob(exec *task.Execution) *WebhookJob {
	return &WebhookJob{
		Execution: exec,
	}
}
