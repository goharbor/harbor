package policy

import "github.com/goharbor/harbor/src/webhook/model"

// Manager manages the webhook policies
type Manager interface {
	// Create new policy
	Create(*model.WebhookPolicy) (int64, error)
	// List the policies, returns the total count, policy list and error
	List(int64) (int64, []*model.WebhookPolicy, error)
	// Get policy with specified ID
	Get(int64) (*model.WebhookPolicy, error)
	// GetByNameAndProjectID get policy by the name and projectID
	GetByNameAndProjectID(string, int64) (*model.WebhookPolicy, error)
	// Update the specified policy
	Update(*model.WebhookPolicy) error
	// Delete the specified policy
	Delete(int64) error
	// Test the specified policy
	Test(*model.WebhookPolicy) error
}
