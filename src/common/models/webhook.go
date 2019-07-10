package models

import "time"

const (
	// WebhookPolicyTable is table name for webhook policies
	WebhookPolicyTable = "webhook_policy"
	// WebhookExecutionTable is table name for webhook execution
	WebhookExecutionTable = "webhook_execution"
)

// WebhookPolicy is the model for a webhook policy.
type WebhookPolicy struct {
	ID           int64     `orm:"pk;auto;column(id)"`
	Name         string    `orm:"column(name)"`
	Description  string    `orm:"column(description)"`
	ProjectID    int64     `orm:"column(project_id)"`
	Targets      string    `orm:"column(targets)"`
	HookTypes    string    `orm:"column(hook_types)"`
	Creator      string    `orm:"column(creator)"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now_add"`
	Enabled      bool      `orm:"column(enabled)"`
}

// TableName set table name for ORM.
func (w *WebhookPolicy) TableName() string {
	return WebhookPolicyTable
}

// WebhookExecution is the model for a webhook execution
type WebhookExecution struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	HookType     string    `orm:"column(hook_type)" json:"hook_type"`
	Status       string    `orm:"column(status)" json:"status"`
	JobDetail    string    `orm:"column(job_detail)" json:"job_detail"`
	UUID         string    `orm:"column(job_uuid)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName set table name for ORM.
func (w *WebhookExecution) TableName() string {
	return WebhookExecutionTable
}

// WebhookExecutionQuery holds query conditions for webhook execution
type WebhookExecutionQuery struct {
	PolicyID  int64
	Statuses  []string
	HookTypes []string
	StartTime *time.Time
	EndTime   *time.Time
	Pagination
}
