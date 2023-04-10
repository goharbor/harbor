// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	// ListPermissions list role permissions according to the query.
	ListPermissions(ctx context.Context, query *q.Query) ([]*model.RolePermission, error)
	// DeletePermissionsByRole get permissions by role type and id
	DeletePermissionsByRole(ctx context.Context, roleType string, roleID int64) error

	// CreateRbacPolicy ...
	CreateRbacPolicy(ctx context.Context, pp *model.PermissionPolicy) (int64, error)
	// DeleteRbacPolicy ...
	DeleteRbacPolicy(ctx context.Context, id int64) error
	// ListRbacPolicies list PermissionPolicy according to the query.
	ListRbacPolicies(ctx context.Context, query *q.Query) ([]*model.PermissionPolicy, error)
	// GetPermissionsByRole ...
	GetPermissionsByRole(ctx context.Context, roleType string, roleID int64) ([]*model.UniversalRolePermission, error)
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

func (m *manager) ListPermissions(ctx context.Context, query *q.Query) ([]*model.RolePermission, error) {
	return m.dao.ListPermissions(ctx, query)
}

func (m *manager) DeletePermissionsByRole(ctx context.Context, roleType string, roleID int64) error {
	return m.dao.DeletePermissionsByRole(ctx, roleType, roleID)
}

func (m *manager) CreateRbacPolicy(ctx context.Context, pp *model.PermissionPolicy) (int64, error) {
	return m.dao.CreateRbacPolicy(ctx, pp)
}

func (m *manager) DeleteRbacPolicy(ctx context.Context, id int64) error {
	return m.dao.DeleteRbacPolicy(ctx, id)
}

func (m *manager) ListRbacPolicies(ctx context.Context, query *q.Query) ([]*model.PermissionPolicy, error) {
	return m.dao.ListRbacPolicies(ctx, query)
}

func (m *manager) GetPermissionsByRole(ctx context.Context, roleType string, roleID int64) ([]*model.UniversalRolePermission, error) {
	return m.dao.GetPermissionsByRole(ctx, roleType, roleID)
}
