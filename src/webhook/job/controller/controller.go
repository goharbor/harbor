package controller

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/webhook/job"
	"github.com/goharbor/harbor/src/webhook/job/manager"
)

type controller struct {
	jobMgr job.Manager
}

// NewController returns a controller implementation
func NewController() job.Controller {
	ctl := &controller{
		jobMgr: manager.NewDefaultManager(),
	}
	return ctl
}

// ListWebhookJobs ...
func (c *controller) ListWebhookJobs(query ...*models.WebhookJobQuery) (int64, []*models.WebhookJob, error) {
	return c.jobMgr.List(query...)
}

// UpdateWebhookJob ...
func (c *controller) UpdateWebhookJob(execution *models.WebhookJob, props ...string) error {
	return c.jobMgr.Update(execution, props...)
}

// ListLastTriggerInfos list webhook infos including hook type and last trigger time
func (c *controller) ListLastTriggerInfos(policyID int64) ([]*models.LastTriggerInfo, error) {
	return c.jobMgr.ListLastTriggerInfos(policyID)
}
