package models

import (
	"encoding/json"
	"time"
)

const (
	// NotificationPolicyTable is table name for notification policies
	NotificationPolicyTable = "notification_policy"
	// NotificationJobTable is table name for notification job
	NotificationJobTable = "notification_job"
)

// NotificationPolicy is the model for a notification policy.
type NotificationPolicy struct {
	ID           int64         `orm:"pk;auto;column(id)" json:"id"`
	Name         string        `orm:"column(name)" json:"name"`
	Description  string        `orm:"column(description)" json:"description"`
	ProjectID    int64         `orm:"column(project_id)" json:"project_id"`
	TargetsDB    string        `orm:"column(targets)" json:"-"`
	Targets      []EventTarget `orm:"-" json:"targets"`
	EventTypesDB string        `orm:"column(event_types)" json:"-"`
	EventTypes   []string      `orm:"-" json:"event_types"`
	Creator      string        `orm:"column(creator)" json:"creator"`
	CreationTime time.Time     `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time     `orm:"column(update_time);auto_now_add" json:"update_time"`
	Enabled      bool          `orm:"column(enabled)" json:"enabled"`
}

// TableName set table name for ORM.
func (w *NotificationPolicy) TableName() string {
	return NotificationPolicyTable
}

// ConvertToDBModel convert struct data in notification policy to DB model data
func (w *NotificationPolicy) ConvertToDBModel() error {
	if len(w.Targets) != 0 {
		targets, err := json.Marshal(w.Targets)
		if err != nil {
			return err
		}
		w.TargetsDB = string(targets)
	}
	if len(w.EventTypes) != 0 {
		eventTypes, err := json.Marshal(w.EventTypes)
		if err != nil {
			return err
		}
		w.EventTypesDB = string(eventTypes)
	}

	return nil
}

// ConvertFromDBModel convert from DB model data to struct data
func (w *NotificationPolicy) ConvertFromDBModel() error {
	targets := []EventTarget{}
	if len(w.TargetsDB) != 0 {
		err := json.Unmarshal([]byte(w.TargetsDB), &targets)
		if err != nil {
			return err
		}
	}
	w.Targets = targets

	types := []string{}
	if len(w.EventTypesDB) != 0 {
		err := json.Unmarshal([]byte(w.EventTypesDB), &types)
		if err != nil {
			return err
		}
	}
	w.EventTypes = types

	return nil
}

// NotificationJob is the model for a notification job
type NotificationJob struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	PolicyID     int64     `orm:"column(policy_id)" json:"policy_id"`
	EventType    string    `orm:"column(event_type)" json:"event_type"`
	NotifyType   string    `orm:"column(notify_type)" json:"notify_type"`
	Status       string    `orm:"column(status)" json:"status"`
	JobDetail    string    `orm:"column(job_detail)" json:"job_detail"`
	UUID         string    `orm:"column(job_uuid)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName set table name for ORM.
func (w *NotificationJob) TableName() string {
	return NotificationJobTable
}

// NotificationJobQuery holds query conditions for notification job
type NotificationJobQuery struct {
	PolicyID   int64
	Statuses   []string
	EventTypes []string
	Pagination
}

// EventTarget defines the structure of target a notification send to
type EventTarget struct {
	Type           string `json:"type"`
	Address        string `json:"address"`
	AuthHeader     string `json:"auth_header,omitempty"`
	SkipCertVerify bool   `json:"skip_cert_verify"`
}
