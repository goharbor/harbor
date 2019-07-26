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

package retention

import "time"

// const definitions
const (
	ExecutionStatusInProgress string = "InProgress"
	ExecutionStatusSucceed    string = "Succeed"
	ExecutionStatusFailed     string = "Failed"
	ExecutionStatusStopped    string = "Stopped"

	TaskStatusPending    string = "Pending"
	TaskStatusInProgress string = "InProgress"
	TaskStatusSucceed    string = "Succeed"
	TaskStatusFailed     string = "Failed"
	TaskStatusStopped    string = "Stopped"

	CandidateKindImage string = "image"
	CandidateKindChart string = "chart"

	ExecutionTriggerManual   string = "Manual"
	ExecutionTriggerSchedule string = "Schedule"
)

// Execution of retention
type Execution struct {
	ID        int64     `json:"id"`
	PolicyID  int64     `json:"policy_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Status    string    `json:"status"`
	Trigger   string    `json:"Trigger"`
	DryRun    bool      `json:"dry_run"`
}

// Task of retention
type Task struct {
	ID          int64     `json:"id"`
	ExecutionID int64     `json:"execution_id"`
	JobID       string    `json:"job_id"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

// History of retention
type History struct {
	ID          int64 `json:"id,omitempty"`
	ExecutionID int64 `json:"execution_id"`
	Rule        struct {
		ID          int    `json:"id"`
		DisplayText string `json:"display_text"`
	} `json:"rule_id"`
	// full path: :ns/:repo:tag
	Artifact  string    `json:"tag"`
	Timestamp time.Time `json:"timestamp"`
}
