//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package purge

import (
	"encoding/json"
	"github.com/goharbor/harbor/src/lib/log"
)

// JobPolicy defines the purge job policy
type JobPolicy struct {
	Trigger           *Trigger               `json:"trigger"`
	DryRun            bool                   `json:"dryrun"`
	RetentionHour     int                    `json:"retention_hour"`
	IncludeOperations string                 `json:"include_operations"`
	ExtraAttrs        map[string]interface{} `json:"extra_attrs"`
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

// String convert map to json string
func String(extras map[string]interface{}) string {
	result, err := json.Marshal(extras)
	if err != nil {
		log.Errorf("failed to convert to json string, value %+v", extras)
	}
	return string(result)
}
