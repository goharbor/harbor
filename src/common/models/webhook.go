package models

import (
	"encoding/json"
	"time"
)

const (
	// WebhookPolicyTable is table name for webhook policies
	WebhookPolicyTable = "webhook_policy"
	// WebhookJobTable is table name for webhook job
	WebhookJobTable = "webhook_job"
)

// WebhookPolicy is the model for a webhook policy.
type WebhookPolicy struct {
	ID           int64        `orm:"pk;auto;column(id)" json:"id"`
	Name         string       `orm:"column(name)" json:"name"`
	Description  string       `orm:"column(description)" json:"description"`
	ProjectID    int64        `orm:"column(project_id)" json:"project_id"`
	TargetsDB    string       `orm:"column(targets)"`
	Targets      []HookTarget `orm:"-" json:"targets"`
	HookTypesDB  string       `orm:"column(hook_types)"`
	HookTypes    []string     `orm:"-" json:"hook_types"`
	Creator      string       `orm:"column(creator)" json:"creator"`
	CreationTime time.Time    `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time    `orm:"column(update_time);auto_now_add" json:"update_time"`
	Enabled      bool         `orm:"column(enabled)" json:"enabled"`
}

// TableName set table name for ORM.
func (w *WebhookPolicy) TableName() string {
	return WebhookPolicyTable
}

// ConvertToDBModel convert struct data in webhook policy to DB model data
func (w *WebhookPolicy) ConvertToDBModel() error {
	if len(w.Targets) != 0 {
		targets, err := json.Marshal(w.Targets)
		if err != nil {
			return err
		}
		w.TargetsDB = string(targets)
	}
	if len(w.HookTypes) != 0 {
		hookTypes, err := json.Marshal(w.HookTypes)
		if err != nil {
			return err
		}
		w.HookTypesDB = string(hookTypes)
	}

	return nil
}

// ConvertFromDBModel convert from DB model data to struct data
func (w *WebhookPolicy) ConvertFromDBModel() error {
	if w == nil {
		return nil
	}

	targets := []HookTarget{}
	if len(w.TargetsDB) != 0 {
		err := json.Unmarshal([]byte(w.TargetsDB), &targets)
		if err != nil {
			return err
		}
	}
	w.Targets = targets

	types := []string{}
	if len(w.HookTypesDB) != 0 {
		err := json.Unmarshal([]byte(w.HookTypesDB), &types)
		if err != nil {
			return err
		}
	}
	w.HookTypes = types

	return nil
}

// WebhookJob is the model for a webhook job
type WebhookJob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	HookType     string    `orm:"column(hook_type)" json:"hook_type"`
	NotifyType   string    `orm:"column(notify_type)" json:"notify_type"`
	Status       string    `orm:"column(status)" json:"status"`
	JobDetail    string    `orm:"column(job_detail)" json:"job_detail"`
	UUID         string    `orm:"column(job_uuid)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName set table name for ORM.
func (w *WebhookJob) TableName() string {
	return WebhookJobTable
}

// WebhookJobQuery holds query conditions for webhook job
type WebhookJobQuery struct {
	PolicyID  int64
	Statuses  []string
	HookTypes []string
	Pagination
}

// HookTarget defines the structure of target a webhook send to
type HookTarget struct {
	Type           string `json:"type"`
	Address        string `json:"address"`
	Token          string `json:"token,omitempty"`
	SkipCertVerify bool   `json:"skip_cert_verify"`
}

// LastTriggerInfo records last trigger time of hook type
type LastTriggerInfo struct {
	HookType     string    `orm:"column(hook_type)"`
	CreationTime time.Time `orm:"column(ct)"`
}
