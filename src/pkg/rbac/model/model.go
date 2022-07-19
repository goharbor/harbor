package model

import (
	"time"

	"github.com/beego/beego/orm"
)

func init() {
	orm.RegisterModel(&RolePermission{})
	orm.RegisterModel(&PermissionPolicy{})
}

// RolePermission records the relations of role and permission
type RolePermission struct {
	ID                 int64     `orm:"pk;auto;column(id)"`
	RoleType           string    `orm:"column(role_type)"`
	RoleID             int64     `orm:"column(role_id)"`
	PermissionPolicyID int64     `orm:"column(permission_policy_id)"`
	CreationTime       time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName for role permission
func (rp *RolePermission) TableName() string {
	return "role_permission"
}

// PermissionPolicy records the policy of rbac
type PermissionPolicy struct {
	ID           int64     `orm:"pk;auto;column(id)"`
	Scope        string    `orm:"column(scope)"`
	Resource     string    `orm:"column(resource)"`
	Action       string    `orm:"column(action)"`
	Effect       string    `orm:"column(effect)"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName for permission policy
func (permissionPolicy *PermissionPolicy) TableName() string {
	return "permission_policy"
}

// UniversalRolePermission ...
type UniversalRolePermission struct {
	RoleType string `orm:"column(role_type)"`
	RoleID   int64  `orm:"column(role_id)"`
	Scope    string `orm:"column(scope)"`
	Resource string `orm:"column(resource)"`
	Action   string `orm:"column(action)"`
	Effect   string `orm:"column(effect)"`
}
