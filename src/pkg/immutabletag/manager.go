package immutabletag

import (
	"github.com/goharbor/harbor/src/pkg/immutabletag/dao"
	"github.com/goharbor/harbor/src/pkg/immutabletag/dao/model"
)

var (
	// Mgr is a global variable for the default immutablerule manager implementation
	Mgr = NewDefaultRuleManager()
)

// Manager ...
type Manager interface {
	// CreateImmutableRule creates the Immutable Rule
	CreateImmutableRule(ir *model.ImmutableRule) (int64, error)
	// UpdateImmutableRule update the immutable rules
	UpdateImmutableRule(projectID int64, ir *model.ImmutableRule) (int64, error)
	// EnableImmutableRule enable/disable immutable rules
	EnableImmutableRule(id int64, enabled bool) (int64, error)
	// GetImmutableRule get immutable rule
	GetImmutableRule(id int64) (*model.ImmutableRule, error)
	// QueryImmutableRuleByProjectID get all immutable rule by project
	QueryImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error)
	// QueryEnabledImmutableRuleByProjectID get all enabled immutable rule by project
	QueryEnabledImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error)
	// DeleteImmutableRule delete the immutable rule
	DeleteImmutableRule(id int64) (int64, error)
}

type defaultRuleManager struct {
	dao dao.ImmutableRuleDao
}

func (drm *defaultRuleManager) CreateImmutableRule(ir *model.ImmutableRule) (int64, error) {
	return drm.dao.CreateImmutableRule(ir)
}

func (drm *defaultRuleManager) UpdateImmutableRule(projectID int64, ir *model.ImmutableRule) (int64, error) {
	return drm.dao.UpdateImmutableRule(projectID, ir)
}

func (drm *defaultRuleManager) EnableImmutableRule(id int64, enabled bool) (int64, error) {
	return drm.dao.ToggleImmutableRule(id, enabled)
}

func (drm *defaultRuleManager) GetImmutableRule(id int64) (*model.ImmutableRule, error) {
	return drm.dao.GetImmutableRule(id)
}

func (drm *defaultRuleManager) QueryImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error) {
	return drm.dao.QueryImmutableRuleByProjectID(projectID)
}

func (drm *defaultRuleManager) QueryEnabledImmutableRuleByProjectID(projectID int64) ([]model.ImmutableRule, error) {
	return drm.dao.QueryEnabledImmutableRuleByProjectID(projectID)
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
