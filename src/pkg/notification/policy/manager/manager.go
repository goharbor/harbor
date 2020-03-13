package manager

import (
	"fmt"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/dao/notification"
	commonhttp "github.com/goharbor/harbor/src/common/http"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// DefaultManager ...
type DefaultManager struct {
}

// NewDefaultManger ...
func NewDefaultManger() *DefaultManager {
	return &DefaultManager{}
}

// Create notification policy
func (m *DefaultManager) Create(policy *models.NotificationPolicy) (int64, error) {
	t := time.Now()
	policy.CreationTime = t
	policy.UpdateTime = t

	err := policy.ConvertToDBModel()
	if err != nil {
		return 0, err
	}
	return notification.AddNotificationPolicy(policy)
}

// List the notification policies, returns the policy list and error
func (m *DefaultManager) List(projectID int64) ([]*models.NotificationPolicy, error) {
	policies := []*models.NotificationPolicy{}
	persisPolicies, err := notification.GetNotificationPolicies(projectID)
	if err != nil {
		return nil, err
	}

	for _, policy := range persisPolicies {
		err := policy.ConvertFromDBModel()
		if err != nil {
			return nil, err
		}
		policies = append(policies, policy)
	}

	return policies, nil
}

// Get notification policy with specified ID
func (m *DefaultManager) Get(id int64) (*models.NotificationPolicy, error) {
	policy, err := notification.GetNotificationPolicy(id)
	if err != nil {
		return nil, err
	}
	if policy == nil {
		return nil, nil
	}
	err = policy.ConvertFromDBModel()
	return policy, err
}

// GetByNameAndProjectID notification policy by the name and projectID
func (m *DefaultManager) GetByNameAndProjectID(name string, projectID int64) (*models.NotificationPolicy, error) {
	policy, err := notification.GetNotificationPolicyByName(name, projectID)
	if err != nil {
		return nil, err
	}
	err = policy.ConvertFromDBModel()
	return policy, err
}

// Update the specified notification policy
func (m *DefaultManager) Update(policy *models.NotificationPolicy) error {
	policy.UpdateTime = time.Now()
	err := policy.ConvertToDBModel()
	if err != nil {
		return err
	}
	return notification.UpdateNotificationPolicy(policy)
}

// Delete the specified notification policy
func (m *DefaultManager) Delete(policyID int64) error {
	return notification.DeleteNotificationPolicy(policyID)
}

// Test the specified notification policy, just test for network connection without request body
func (m *DefaultManager) Test(policy *models.NotificationPolicy) error {
	for _, target := range policy.Targets {
		switch target.Type {
		case "http":
			return m.policyHTTPTest(target.Address, target.SkipCertVerify)
		default:
			return fmt.Errorf("invalid policy target type: %s", target.Type)
		}
	}
	return nil
}

func (m *DefaultManager) policyHTTPTest(address string, skipCertVerify bool) error {
	req, err := http.NewRequest(http.MethodPost, address, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := http.Client{
		Transport: commonhttp.GetHTTPTransport(skipCertVerify),
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	log.Debugf("policy test success with address %s, skip cert verify :%v", address, skipCertVerify)

	return nil
}

// GetRelatedPolices get policies including event type in project
func (m *DefaultManager) GetRelatedPolices(projectID int64, eventType string) ([]*models.NotificationPolicy, error) {
	policies, err := m.List(projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to get notification policies with projectID %d: %v", projectID, err)
	}

	var result []*models.NotificationPolicy

	for _, ply := range policies {
		if !ply.Enabled {
			continue
		}
		for _, t := range ply.EventTypes {
			if t != eventType {
				continue
			}
			result = append(result, ply)
		}
	}
	return result, nil
}
