package models

import (
	"github.com/astaxie/beego/orm"
	"time"
)

func init() {
	orm.RegisterModel(
		new (RetentionPolicy),
		new (RetentionExecution),
		new (RetentionTask),
		new (RetentionScheduleJob),
		)
}

type RetentionPolicy struct {
	ID int64 `orm:"pk;auto;column(id)" json:"id"`
	// 'system', 'project' and 'repository'
	ScopeLevel      string
	ScopeReference  int64
	TriggerKind     string
	// json format, include algorithm, rules, exclusions
	Data       string
	CreateTime time.Time
	UpdateTime time.Time
}

type RetentionExecution struct {
	ID         int64 `orm:"pk;auto;column(id)" json:"id"`
	PolicyID   int64
	Status     string
	StatusText string
	Dry        bool
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

type RetentionTask struct {
	ID          int64
	ExecutionID int64
	// image, chart
	ResourceType string
	Resource     string
	Status       string
	StartTime    time.Time
	EndTime      time.Time
}

type RetentionScheduleJob struct {
	ID         int64
	Status     string
	PolicyID   int64
	JobID      int64
	CreateTime time.Time
	UpdateTime time.Time
}
