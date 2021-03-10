package model

import (
	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/notification/job/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
)

// NotificationJob ...
type NotificationJob struct {
	*model.Job
}

// ToSwagger ...
func (n *NotificationJob) ToSwagger() *models.WebhookJob {
	return &models.WebhookJob{
		ID:           n.ID,
		EventType:    n.EventType,
		JobDetail:    n.JobDetail,
		NotifyType:   n.NotifyType,
		PolicyID:     n.PolicyID,
		Status:       n.Status,
		CreationTime: strfmt.DateTime(n.CreationTime),
		UpdateTime:   strfmt.DateTime(n.UpdateTime),
	}
}

// NewNotificationJob ...
func NewNotificationJob(j *model.Job) *NotificationJob {
	return &NotificationJob{
		Job: j,
	}
}
