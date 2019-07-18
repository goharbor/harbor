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
	"github.com/goharbor/harbor/src/pkg/retention/policy/rule"
)

const (
	// AlgorithmOR for OR algorithm
	AlgorithmOR = "or"
)

// Metadata of policy
type Metadata struct {
	// ID of the policy
	ID int64 `json:"id"`

	// Algorithm applied to the rules
	// "OR" / "AND"
	Algorithm string `json:"algorithm"`

	// Rule collection
	Rules []rule.Metadata `json:"rules"`

	// Trigger about how to launch the policy
	Trigger *Trigger `json:"trigger"`

	// Which scope the policy will be applied to
	Scope *Scope `json:"scope"`

	// The max number of rules in a policy
	Capacity int `json:"cap"`
}

// Trigger of the policy
type Trigger struct {
	// Const string to declare the trigger type
	// 'Schedule'
	Kind string `json:"kind"`

	// Settings for the specified trigger
	// '[cron]="* 22 11 * * *"' for the 'Schedule'
	Settings map[string]interface{} `json:"settings"`

	// References of the trigger
	// e.g: schedule job ID
	References map[string]interface{} `json:"references"`
}

// Scope definition
type Scope struct {
	// Scope level declaration
	// 'system', 'project' and 'repository'
	Level string `json:"level"`

	// The reference identity for the specified level
	// 0 for 'system', project ID for 'project' and repo ID for 'repository'
	Reference int64 `json:"ref"`
}

// LiteMeta contains partial metadata of policy
type LiteMeta struct {
	// Algorithm applied to the rules
	// "OR" / "AND"
	Algorithm string `json:"algorithm"`

	// Rule collection
	Rules []*rule.Metadata `json:"rules"`
}
