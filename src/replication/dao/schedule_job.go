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
	"time"

	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/replication/dao/models"
)

// ScheduleJob is the DAO for schedule job
var ScheduleJob ScheduleJobDAO = &scheduleJobDAO{}

// ScheduleJobDAO ...
type ScheduleJobDAO interface {
	Add(*models.ScheduleJob) (int64, error)
	Get(int64) (*models.ScheduleJob, error)
	Update(*models.ScheduleJob, ...string) error
	Delete(int64) error
	List(...*models.ScheduleJobQuery) ([]*models.ScheduleJob, error)
}

type scheduleJobDAO struct{}

func (s *scheduleJobDAO) Add(sj *models.ScheduleJob) (int64, error) {
	now := time.Now()
	sj.CreationTime = now
	sj.UpdateTime = now
	return dao.GetOrmer().Insert(sj)
}

func (s *scheduleJobDAO) Get(id int64) (*models.ScheduleJob, error) {
	sj := &models.ScheduleJob{
		ID: id,
	}
	if err := dao.GetOrmer().Read(sj); err != nil {
		if err == orm.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return sj, nil
}

func (s *scheduleJobDAO) Update(sj *models.ScheduleJob, props ...string) error {
	if sj.UpdateTime.IsZero() {
		now := time.Now()
		sj.UpdateTime = now
		if len(props) > 0 {
			props = append(props, "UpdateTime")
		}
	}

	_, err := dao.GetOrmer().Update(sj, props...)
	return err
}

func (s *scheduleJobDAO) Delete(id int64) error {
	_, err := dao.GetOrmer().Delete(&models.ScheduleJob{
		ID: id,
	})
	return err
}

func (s *scheduleJobDAO) List(query ...*models.ScheduleJobQuery) ([]*models.ScheduleJob, error) {
	qs := dao.GetOrmer().QueryTable(&models.ScheduleJob{})
	if len(query) > 0 && query[0] != nil {
		if query[0].PolicyID > 0 {
			qs = qs.Filter("PolicyID", query[0].PolicyID)
		}
	}
	sjs := []*models.ScheduleJob{}
	_, err := qs.All(&sjs)
	if err != nil {
		return nil, err
	}
	return sjs, nil
}
