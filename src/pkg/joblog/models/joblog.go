package models

import (
	"time"

	"github.com/beego/beego/orm"
)

func init() {
	orm.RegisterModel(&JobLog{})
}

// JobLogTable is the name of the table that record the job execution result.
const JobLogTable = "job_log"

// JobLog holds information about logs which are used to record the result of execution of a job.
type JobLog struct {
	LogID        int       `orm:"pk;auto;column(log_id)" json:"log_id"`
	UUID         string    `orm:"column(job_uuid)" json:"uuid"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	Content      string    `orm:"column(content)" json:"content"`
}

// TableName is required by by beego orm to map JobLog to table job_log
func (a *JobLog) TableName() string {
	return JobLogTable
}
