// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package gc

import (
	"time"
)

// Policy ...
type Policy struct {
	Trigger        *Trigger               `json:"trigger"`
	DeleteUntagged bool                   `json:"deleteuntagged"`
	DryRun         bool                   `json:"dryrun"`
	Workers        int                    `json:"workers"`
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
	Workers        int
	JobID          string
	CreationTime   time.Time
	StartTime      time.Time
	UpdateTime     time.Time
	EndTime        time.Time
}
