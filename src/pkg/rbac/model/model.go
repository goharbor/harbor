package model

import (
	"github.com/astaxie/beego/orm"
	"time"
)

func init() {
	orm.RegisterModel(&RolePermission{})
	orm.RegisterModel(&RbacPolicy{})
}

// RolePermission records the relations of role and permission
type RolePermission struct {
	ID           int64     `orm:"pk;auto;column(id)"`
	RoleType     string    `orm:"column(role_type)"`
	RoleID       int64     `orm:"column(role_id)"`
	RBACPolicyID int64     `orm:"column(rbac_policy_id)"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName for role permission
func (rp *RolePermission) TableName() string {
	return "role_permission"
}

// RbacPolicy records the policy of rbac
type RbacPolicy struct {
	ID           int64     `orm:"pk;auto;column(id)"`
	Scope        string    `orm:"column(scope)"`
	Resource     string    `orm:"column(resource)"`
	Action       string    `orm:"column(action)"`
	Effect       string    `orm:"column(effect)"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
}

// TableName for rbac policy
func (rbacPolicy *RbacPolicy) TableName() string {
	return "rbac_policy"
}

// RolePermissions ...
type RolePermissions struct {
	RoleType string `orm:"column(role_type)"`
	RoleID   int64  `orm:"column(role_id)"`
	Scope    string `orm:"column(scope)"`
	Resource string `orm:"column(resource)"`
	Action   string `orm:"column(action)"`
	Effect   string `orm:"column(effect)"`
}
