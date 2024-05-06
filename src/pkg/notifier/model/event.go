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

import (
	"github.com/goharbor/harbor/src/controller/event/model"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
)

// HookEvent is hook related event data to publish
type HookEvent struct {
	ProjectID int64
	PolicyID  int64
	EventType string
	Target    *policy_model.EventTarget
	Payload   *Payload
}

// Payload of notification event
type Payload struct {
	Type      string     `json:"type"`
	OccurAt   int64      `json:"occur_at"`
	Operator  string     `json:"operator"`
	EventData *EventData `json:"event_data,omitempty"`
}

// EventData of notification event payload
type EventData struct {
	Resources   []*Resource        `json:"resources,omitempty"`
	Repository  *Repository        `json:"repository,omitempty"`
	Replication *model.Replication `json:"replication,omitempty"`
	Retention   *model.Retention   `json:"retention,omitempty"`
	Scan        *model.Scan        `json:"scan,omitempty"`
	Custom      map[string]string  `json:"custom_attributes,omitempty"`
}

// Resource describe infos of resource triggered notification
type Resource struct {
	Digest       string                 `json:"digest,omitempty"`
	Tag          string                 `json:"tag,omitempty"`
	ResourceURL  string                 `json:"resource_url,omitempty"`
	ScanOverview map[string]interface{} `json:"scan_overview,omitempty"`
	SBOMOverview map[string]interface{} `json:"sbom_overview,omitempty"`
}

// Repository info of notification event
type Repository struct {
	DateCreated  int64  `json:"date_created,omitempty"`
	Name         string `json:"name"`
	Namespace    string `json:"namespace"`
	RepoFullName string `json:"repo_full_name"`
	RepoType     string `json:"repo_type"`
}
