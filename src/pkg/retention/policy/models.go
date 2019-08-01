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

package policy

import (
	"github.com/astaxie/beego/validation"
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

const (
	// AlgorithmOR for OR algorithm
	AlgorithmOR = "or"

	// TriggerKindSchedule Schedule
	TriggerKindSchedule = "Schedule"

	// TriggerReferencesJobid job_id
	TriggerReferencesJobid = "job_id"
	// TriggerSettingsCron cron
	TriggerSettingsCron = "cron"

	// ScopeLevelProject project
	ScopeLevelProject = "project"
)

// Metadata of policy
type Metadata struct {
	// ID of the policy
	ID int64 `json:"id"`

	// Algorithm applied to the rules
	// "OR" / "AND"
	Algorithm string `json:"algorithm" valid:"Required;Match(or)"`

	// Rule collection
	Rules []rule.Metadata `json:"rules"`

	// Trigger about how to launch the policy
	Trigger *Trigger `json:"trigger" valid:"Required"`

	// Which scope the policy will be applied to
	Scope *Scope `json:"scope" valid:"Required"`

	// The max number of rules in a policy
	Capacity int `json:"cap"`
}

// Valid Valid
func (m *Metadata) Valid(v *validation.Validation) {
	if m.Trigger != nil && m.Trigger.Kind == TriggerKindSchedule {
		if m.Trigger.Settings == nil {
			_ = v.SetError("Trigger.Settings", "Trigger.Settings is required")
		} else {
			if _, ok := m.Trigger.Settings[TriggerSettingsCron]; !ok {
				_ = v.SetError("Trigger.Settings", "cron in Trigger.Settings is required")
			}
		}

	}
}

// Trigger of the policy
type Trigger struct {
	// Const string to declare the trigger type
	// 'Schedule'
	Kind string `json:"kind" valid:"Required"`

	// Settings for the specified trigger
	// '[cron]="* 22 11 * * *"' for the 'Schedule'
	Settings map[string]interface{} `json:"settings" valid:"Required"`

	// References of the trigger
	// e.g: schedule job ID
	References map[string]interface{} `json:"references"`
}

// Scope definition
type Scope struct {
	// Scope level declaration
	// 'system', 'project' and 'repository'
	Level string `json:"level" valid:"Required;Match(/^(project)$/)"`

	// The reference identity for the specified level
	// 0 for 'system', project ID for 'project' and repo ID for 'repository'
	Reference int64 `json:"ref" valid:"Required"`
}
