package execution

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Manager manages webhook executions
type Manager interface {
	// Create create a webhook execution
	Create(execution *models.WebhookExecution) (int64, error)

	// Get get a webhook execution
	Get(int64) (*models.WebhookExecution, error)

	// List list webhook executions
	List(...*models.WebhookExecutionQuery) (int64, []*models.WebhookExecution, error)

	// Update the data of the specified webhook execution, the "props" are the
	//	// properties of execution that need to be updated
	Update(execution *models.WebhookExecution, props ...string) error

	// Delete delete webhook execution
	Delete(int64) error
}
