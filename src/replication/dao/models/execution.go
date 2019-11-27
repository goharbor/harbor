package models

import (
	"time"

	"github.com/goharbor/harbor/src/replication/model"
)

const (
	// ExecutionTable is the table name for replication executions
	ExecutionTable = "replication_execution"
	// TaskTable is table name for replication tasks
	TaskTable = "replication_task"
)

// execution/task status/trigger const
const (
	ExecutionStatusFailed     string = "Failed"
	ExecutionStatusSucceed    string = "Succeed"
	ExecutionStatusStopped    string = "Stopped"
	ExecutionStatusInProgress string = "InProgress"

	ExecutionTriggerManual   string = "Manual"
	ExecutionTriggerEvent    string = "Event"
	ExecutionTriggerSchedule string = "Schedule"

	// The task has been persisted in db but not submitted to Jobservice
	TaskStatusInitialized string = "Initialized"
	TaskStatusPending     string = "Pending"
	TaskStatusInProgress  string = "InProgress"
	TaskStatusSucceed     string = "Succeed"
	TaskStatusFailed      string = "Failed"
	TaskStatusStopped     string = "Stopped"
)

// ExecutionPropsName defines the names of fields of Execution
var ExecutionPropsName = ExecutionFieldsName{
	ID:         "ID",
	PolicyID:   "PolicyID",
	Status:     "Status",
	StatusText: "StatusText",
	Total:      "Total",
	Failed:     "Failed",
	Succeed:    "Succeed",
	InProgress: "InProgress",
	Stopped:    "Stopped",
	Trigger:    "Trigger",
	StartTime:  "StartTime",
	EndTime:    "EndTime",
}

// ExecutionFieldsName defines the props of Execution
type ExecutionFieldsName struct {
	ID         string
	PolicyID   string
	Status     string
	StatusText string
	Total      string
	Failed     string
	Succeed    string
	InProgress string
	Stopped    string
	Trigger    string
	StartTime  string
	EndTime    string
}

// Execution holds information about once replication execution.
type Execution struct {
	ID         int64             `orm:"pk;auto;column(id)" json:"id"`
	PolicyID   int64             `orm:"column(policy_id)" json:"policy_id"`
	Status     string            `orm:"column(status)" json:"status"`
	StatusText string            `orm:"column(status_text)" json:"status_text"`
	Total      int               `orm:"column(total)" json:"total"`
	Failed     int               `orm:"column(failed)" json:"failed"`
	Succeed    int               `orm:"column(succeed)" json:"succeed"`
	InProgress int               `orm:"column(in_progress)" json:"in_progress"`
	Stopped    int               `orm:"column(stopped)" json:"stopped"`
	Trigger    model.TriggerType `orm:"column(trigger)" json:"trigger"`
	StartTime  time.Time         `orm:"column(start_time)" json:"start_time"`
	EndTime    time.Time         `orm:"column(end_time)" json:"end_time"`
}

// TaskPropsName defines the names of fields of Task
var TaskPropsName = TaskFieldsName{
	ID:           "ID",
	ExecutionID:  "ExecutionID",
	ResourceType: "ResourceType",
	SrcResource:  "SrcResource",
	DstResource:  "DstResource",
	JobID:        "JobID",
	Status:       "Status",
	StartTime:    "StartTime",
	EndTime:      "EndTime",
}

// TaskFieldsName defines the props of Task
type TaskFieldsName struct {
	ID           string
	ExecutionID  string
	ResourceType string
	SrcResource  string
	DstResource  string
	JobID        string
	Status       string
	StartTime    string
	EndTime      string
}

// Task represent the tasks in one execution.
type Task struct {
	ID             int64     `orm:"pk;auto;column(id)" json:"id"`
	ExecutionID    int64     `orm:"column(execution_id)" json:"execution_id"`
	ResourceType   string    `orm:"column(resource_type)" json:"resource_type"`
	SrcResource    string    `orm:"column(src_resource)" json:"src_resource"`
	DstResource    string    `orm:"column(dst_resource)" json:"dst_resource"`
	Operation      string    `orm:"column(operation)" json:"operation"`
	JobID          string    `orm:"column(job_id)" json:"job_id"`
	Status         string    `orm:"column(status)" json:"status"`
	StatusRevision int64     `orm:"column(status_revision)"`
	StartTime      time.Time `orm:"column(start_time)" json:"start_time"`
	EndTime        time.Time `orm:"column(end_time)" json:"end_time,omitempty"`
}

// TableName is required by by beego orm to map Execution to table replication_execution
func (r *Execution) TableName() string {
	return ExecutionTable
}

// TableName is required by by beego orm to map Task to table replication_task
func (r *Task) TableName() string {
	return TaskTable
}

// ExecutionQuery holds the query conditions for replication executions
type ExecutionQuery struct {
	PolicyID int64
	Statuses []string
	Trigger  string
	Pagination
}

// TaskQuery holds the query conditions for replication task
type TaskQuery struct {
	ExecutionID  int64
	JobID        string
	Statuses     []string
	ResourceType string
	Pagination
}

// TaskStat holds statistics of task by status
type TaskStat struct {
	Status string `orm:"column(status)"`
	C      int    `orm:"column(c)"`
}
