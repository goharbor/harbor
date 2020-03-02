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
	"fmt"

	"github.com/goharbor/harbor/src/pkg/scheduler/model"
)

// FakeManager ...
type FakeManager struct {
	idCounter int64
	Schedules []*model.Schedule
}

// Create ...
func (f *FakeManager) Create(schedule *model.Schedule) (int64, error) {
	f.idCounter++
	id := f.idCounter
	schedule.ID = id
	f.Schedules = append(f.Schedules, schedule)
	return id, nil
}

// Update ...
func (f *FakeManager) Update(schedule *model.Schedule, props ...string) error {
	for i, sch := range f.Schedules {
		if sch.ID == schedule.ID {
			f.Schedules[i] = schedule
			return nil
		}
	}
	return fmt.Errorf("the execution %d not found", schedule.ID)
}

// Delete ...
func (f *FakeManager) Delete(id int64) error {
	length := len(f.Schedules)
	for i, sch := range f.Schedules {
		if sch.ID == id {
			f.Schedules = f.Schedules[:i]
			if i != length-1 {
				f.Schedules = append(f.Schedules, f.Schedules[i+1:]...)
			}
			return nil
		}
	}
	return fmt.Errorf("the execution %d not found", id)
}

// Get ...
func (f *FakeManager) Get(id int64) (*model.Schedule, error) {
	for _, sch := range f.Schedules {
		if sch.ID == id {
			return sch, nil
		}
	}
	return nil, fmt.Errorf("the execution %d not found", id)
}

// List ...
func (f *FakeManager) List(...*model.ScheduleQuery) ([]*model.Schedule, error) {
	return f.Schedules, nil
}
