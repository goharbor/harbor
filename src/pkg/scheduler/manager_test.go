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
	"testing"

	"github.com/goharbor/harbor/src/pkg/scheduler/model"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var mgr *manager

type fakeScheduleDao struct {
	schedules []*model.Schedule
	mock.Mock
}

func (f *fakeScheduleDao) Create(*model.Schedule) (int64, error) {
	f.Called()
	return 1, nil
}
func (f *fakeScheduleDao) Update(*model.Schedule, ...string) error {
	f.Called()
	return nil
}
func (f *fakeScheduleDao) Delete(int64) error {
	f.Called()
	return nil
}
func (f *fakeScheduleDao) Get(int64) (*model.Schedule, error) {
	f.Called()
	return nil, nil
}
func (f *fakeScheduleDao) List(query ...*model.ScheduleQuery) ([]*model.Schedule, error) {
	f.Called()
	if len(query) == 0 || query[0] == nil {
		return f.schedules, nil
	}
	result := []*model.Schedule{}
	for _, sch := range f.schedules {
		if sch.JobID == query[0].JobID {
			result = append(result, sch)
		}
	}
	return result, nil
}

type managerTestSuite struct {
	suite.Suite
}

func (m *managerTestSuite) SetupTest() {
	// recreate schedule manager
	mgr = &manager{
		scheduleDao: &fakeScheduleDao{},
	}
}

func (m *managerTestSuite) TestCreate() {
	t := m.T()
	mgr.scheduleDao.(*fakeScheduleDao).On("Create", mock.Anything)
	mgr.Create(nil)
	mgr.scheduleDao.(*fakeScheduleDao).AssertCalled(t, "Create")
}

func (m *managerTestSuite) TestUpdate() {
	t := m.T()
	mgr.scheduleDao.(*fakeScheduleDao).On("Update", mock.Anything)
	mgr.Update(nil)
	mgr.scheduleDao.(*fakeScheduleDao).AssertCalled(t, "Update")
}

func (m *managerTestSuite) TestDelete() {
	t := m.T()
	mgr.scheduleDao.(*fakeScheduleDao).On("Delete", mock.Anything)
	mgr.Delete(1)
	mgr.scheduleDao.(*fakeScheduleDao).AssertCalled(t, "Delete")
}

func (m *managerTestSuite) TestGet() {
	t := m.T()
	mgr.scheduleDao.(*fakeScheduleDao).On("Get", mock.Anything)
	mgr.Get(1)
	mgr.scheduleDao.(*fakeScheduleDao).AssertCalled(t, "Get")
}

func (m *managerTestSuite) TestList() {
	t := m.T()
	mgr.scheduleDao.(*fakeScheduleDao).On("List", mock.Anything)
	mgr.List(nil)
	mgr.scheduleDao.(*fakeScheduleDao).AssertCalled(t, "List")
}

func TestManager(t *testing.T) {
	suite.Run(t, &managerTestSuite{})
}
