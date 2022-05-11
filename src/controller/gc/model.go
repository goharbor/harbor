package gc

import (
	"time"
)

// Policy ...
type Policy struct {
	Trigger        *Trigger               `json:"trigger"`
	DeleteUntagged bool                   `json:"deleteuntagged"`
	DryRun         bool                   `json:"dryrun"`
	ExtraAttrs     map[string]interface{} `json:"extra_attrs"`
}

// TriggerType represents the type of trigger.
type TriggerType string

// Trigger holds info for a trigger
type Trigger struct {
	Type     TriggerType      `json:"type"`
	Settings *TriggerSettings `json:"trigger_settings"`
}

// TriggerSettings is the setting about the trigger
type TriggerSettings struct {
	Cron string `json:"cron"`
}

// Execution model for gc
type Execution struct {
	ID            int64
	Status        string
	StatusMessage string
	Trigger       string
	ExtraAttrs    map[string]interface{}
	StartTime     time.Time
	UpdateTime    time.Time
}

// Task model for gc
type Task struct {
	ID             int64
	ExecutionID    int64
	Status         string
	StatusMessage  string
	RunCount       int32
	DeleteUntagged bool
	DryRun         bool
	JobID          string
	CreationTime   time.Time
	StartTime      time.Time
	UpdateTime     time.Time
	EndTime        time.Time
}
