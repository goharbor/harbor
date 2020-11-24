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

package replication

import (
	"time"

	"github.com/goharbor/harbor/src/pkg/task/dao"
)

// Execution model for replication
type Execution struct {
	ID            int64
	PolicyID      int64
	Status        string
	StatusMessage string
	Metrics       *dao.Metrics
	Trigger       string
	StartTime     time.Time
	EndTime       time.Time
}

// Task model for replication
type Task struct {
	ID                  int64
	ExecutionID         int64
	Status              string
	StatusMessage       string
	RunCount            int32
	ResourceType        string
	SourceResource      string
	DestinationResource string
	Operation           string
	JobID               string
	CreationTime        time.Time
	StartTime           time.Time
	UpdateTime          time.Time
	EndTime             time.Time
}
