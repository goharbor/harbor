package immutabletag

import (
	"github.com/goharbor/harbor/src/pkg/immutabletag/model"
)

// APIController to handle the requests related with immutabletag
type APIController interface {
	// GetImmutableRule ...
	GetImmutableRule(id int64) (*model.Metadata, error)

	// CreateImmutableRule ...
	CreateImmutableRule(m *model.Metadata) (int64, error)

	// DeleteImmutableRule ...
	DeleteImmutableRule(id int64) error

	// UpdateImmutableRule ...
	UpdateImmutableRule(pid int64, m *model.Metadata) error

	// ListImmutableRules ...
	ListImmutableRules(pid int64) ([]model.Metadata, error)
}

// DefaultAPIController ...
type DefaultAPIController struct {
	manager Manager
}

// GetImmutableRule ...
func (r *DefaultAPIController) GetImmutableRule(id int64) (*model.Metadata, error) {
	return r.manager.GetImmutableRule(id)
}

// DeleteImmutableRule ...
func (r *DefaultAPIController) DeleteImmutableRule(id int64) error {
	_, err := r.manager.DeleteImmutableRule(id)
	return err
}

// CreateImmutableRule ...
func (r *DefaultAPIController) CreateImmutableRule(m *model.Metadata) (int64, error) {
	return r.manager.CreateImmutableRule(m)
}

// UpdateImmutableRule ...
func (r *DefaultAPIController) UpdateImmutableRule(pid int64, m *model.Metadata) error {
	m0, err := r.manager.GetImmutableRule(m.ID)
	if err != nil {
		return err
	}
	if m0.Disabled != m.Disabled {
		_, err := r.manager.EnableImmutableRule(m.ID, m.Disabled)
		return err
	}
	_, err = r.manager.UpdateImmutableRule(pid, m)
	return err
}

// ListImmutableRules ...
func (r *DefaultAPIController) ListImmutableRules(pid int64) ([]model.Metadata, error) {
	return r.manager.QueryImmutableRuleByProjectID(pid)
}

// NewAPIController ...
func NewAPIController(immutableMgr Manager) APIController {
	return &DefaultAPIController{
		manager: immutableMgr,
	}
}
