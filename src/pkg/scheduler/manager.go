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
	"github.com/goharbor/harbor/src/pkg/scheduler/dao"
	"github.com/goharbor/harbor/src/pkg/scheduler/model"
)

var (
	// GlobalManager is an instance of the default manager that
	// can be used globally
	GlobalManager = NewManager()
)

// Manager manages the schedule of the scheduler
type Manager interface {
	Create(*model.Schedule) (int64, error)
	Update(*model.Schedule, ...string) error
	Delete(int64) error
	Get(int64) (*model.Schedule, error)
	List(...*model.ScheduleQuery) ([]*model.Schedule, error)
}

// NewManager returns an instance of the default manager
func NewManager() Manager {
	return &manager{
		scheduleDao: dao.New(),
	}
}

type manager struct {
	scheduleDao dao.ScheduleDao
}

func (m *manager) Create(schedule *model.Schedule) (int64, error) {
	return m.scheduleDao.Create(schedule)
}

func (m *manager) Update(schedule *model.Schedule, props ...string) error {
	return m.scheduleDao.Update(schedule, props...)
}

func (m *manager) Delete(id int64) error {
	return m.scheduleDao.Delete(id)
}

func (m *manager) List(query ...*model.ScheduleQuery) ([]*model.Schedule, error) {
	return m.scheduleDao.List(query...)
}

func (m *manager) Get(id int64) (*model.Schedule, error) {
	return m.scheduleDao.Get(id)
}
