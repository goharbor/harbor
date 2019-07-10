package models

import "time"

// WebhookPolicy defines the structure of a webhook policy for API
type WebhookPolicy struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	ProjectID    int64         `json:"project_id"`
	Targets      []*HookTarget `json:"targets"`
	HookTypes    []string      `json:"hook_types"`
	Creator      string        `json:"creator"`
	CreationTime time.Time     `json:"creation_time"`
	UpdateTime   time.Time     `json:"update_time"`
	Enabled      bool          `json:"enabled"`
}

// HookTarget defines the structure of target a webhook send to for API
type HookTarget struct {
	Type       string `json:"type"`
	Address    string `json:"address"`
	Attachment string `json:"attachment"`
	Secret     string `json:"secret,omitempty"`
}
