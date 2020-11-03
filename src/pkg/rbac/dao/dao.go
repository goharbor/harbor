package dao

import (
	"context"
	"fmt"
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
	// ListPermission ...
	ListPermission(ctx context.Context, query *q.Query) ([]*model.RolePermission, error)
	// DeletePermissionByRole ...
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
	return ormer.InsertOrUpdate(rp, "role_type, role_id, rbac_policy_id")
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

func (d *dao) ListPermission(ctx context.Context, query *q.Query) ([]*model.RolePermission, error) {
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

func (d *dao) DeletePermissionByRole(ctx context.Context, roleType string, roleID int64) error {
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

func (d *dao) CreateRbacPolicy(ctx context.Context, rp *model.RbacPolicy) (id int64, err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}
	rp.CreationTime = time.Now()
	return ormer.InsertOrUpdate(rp, "scope, resource, action, effect")
}

func (d *dao) DeleteRbacPolicy(ctx context.Context, id int64) (err error) {
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}
	n, err := ormer.Delete(&model.RbacPolicy{
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

func (d *dao) ListRbacPolicy(ctx context.Context, query *q.Query) ([]*model.RbacPolicy, error) {
	rps := []*model.RbacPolicy{}
	qs, err := orm.QuerySetter(ctx, &model.RbacPolicy{}, query)
	if err != nil {
		return nil, err
	}
	if _, err = qs.All(&rps); err != nil {
		return nil, err
	}
	return rps, nil
}

func (d *dao) GetPermissionsByRole(ctx context.Context, roleType string, roleID int64) ([]*model.RolePermissions, error) {
	var rps []*model.RolePermissions
	ormer, err := orm.FromContext(ctx)
	if err != nil {
		return rps, err
	}
	sql := fmt.Sprintf(`SELECT rper.role_type, rper.role_id, rpo.scope, rpo.resource, rpo.action, rpo.effect FROM role_permission AS rper LEFT JOIN rbac_policy rpo ON (rper.rbac_policy_id=rpo.id) where rper.role_type='%s' and rper.role_id=%d`, roleType, roleID)

	_, err = ormer.Raw(sql).QueryRows(&rps)
	if err != nil {
		return rps, err
	}

	return rps, nil
}
