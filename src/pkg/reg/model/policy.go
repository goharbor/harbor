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

package model

// const definition
const (
	FilterTypeResource = "resource"
	FilterTypeName     = "name"
	FilterTypeTag      = "tag"
	FilterTypeLabel    = "label"

	TriggerTypeManual     = "manual"
	TriggerTypeScheduled  = "scheduled"
	TriggerTypeEventBased = "event_based"
)

// Filter holds the info of the filter
type Filter struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

// Trigger holds info for a trigger
type Trigger struct {
	Type     string           `json:"type"`
	Settings *TriggerSettings `json:"trigger_settings"`
}

// TriggerSettings is the setting about the trigger
type TriggerSettings struct {
	Cron string `json:"cron"`
}
