package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

func init() {
	orm.RegisterModel(
		new(RetentionPolicy),
		new(RetentionExecution),
		new(RetentionTask),
		new(RetentionScheduleJob),
	)
}

// RetentionPolicy Retention Policy
type RetentionPolicy struct {
	ID int64 `orm:"pk;auto;column(id)" json:"id"`
	// 'system', 'project' and 'repository'
	ScopeLevel     string
	ScopeReference int64
	TriggerKind    string
	// json format, include algorithm, rules, exclusions
	Data       string
	CreateTime time.Time
	UpdateTime time.Time
}

// RetentionExecution Retention Execution
type RetentionExecution struct {
	ID       int64 `orm:"pk;auto;column(id)" json:"id"`
	PolicyID int64 `orm:"column(policy_id)"`
	Status   string
	DryRun   bool
	// manual, scheduled
	Trigger    string
	Total      int
	Succeed    int
	Failed     int
	InProgress int
	Stopped    int
	StartTime  time.Time
	EndTime    time.Time
}

// RetentionTask ...
type RetentionTask struct {
	ID          int64     `orm:"pk;auto;column(id)"`
	ExecutionID int64     `orm:"column(execution_id)"`
	Status      string    `orm:"column(status)"`
	StartTime   time.Time `orm:"column(start_time)"`
	EndTime     time.Time `orm:"column(end_time)"`
}

// RetentionScheduleJob Retention Schedule Job
type RetentionScheduleJob struct {
	ID         int64 `orm:"pk;auto;column(id)" json:"id"`
	Status     string
	PolicyID   int64 `orm:"column(policy_id)"`
	JobID      int64 `orm:"column(job_id)"`
	CreateTime time.Time
	UpdateTime time.Time
}
