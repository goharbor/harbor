package model

import (
	"time"

	"github.com/beego/beego/orm"
)

func init() {
	orm.RegisterModel(&Job{})
}

// Job is the model for a notification job
type Job struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	EventType    string    `orm:"column(event_type)" json:"event_type"`
	NotifyType   string    `orm:"column(notify_type)" json:"notify_type"`
	Status       string    `orm:"column(status)" json:"status"`
	JobDetail    string    `orm:"column(job_detail)" json:"job_detail"`
	UUID         string    `orm:"column(job_uuid)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time" sort:"default:desc"`
}

// TableName set table name for ORM.
func (j *Job) TableName() string {
	return "notification_job"
}
