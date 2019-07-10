package model

import "time"

// WebhookPolicy defines the structure of a webhook policy
type WebhookPolicy struct {
	ID           int64
	Name         string
	Description  string
	ProjectID    int64
	Targets      []HookTarget
	HookTypes    []string
	Creator      string
	CreationTime time.Time
	UpdateTime   time.Time
	Enabled      bool
}

// HookTarget defines the structure of target a webhook send to
type HookTarget struct {
	Type       string
	Address    string
	Attachment string
	Secret     string
}
