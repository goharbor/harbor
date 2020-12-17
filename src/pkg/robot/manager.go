package robot

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/robot/dao"
	"github.com/goharbor/harbor/src/pkg/robot/model"
)

var (
	// Mgr is a global variable for the default robot account manager implementation
	Mgr = NewManager()
)

// Manager ...
type Manager interface {
	// Get ...
	Get(ctx context.Context, id int64) (*model.Robot, error)

	// Count returns the total count of robots according to the query
	Count(ctx context.Context, query *q.Query) (total int64, err error)

	// Create ...
	Create(ctx context.Context, m *model.Robot) (int64, error)

	// Delete ...
	Delete(ctx context.Context, id int64) error

	// DeleteByProjectID ...
	DeleteByProjectID(ctx context.Context, projectID int64) error

	// Update ...
	Update(ctx context.Context, m *model.Robot, props ...string) error

	// List ...
	List(ctx context.Context, query *q.Query) ([]*model.Robot, error)
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

// NewManager return a new instance of defaultRobotManager
func NewManager() Manager {
	return &manager{
		dao: dao.New(),
	}
}

// Get ...
func (m *manager) Get(ctx context.Context, id int64) (*model.Robot, error) {
	return m.dao.Get(ctx, id)
}

// Count ...
func (m *manager) Count(ctx context.Context, query *q.Query) (total int64, err error) {
	return m.dao.Count(ctx, query)
}

// Create ...
func (m *manager) Create(ctx context.Context, r *model.Robot) (int64, error) {
	return m.dao.Create(ctx, r)
}

// Delete ...
func (m *manager) Delete(ctx context.Context, id int64) error {
	return m.dao.Delete(ctx, id)
}

// DeleteByProjectID ...
func (m *manager) DeleteByProjectID(ctx context.Context, projectID int64) error {
	return m.dao.DeleteByProjectID(ctx, projectID)
}

// Update ...
func (m *manager) Update(ctx context.Context, r *model.Robot, props ...string) error {
	return m.dao.Update(ctx, r, props...)
}

// List ...
func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.Robot, error) {
	return m.dao.List(ctx, query)
}
