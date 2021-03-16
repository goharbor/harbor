package dao

import (
	"context"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/rbac/model"
	"time"
)

// DAO is the data access object interface for rbac policy
type DAO interface {
	// CreatePermission ...
	CreatePermission(ctx context.Context, rp *model.RolePermission) (int64, error)
	// DeletePermission ...
	DeletePermission(ctx context.Context, id int64) error
	// ListPermissions ...
	ListPermissions(ctx context.Context, query *q.Query) ([]*model.RolePermission, error)
	// DeletePermissionsByRole ...
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

// New returns an instance of the default DAO
func New() DAO {
	return &dao{}
}

type dao struct{}

func (d *dao) CreatePermission(ctx context.Context, rp *model.RolePermission) (id int64, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	rp.CreationTime = time.Now()
	return ormer.InsertOrUpdate(rp, "role_type, role_id, permission_policy_id")
}

func (d *dao) DeletePermission(ctx context.Context, id int64) (err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.RolePermission{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("role permission %d not found", id)
	}
	return nil
}

func (d *dao) ListPermissions(ctx context.Context, query *q.Query) ([]*model.RolePermission, error) {
	rps := []*model.RolePermission{}
	qs, err := orm.QuerySetter(ctx, &model.RolePermission{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&rps); err != nil {
		return nil, err
	}
	return rps, nil
}

func (d *dao) DeletePermissionsByRole(ctx context.Context, roleType string, roleID int64) error {
	qs, err := orm.QuerySetter(ctx, &model.RolePermission{}, &q.Query{
		Keywords: map[string]interface{}{
			"role_type": roleType,
			"role_id":   roleID,
		},
	})
	if err != nil {
		return err
	}
	n, err := qs.Delete()
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("role permission %s:%d not found", roleType, roleID)
	}
	return err
}

func (d *dao) CreateRbacPolicy(ctx context.Context, pp *model.PermissionPolicy) (id int64, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	pp.CreationTime = time.Now()
	return ormer.InsertOrUpdate(pp, "scope, resource, action, effect")
}

func (d *dao) DeleteRbacPolicy(ctx context.Context, id int64) (err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.PermissionPolicy{
		ID: id,
	})
	if err != nil {
		return err
	}
	if n == 0 {
		return errors.NotFoundError(nil).WithMessage("rbac policy %d not found", id)
	}
	return nil
}

func (d *dao) ListRbacPolicies(ctx context.Context, query *q.Query) ([]*model.PermissionPolicy, error) {
	pps := []*model.PermissionPolicy{}
	qs, err := orm.QuerySetter(ctx, &model.PermissionPolicy{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&pps); err != nil {
		return nil, err
	}
	return pps, nil
}

func (d *dao) GetPermissionsByRole(ctx context.Context, roleType string, roleID int64) ([]*model.UniversalRolePermission, error) {
	var rps []*model.UniversalRolePermission
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return rps, err
	}
	_, err = ormer.Raw("SELECT rper.role_type, rper.role_id, ppo.scope, ppo.resource, ppo.action, ppo.effect FROM role_permission AS rper LEFT JOIN permission_policy ppo ON (rper.permission_policy_id=ppo.id) where rper.role_type=? and rper.role_id=?", roleType, roleID).QueryRows(&rps)
	if err != nil {
		return rps, err
	}

	return rps, nil
}
