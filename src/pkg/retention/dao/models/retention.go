package models

import (
	"time"

	"github.com/astaxie/beego/orm"
)

// const definitions
const (
	ExecutionStatusInProgress string = "InProgress"
	ExecutionStatusSucceed    string = "Succeed"
	ExecutionStatusFailed     string = "Failed"
	ExecutionStatusStopped    string = "Stopped"
)

func init() {
	orm.RegisterModel(
		new(RetentionPolicy),
		new(RetentionExecution),
		new(RetentionTask),
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
	DryRun   bool
	// manual, scheduled
	Trigger   string
	StartTime time.Time
	EndTime   time.Time `orm:"-"`
	Status    string    `orm:"-"`
}

// RetentionTask ...
type RetentionTask struct {
	ID          int64     `orm:"pk;auto;column(id)"`
	ExecutionID int64     `orm:"column(execution_id)"`
	Repository  string    `orm:"column(repository)"`
	JobID       string    `orm:"column(job_id)"`
	Status      string    `orm:"column(status)"`
	StatusCode  int       `orm:"column(status_code)"`
	StartTime   time.Time `orm:"column(start_time)"`
	EndTime     time.Time `orm:"column(end_time)"`
	Total       int       `orm:"column(total)"`
	Retained    int       `orm:"column(retained)"`
}
