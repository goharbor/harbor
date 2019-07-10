package impl

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	persistModels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/webhook/model"
)

var testPayload = model.Payload{
	Type:      model.EventTypeTestPolicy,
	OccurAt:   time.Now().Unix(),
	MediaType: "containerImage",
	EventData: []*model.EventData{
		{
			Digest:      "sha256:457f4aa83fc9a6663ab9d1b0a6e2dce25a12a943ed5bf2c1747c58d48bbb4917",
			Tag:         "testTag",
			ResourceURL: "repo.harbor.com/testnamespace/repoTest:testTag",
		},
	},
	Repository: &model.Repository{
		DateCreated:  time.Now().Unix(),
		Name:         "repoTest",
		RepoFullName: "testnamespace/repoTest",
		Namespace:    "testnamespace",
		RepoType:     "public",
	},
	Operator: "test",
}

// DefaultManager ...
type DefaultManager struct {
	client *http.Client
}

// NewDefaultManger ...
func NewDefaultManger() *DefaultManager {
	return &DefaultManager{
		client: &http.Client{},
	}
}

// Create webhook policy
func (m *DefaultManager) Create(policy *model.WebhookPolicy) (int64, error) {
	t := time.Now()
	policy.CreationTime = t
	policy.UpdateTime = t

	ply, err := convertToPersistModel(policy)
	if err != nil {
		return 0, err
	}
	return dao.AddWebhookPolicy(ply)
}

// List the webhook policies, returns the total count, policy list and error
func (m *DefaultManager) List(projectID int64) (int64, []*model.WebhookPolicy, error) {
	var policies []*model.WebhookPolicy
	total, persisPolicies, err := dao.GetWebhookPolicies(projectID)
	if err != nil {
		return total, nil, err
	}

	for _, policy := range persisPolicies {
		ply, err := convertFromPersistModel(policy)
		if err != nil {
			return 0, nil, err
		}
		policies = append(policies, ply)
	}

	if policies == nil {
		policies = []*model.WebhookPolicy{}
	}
	return total, policies, nil
}

// Get webhook policy with specified ID
func (m *DefaultManager) Get(id int64) (*model.WebhookPolicy, error) {
	policy, err := dao.GetWebhookPolicy(id)
	if err != nil {
		return nil, err
	}
	return convertFromPersistModel(policy)
}

// GetByNameAndProjectID webhook policy by the name and projectID
func (m *DefaultManager) GetByNameAndProjectID(name string, projectID int64) (*model.WebhookPolicy, error) {
	policy, err := dao.GetWebhookPolicyByName(name, projectID)
	if err != nil {
		return nil, err
	}
	return convertFromPersistModel(policy)
}

// Update the specified webhook policy
func (m *DefaultManager) Update(policy *model.WebhookPolicy) error {
	policy.UpdateTime = time.Now()
	ply, err := convertToPersistModel(policy)
	if err != nil {
		return err
	}
	return dao.UpdateWebhookPolicy(ply)
}

// Delete the specified webhook policy
func (m *DefaultManager) Delete(policyID int64) error {
	return dao.DeleteWebhookPolicy(policyID)
}

// Test the specified webhook policy
func (m *DefaultManager) Test(policy *model.WebhookPolicy) error {
	p, err := json.Marshal(testPayload)
	if err != nil {
		return err
	}

	for _, target := range policy.Targets {
		switch target.Type {
		case "http":
			return m.policyHTTPTest(target.Address, target.Secret, p)
		default:
			return fmt.Errorf("invalid policy target type: %s", target.Type)
		}
	}
	return nil
}

func (m *DefaultManager) policyHTTPTest(address, secret string, payload []byte) error {
	p := bytes.NewReader(payload)
	req, err := http.NewRequest(http.MethodPost, address, p)
	if err != nil {
		return err
	}

	if secret != "" {
		req.Header.Set("Authorization", "Secret "+secret)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 300 {
		return fmt.Errorf("policy test failed with response code %d", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("read policy test response body failed: %v", err)
	}
	log.Debugf("policy test response code %d, body: %s", resp.StatusCode, string(body))

	return nil
}

func convertToPersistModel(policy *model.WebhookPolicy) (*persistModels.WebhookPolicy, error) {
	if policy == nil {
		return nil, errors.New("nil webhook policy model")
	}

	ply := &persistModels.WebhookPolicy{
		ID:           policy.ID,
		Name:         policy.Name,
		Description:  policy.Description,
		ProjectID:    policy.ProjectID,
		Creator:      policy.Creator,
		CreationTime: policy.CreationTime,
		UpdateTime:   policy.UpdateTime,
		Enabled:      policy.Enabled,
	}

	if len(policy.Targets) != 0 {
		targets, err := json.Marshal(policy.Targets)
		if err != nil {
			return nil, err
		}
		ply.Targets = string(targets)
	}

	if len(policy.HookTypes) > 0 {
		hookTypes, err := json.Marshal(policy.HookTypes)
		if err != nil {
			return nil, err
		}
		ply.HookTypes = string(hookTypes)
	}
	return ply, nil
}

func convertFromPersistModel(policy *persistModels.WebhookPolicy) (*model.WebhookPolicy, error) {
	if policy == nil {
		return nil, nil
	}

	ply := model.WebhookPolicy{
		ID:           policy.ID,
		Name:         policy.Name,
		Description:  policy.Description,
		ProjectID:    policy.ProjectID,
		Creator:      policy.Creator,
		CreationTime: policy.CreationTime,
		UpdateTime:   policy.UpdateTime,
		Enabled:      policy.Enabled,
	}

	hookTypes, err := parseHookTypes(policy.HookTypes)
	if err != nil {
		return nil, err
	}
	ply.HookTypes = hookTypes

	targets, err := parseTargets(policy.Targets)
	if err != nil {
		return nil, err
	}
	ply.Targets = targets
	return &ply, nil
}

func parseTargets(targets string) ([]model.HookTarget, error) {
	if len(targets) == 0 {
		return nil, nil
	}
	var t []model.HookTarget
	if err := json.Unmarshal([]byte(targets), &t); err != nil {
		return nil, err
	}

	return t, nil
}

func parseHookTypes(hookTypes string) ([]string, error) {
	if len(hookTypes) == 0 {
		return nil, nil
	}
	var h []string
	if err := json.Unmarshal([]byte(hookTypes), &h); err != nil {
		return nil, err
	}
	return h, nil
}
