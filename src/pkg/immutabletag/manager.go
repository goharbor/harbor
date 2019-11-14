package immutabletag

import (
	"encoding/json"
	"sort"

	"github.com/goharbor/harbor/src/pkg/immutabletag/dao"
	dao_model "github.com/goharbor/harbor/src/pkg/immutabletag/dao/model"
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
)

var (
	// Mgr is a global variable for the default immutablerule manager implementation
	Mgr = NewDefaultRuleManager()
)

// Manager ...
type Manager interface {
	// CreateImmutableRule creates the Immutable Rule
	CreateImmutableRule(m *model.Metadata) (int64, error)
	// UpdateImmutableRule update the immutable rules
	UpdateImmutableRule(projectID int64, ir *model.Metadata) (int64, error)
	// EnableImmutableRule enable/disable immutable rules
	EnableImmutableRule(id int64, enabled bool) (int64, error)
	// GetImmutableRule get immutable rule
	GetImmutableRule(id int64) (*model.Metadata, error)
	// QueryImmutableRuleByProjectID get all immutable rule by project
	QueryImmutableRuleByProjectID(projectID int64) ([]model.Metadata, error)
	// QueryEnabledImmutableRuleByProjectID get all enabled immutable rule by project
	QueryEnabledImmutableRuleByProjectID(projectID int64) ([]model.Metadata, error)
	// DeleteImmutableRule delete the immutable rule
	DeleteImmutableRule(id int64) (int64, error)
}

type defaultRuleManager struct {
	dao dao.ImmutableRuleDao
}

func (drm *defaultRuleManager) CreateImmutableRule(ir *model.Metadata) (int64, error) {
	daoRule := &dao_model.ImmutableRule{}
	daoRule.Disabled = ir.Disabled
	daoRule.ProjectID = ir.ProjectID
	data, _ := json.Marshal(ir)
	daoRule.TagFilter = string(data)
	return drm.dao.CreateImmutableRule(daoRule)
}

func (drm *defaultRuleManager) UpdateImmutableRule(projectID int64, ir *model.Metadata) (int64, error) {
	daoRule := &dao_model.ImmutableRule{}
	data, _ := json.Marshal(ir)
	daoRule.ID = ir.ID
	daoRule.TagFilter = string(data)
	return drm.dao.UpdateImmutableRule(projectID, daoRule)
}

func (drm *defaultRuleManager) EnableImmutableRule(id int64, enabled bool) (int64, error) {
	return drm.dao.ToggleImmutableRule(id, enabled)
}

func (drm *defaultRuleManager) GetImmutableRule(id int64) (*model.Metadata, error) {
	daoRule, err := drm.dao.GetImmutableRule(id)
	if err != nil {
		return nil, err
	}
	rule := &model.Metadata{}
	if daoRule == nil {
		return nil, nil
	}
	if err = json.Unmarshal([]byte(daoRule.TagFilter), rule); err != nil {
		return nil, err
	}
	rule.ID = daoRule.ID
	rule.Disabled = daoRule.Disabled
	return rule, nil
}

func (drm *defaultRuleManager) QueryImmutableRuleByProjectID(projectID int64) ([]model.Metadata, error) {
	daoRules, err := drm.dao.QueryImmutableRuleByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	var rules []model.Metadata
	for _, daoRule := range daoRules {
		rule := model.Metadata{}
		if err = json.Unmarshal([]byte(daoRule.TagFilter), &rule); err != nil {
			return nil, err
		}
		rule.ID = daoRule.ID
		rule.Disabled = daoRule.Disabled
		rules = append(rules, rule)
	}
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].ID < rules[j].ID
	})
	return rules, nil
}

func (drm *defaultRuleManager) QueryEnabledImmutableRuleByProjectID(projectID int64) ([]model.Metadata, error) {
	daoRules, err := drm.dao.QueryEnabledImmutableRuleByProjectID(projectID)
	if err != nil {
		return nil, err
	}
	var rules []model.Metadata
	for _, daoRule := range daoRules {
		rule := model.Metadata{}
		if err = json.Unmarshal([]byte(daoRule.TagFilter), &rule); err != nil {
			return nil, err
		}
		rule.ID = daoRule.ID
		rules = append(rules, rule)
	}
	return rules, nil
}

func (drm *defaultRuleManager) DeleteImmutableRule(id int64) (int64, error) {
	return drm.dao.DeleteImmutableRule(id)
}

// NewDefaultRuleManager return a new instance of defaultRuleManager
func NewDefaultRuleManager() Manager {
	return &defaultRuleManager{
		dao: dao.New(),
	}
}
