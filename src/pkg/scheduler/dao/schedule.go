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

package dao

import (
	"errors"
	"fmt"
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/scheduler/model"
)

// ScheduleDao defines the method that a schedule data access model should implement
type ScheduleDao interface {
	Create(*model.Schedule) (int64, error)
	Update(*model.Schedule, ...string) error
	Delete(int64) error
	Get(int64) (*model.Schedule, error)
	List(...*model.ScheduleQuery) ([]*model.Schedule, error)
}

// New returns an instance of the default schedule data access model implementation
func New() ScheduleDao {
	return &scheduleDao{}
}

type scheduleDao struct{}

func (s *scheduleDao) Create(schedule *model.Schedule) (int64, error) {
	if schedule == nil {
		return 0, errors.New("nil schedule")
	}
	now := time.Now()
	schedule.CreationTime = &now
	schedule.UpdateTime = &now
	return dao.GetOrmer().Insert(schedule)
}

func (s *scheduleDao) Update(schedule *model.Schedule, cols ...string) error {
	if schedule == nil {
		return errors.New("nil schedule")
	}
	if schedule.ID <= 0 {
		return fmt.Errorf("invalid ID: %d", schedule.ID)
	}
	now := time.Now()
	schedule.UpdateTime = &now
	_, err := dao.GetOrmer().Update(schedule, cols...)
	return err
}

func (s *scheduleDao) Delete(id int64) error {
	_, err := dao.GetOrmer().Delete(&model.Schedule{
		ID: id,
	})
	return err
}

func (s *scheduleDao) Get(id int64) (*model.Schedule, error) {
	schedule := &model.Schedule{
		ID: id,
	}
	if err := dao.GetOrmer().Read(schedule); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return schedule, nil
}

func (s *scheduleDao) List(query ...*model.ScheduleQuery) ([]*model.Schedule, error) {
	qs := dao.GetOrmer().QueryTable(&model.Schedule{})
	if len(query) > 0 && query[0] != nil {
		if len(query[0].JobID) > 0 {
			qs = qs.Filter("JobID", query[0].JobID)
		}
	}
	schedules := []*model.Schedule{}
	_, err := qs.All(&schedules)
	if err != nil {
		return nil, err
	}
	return schedules, nil
}
