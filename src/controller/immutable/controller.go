package immutable

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/immutable"

	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

var (
	// Ctr is a global variable for the default immutable controller implementation
	Ctr = NewAPIController(immutable.NewDefaultRuleManager())
)

// Controller to handle the requests related with immutable
type Controller interface {
	// GetImmutableRule ...
	GetImmutableRule(ctx context.Context, id int64) (*model.Metadata, error)

	// CreateImmutableRule ...
	CreateImmutableRule(ctx context.Context, m *model.Metadata) (int64, error)

	// DeleteImmutableRule ...
	DeleteImmutableRule(ctx context.Context, id int64) error

	// UpdateImmutableRule ...
	UpdateImmutableRule(ctx context.Context, projectID int64, m *model.Metadata) error

	// ListImmutableRules ...
	ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.Metadata, error)

	// Count count the immutable rules
	Count(ctx context.Context, query *q.Query) (int64, error)
}

// DefaultAPIController ...
type DefaultAPIController struct {
	manager immutable.Manager
}

// GetImmutableRule ...
func (r *DefaultAPIController) GetImmutableRule(ctx context.Context, id int64) (*model.Metadata, error) {
	return r.manager.GetImmutableRule(ctx, id)
}

// DeleteImmutableRule ...
func (r *DefaultAPIController) DeleteImmutableRule(ctx context.Context, id int64) error {
	return r.manager.DeleteImmutableRule(ctx, id)
}

// CreateImmutableRule ...
func (r *DefaultAPIController) CreateImmutableRule(ctx context.Context, m *model.Metadata) (int64, error) {
	return r.manager.CreateImmutableRule(ctx, m)
}

// UpdateImmutableRule ...
func (r *DefaultAPIController) UpdateImmutableRule(ctx context.Context, projectID int64, m *model.Metadata) error {
	m0, err := r.manager.GetImmutableRule(ctx, m.ID)
	if err != nil {
		return err
	}
	if m0 == nil {
		return fmt.Errorf("the immutable tag rule is not found id:%v", m.ID)
	}
	if m0.Disabled != m.Disabled {
		return r.manager.EnableImmutableRule(ctx, m.ID, m.Disabled)
	}
	return r.manager.UpdateImmutableRule(ctx, projectID, m)
}

// ListImmutableRules ...
func (r *DefaultAPIController) ListImmutableRules(ctx context.Context, query *q.Query) ([]*model.Metadata, error) {
	return r.manager.ListImmutableRules(ctx, query)
}

// Count count the immutable rules
func (r *DefaultAPIController) Count(ctx context.Context, query *q.Query) (int64, error) {
	return r.manager.Count(ctx, query)
}

// NewAPIController ...
func NewAPIController(immutableMgr immutable.Manager) Controller {
	return &DefaultAPIController{
		manager: immutableMgr,
	}
}
