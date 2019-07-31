package models

import "time"

// NotificationPolicyForUI defines the structure of notification policy info display in UI
type NotificationPolicyForUI struct {
	EventType       string     `json:"event_type"`
	Enabled         bool       `json:"enabled"`
	CreationTime    *time.Time `json:"creation_time"`
	LastTriggerTime *time.Time `json:"last_trigger_time,omitempty"`
}
