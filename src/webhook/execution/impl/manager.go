package impl

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/webhook/execution"
)

// DefaultManager ..
type DefaultManager struct {
}

// NewDefaultManager ...
func NewDefaultManager() execution.Manager {
	return &DefaultManager{}
}

// Create ...
func (d *DefaultManager) Create(execution *models.WebhookExecution) (int64, error) {
	return dao.AddWebhookExecution(execution)
}

// Get ...
func (d *DefaultManager) Get(id int64) (*models.WebhookExecution, error) {
	return dao.GetWebhookExecution(id)
}

// List ...
func (d *DefaultManager) List(query ...*models.WebhookExecutionQuery) (int64, []*models.WebhookExecution, error) {
	total, err := dao.GetTotalCountOfWebhookExecutions(query...)
	if err != nil {
		return 0, nil, err
	}

	executions, err := dao.GetWebhookExecutions(query...)
	if err != nil {
		return 0, nil, err
	}
	return total, executions, nil
}

// Update ...
func (d *DefaultManager) Update(execution *models.WebhookExecution, props ...string) error {
	n, err := dao.UpdateWebhookExecution(execution, props...)
	if err != nil {
		return err
	}

	if n == 0 {
		return fmt.Errorf("execution %d not found", execution.ID)
	}
	return nil
}

// Delete ...
func (d *DefaultManager) Delete(id int64) error {
	return dao.DeleteWebhookExecution(id)
}
