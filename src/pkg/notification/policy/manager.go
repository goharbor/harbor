package policy

import (
	"github.com/goharbor/harbor/src/common/models"
)

// Manager manages the notification policies
type Manager interface {
	// Create new policy
	Create(*models.NotificationPolicy) (int64, error)
	// List the policies, returns the policy list and error
	List(int64) ([]*models.NotificationPolicy, error)
	// Get policy with specified ID
	Get(int64) (*models.NotificationPolicy, error)
	// GetByNameAndProjectID get policy by the name and projectID
	GetByNameAndProjectID(string, int64) (*models.NotificationPolicy, error)
	// Update the specified policy
	Update(*models.NotificationPolicy) error
	// Delete the specified policy
	Delete(int64) error
	// Test the specified policy
	Test(*models.NotificationPolicy) error
	// GetRelatedPolices get event type related policies in project
	GetRelatedPolices(int64, string) ([]*models.NotificationPolicy, error)
}
