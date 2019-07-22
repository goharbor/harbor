package models

import "time"

// WebhookPolicy defines the structure of a webhook policy for API
type WebhookPolicy struct {
	ID           int64         `json:"id"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Targets      []*HookTarget `json:"targets"`
	HookTypes    []string      `json:"hook_types"`
	Creator      string        `json:"creator"`
	CreationTime time.Time     `json:"creation_time"`
	UpdateTime   time.Time     `json:"update_time"`
	Enabled      bool          `json:"enabled"`
}

// HookTarget defines the structure of target a webhook send to for API
type HookTarget struct {
	Type           string `json:"type"`
	Address        string `json:"address"`
	Token          string `json:"token,omitempty"`
	SkipCertVerify bool   `json:"skip_cert_verify"`
}

// WebhookPolicyForUI defines the structure of webhook policy info display in UI
type WebhookPolicyForUI struct {
	HookType        string     `json:"hook_type"`
	Enabled         bool       `json:"enabled"`
	CreationTime    *time.Time `json:"creation_time"`
	LastTriggerTime *time.Time `json:"last_trigger_time,omitempty"`
}
