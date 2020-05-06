package model

import (
	beego_orm "github.com/astaxie/beego/orm"
	"time"
)

func init() {
	beego_orm.RegisterModel(&AuditLog{})
}

// AuditLog ...
type AuditLog struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	Operation    string    `orm:"column(operation)" json:"operation"`
	ResourceType string    `orm:"column(resource_type)"  json:"resource_type"`
	Resource     string    `orm:"column(resource)" json:"resource"`
	Username     string    `orm:"column(username)"  json:"username"`
	OpTime       time.Time `orm:"column(op_time)" json:"op_time"`
}

// TableName for audit log
func (a *AuditLog) TableName() string {
	return "audit_log"
}
