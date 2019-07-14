package operation

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/webhook/execution"
	"github.com/goharbor/harbor/src/webhook/execution/manager"
)

// Controller handles the webhook related operations
type Controller interface {
	// GetWebhookExecution get a webhook execution
	GetWebhookExecution(int64) (*models.WebhookExecution, error)

	// ListWebhookExecutions list webhook executions
	ListWebhookExecutions(...*models.WebhookExecutionQuery) (int64, []*models.WebhookExecution, error)

	// UpdateWebhookExecution update webhook execution
	UpdateWebhookExecution(*models.WebhookExecution, ...string) error

	// DeleteWebhookExecution delete webhook execution
	DeleteWebhookExecution(int64) error

	// ListLastTriggerInfos list executions info including hook type and last trigger time
	ListLastTriggerInfos() ([]*models.LastTriggerInfo, error)
}

type controller struct {
	execMgr execution.Manager
}

// NewController returns a controller implementation
func NewController() Controller {
	ctl := &controller{
		execMgr: manager.NewDefaultManager(),
	}
	return ctl
}

// GetWebhookExecution ...
func (c *controller) GetWebhookExecution(id int64) (*models.WebhookExecution, error) {
	return c.execMgr.Get(id)
}

// ListWebhookExecutions ...
func (c *controller) ListWebhookExecutions(query ...*models.WebhookExecutionQuery) (int64, []*models.WebhookExecution, error) {
	return c.execMgr.List(query...)
}

// UpdateWebhookExecution ...
func (c *controller) UpdateWebhookExecution(execution *models.WebhookExecution, props ...string) error {
	return c.execMgr.Update(execution, props...)
}

// DeleteWebhookExecution ...
func (c *controller) DeleteWebhookExecution(id int64) error {
	return c.execMgr.Delete(id)
}

// ListLastTriggerInfos ...
func (c *controller) ListLastTriggerInfos() ([]*models.LastTriggerInfo, error) {
	return c.execMgr.ListLastTriggerInfos()
}
