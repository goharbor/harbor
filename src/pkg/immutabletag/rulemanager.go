package immutabletag

import (
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
)

// RuleManager ...
type RuleManager interface {
	// CreateImmutableRule creates the Immutable Rule
	CreateImmutableRule(ir *models.ImmutableRule) (int64, error)
	// UpdateImmutableRule update the immutable rules
	UpdateImmutableRule(projectID int64, ir *models.ImmutableRule) (int64, error)
	// EnableImmutableRule enable/disable immutable rules
	EnableImmutableRule(id int64, enabled bool) (int64, error)
	// GetImmutableRule get immutable rule
	GetImmutableRule(id int64) (*models.ImmutableRule, error)
	// QueryImmutableRuleByProjectID get all immutable rule by project
	QueryImmutableRuleByProjectID(projectID int64) ([]models.ImmutableRule, error)
	// QueryEnabledImmutableRuleByProjectID get all enabled immutable rule by project
	QueryEnabledImmutableRuleByProjectID(projectID int64) ([]models.ImmutableRule, error)
	// DeleteImmutableRule delete the immutable rule
	DeleteImmutableRule(id int64) (int64, error)
}

type defaultRuleManager struct{}

func (drm *defaultRuleManager) CreateImmutableRule(ir *models.ImmutableRule) (int64, error) {
	return dao.CreateImmutableRule(ir)
}

func (drm *defaultRuleManager) UpdateImmutableRule(projectID int64, ir *models.ImmutableRule) (int64, error) {
	return dao.UpdateImmutableRule(projectID, ir)
}

func (drm *defaultRuleManager) EnableImmutableRule(id int64, enabled bool) (int64, error) {
	return dao.ToggleImmutableRule(id, enabled)
}

func (drm *defaultRuleManager) GetImmutableRule(id int64) (*models.ImmutableRule, error) {
	return dao.GetImmutableRule(id)
}

func (drm *defaultRuleManager) QueryImmutableRuleByProjectID(projectID int64) ([]models.ImmutableRule, error) {
	return dao.QueryImmutableRuleByProjectID(projectID)
}

func (drm *defaultRuleManager) QueryEnabledImmutableRuleByProjectID(projectID int64) ([]models.ImmutableRule, error) {
	return dao.QueryEnabledImmutableRuleByProjectID(projectID)
}

func (drm *defaultRuleManager) DeleteImmutableRule(id int64) (int64, error) {
	return dao.DeleteImmutableRule(id)
}

// NewDefaultRuleManager return a new instance of defaultRuleManager
func NewDefaultRuleManager() RuleManager {
	return &defaultRuleManager{}
}
