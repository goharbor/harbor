package policy

import (
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/webhook/policy/manager"
)

// Controller handles the webhook policy related operations
type Controller interface {
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

type controller struct {
	policyMgr Manager
}

// NewController returns a controller implementation
func NewController() Controller {
	ctl := &controller{
		policyMgr: manager.NewDefaultManger(),
	}
	return ctl
}

// Create new policy
func (c *controller) Create(policy *models.WebhookPolicy) (int64, error) {
	return c.policyMgr.Create(policy)
}

// List the policies, returns the total count, policy list and error
func (c *controller) List(projectID int64) (int64, []*models.WebhookPolicy, error) {
	return c.policyMgr.List(projectID)
}

// Get policy with specified ID
func (c *controller) Get(id int64) (*models.WebhookPolicy, error) {
	return c.policyMgr.Get(id)
}

// GetByNameAndProjectID get policy by the name and projectID
func (c *controller) GetByNameAndProjectID(name string, projectID int64) (*models.WebhookPolicy, error) {
	return c.policyMgr.GetByNameAndProjectID(name, projectID)
}

// Update the specified policy
func (c *controller) Update(policy *models.WebhookPolicy) error {
	return c.policyMgr.Update(policy)
}

// Delete the specified policy
func (c *controller) Delete(id int64) error {
	return c.policyMgr.Delete(id)
}

// Test the specified policy
func (c *controller) Test(policy *models.WebhookPolicy) error {
	return c.policyMgr.Test(policy)
}

// GetRelatedPolices get hook type related policies in project
func (c *controller) GetRelatedPolices(projectID int64, hookType string) ([]*models.WebhookPolicy, error) {
	return c.policyMgr.GetRelatedPolices(projectID, hookType)
}
