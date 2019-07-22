package policy

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Manager manages the webhook policies
type Manager interface {
	// Create new policy
	Create(*models.WebhookPolicy) (int64, error)
	// List the policies, returns the total count, policy list and error
	List(int64) (int64, []*models.WebhookPolicy, error)
	// Get policy with specified ID
	Get(int64) (*models.WebhookPolicy, error)
	// GetByNameAndProjectID get policy by the name and projectID
	GetByNameAndProjectID(string, int64) (*models.WebhookPolicy, error)
	// Update the specified policy
	Update(*models.WebhookPolicy) error
	// Delete the specified policy
	Delete(int64) error
	// Test the specified policy
	Test(*models.WebhookPolicy) error
	// GetRelatedPolices get hook type related policies in project
	GetRelatedPolices(int64, string) ([]*models.WebhookPolicy, error)
}
