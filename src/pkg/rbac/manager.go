package rbac

import (
	"context"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/rbac/dao"
	"github.com/goharbor/harbor/src/pkg/rbac/model"
)

var (
	// Mgr is a global role permission/rbac policy manager instance
	Mgr = NewManager()
)

// Manager is the interface of role permission and rbac policy
type Manager interface {
	// CreatePermission ...
	CreatePermission(ctx context.Context, rp *model.RolePermission) (int64, error)
	// DeletePermission ...
	DeletePermission(ctx context.Context, id int64) error
	// ListPermission list role permissions according to the query.
	ListPermission(ctx context.Context, query *q.Query) ([]*model.RolePermission, error)
	// DeletePermissionByRole get permissions by role type and id
	DeletePermissionByRole(ctx context.Context, roleType string, roleID int64) error

	// CreateRbacPolicy ...
	CreateRbacPolicy(ctx context.Context, rp *model.RbacPolicy) (int64, error)
	// DeleteRbacPolicy ...
	DeleteRbacPolicy(ctx context.Context, id int64) error
	// ListRbacPolicy list RbacPolicy according to the query.
	ListRbacPolicy(ctx context.Context, query *q.Query) ([]*model.RbacPolicy, error)
	// GetPermissionsByRole ...
	GetPermissionsByRole(ctx context.Context, roleType string, roleID int64) ([]*model.RolePermissions, error)
}

// NewManager returns an instance of the default manager
func NewManager() Manager {
	return &manager{
		dao.New(),
	}
}

var _ Manager = &manager{}

type manager struct {
	dao dao.DAO
}

func (m *manager) CreatePermission(ctx context.Context, rp *model.RolePermission) (int64, error) {
	return m.dao.CreatePermission(ctx, rp)
}

func (m *manager) DeletePermission(ctx context.Context, id int64) error {
	return m.dao.DeletePermission(ctx, id)
}

func (m *manager) ListPermission(ctx context.Context, query *q.Query) ([]*model.RolePermission, error) {
	return m.dao.ListPermission(ctx, query)
}

func (m *manager) DeletePermissionByRole(ctx context.Context, roleType string, roleID int64) error {
	return m.dao.DeletePermissionByRole(ctx, roleType, roleID)
}

func (m *manager) CreateRbacPolicy(ctx context.Context, rp *model.RbacPolicy) (int64, error) {
	return m.dao.CreateRbacPolicy(ctx, rp)
}

func (m *manager) DeleteRbacPolicy(ctx context.Context, id int64) error {
	return m.dao.DeleteRbacPolicy(ctx, id)
}

func (m *manager) ListRbacPolicy(ctx context.Context, query *q.Query) ([]*model.RbacPolicy, error) {
	return m.dao.ListRbacPolicy(ctx, query)
}

func (m *manager) GetPermissionsByRole(ctx context.Context, roleType string, roleID int64) ([]*model.RolePermissions, error) {
	return m.dao.GetPermissionsByRole(ctx, roleType, roleID)
}
