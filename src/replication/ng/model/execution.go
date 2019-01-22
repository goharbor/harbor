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

import "time"

// Execution defines an execution of the replication
type Execution struct {
	ID         int64     `json:"id"`
	PolicyID   int64     `json:"policy_id"`
	Total      int       `json:"total"`
	Failed     int       `json:"failed"`
	Succeed    int       `json:"succeed"`
	Pending    int       `json:"pending"`
	InProgress int       `json:"in_progress"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
}

// Task holds the information of one replication task
type Task struct {
	ID           int64        `json:"id"`
	ExecutionID  int64        `json:"execution_id"`
	ResourceType ResourceType `json:"resource_type"`
	SrcResource  string       `json:"src_resource"`
	DstResource  string       `json:"dst_resource"`
	JobID        string       `json:"job_id"`
	Status       string       `json:"status"`
	StartTime    time.Time    `json:"start_time"`
	EndTime      time.Time    `json:"end_time"`
}
