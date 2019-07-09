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

// Execution of retention
type Execution struct {
	ID        string    `json:"id"`
	PolicyID  string    `json:"policy_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

// TaskSubmitResult is the result of task submitting
// If the task is submitted successfully, JobID will be set
// and the Error is nil
type TaskSubmitResult struct {
	JobID string
	Error error
}

// History of retention
type History struct {
	ExecutionID string `json:"execution_id"`
	Rule        struct {
		ID          string `json:"id"`
		DisplayText string `json:"display_text"`
	} `json:"rule_id"`
	// full path: :ns/:repo:tag
	Artifact  string    `json:"tag"`
	Timestamp time.Time `json:"timestamp"`
}
