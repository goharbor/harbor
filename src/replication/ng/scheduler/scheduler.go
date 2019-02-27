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

package scheduler

import (
	"github.com/goharbor/harbor/src/replication/ng/model"
)

// ScheduleItem is an item that can be scheduled
type ScheduleItem struct {
	TaskID      int64 // used as the param in the hook
	SrcResource *model.Resource
	DstResource *model.Resource
}

// ScheduleResult is the result of the schedule for one item
type ScheduleResult struct {
	TaskID int64
	Error  error
}

// Scheduler schedules
type Scheduler interface {
	// Preprocess the resources and returns the item list that can be scheduled
	Preprocess([]*model.Resource, []*model.Resource) ([]*ScheduleItem, error)
	// Schedule the items. If got error when scheduling one of the items,
	// the error should be put in the corresponding ScheduleResult and the
	// returning error of this function should be nil
	Schedule([]*ScheduleItem) ([]*ScheduleResult, error)
	// Stop the job specified by ID
	Stop(id string) error
}
