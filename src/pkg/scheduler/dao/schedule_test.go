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
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/pkg/scheduler/model"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

var schDao = &scheduleDao{}

type scheduleTestSuite struct {
	suite.Suite
	scheduleID int64
}

func (s *scheduleTestSuite) SetupSuite() {
	dao.PrepareTestForPostgresSQL()
}

func (s *scheduleTestSuite) SetupTest() {
	t := s.T()
	id, err := schDao.Create(&model.Schedule{
		JobID:  "1",
		Status: "pending",
	})
	require.Nil(t, err)
	s.scheduleID = id
}
func (s *scheduleTestSuite) TearDownTest() {
	// clear
	dao.GetOrmer().Raw("delete from schedule").Exec()
}

func (s *scheduleTestSuite) TestCreate() {
	t := s.T()
	// nil schedule
	_, err := schDao.Create(nil)
	require.NotNil(t, err)

	// pass
	_, err = schDao.Create(&model.Schedule{
		JobID: "1",
	})
	require.Nil(t, err)
}

func (s *scheduleTestSuite) TestUpdate() {
	t := s.T()
	// nil schedule
	err := schDao.Update(nil)
	require.NotNil(t, err)

	// invalid ID
	err = schDao.Update(&model.Schedule{})
	require.NotNil(t, err)

	// pass
	err = schDao.Update(&model.Schedule{
		ID:     s.scheduleID,
		Status: "running",
	})
	require.Nil(t, err)
	schedule, err := schDao.Get(s.scheduleID)
	require.Nil(t, err)
	assert.Equal(t, "running", schedule.Status)
}

func (s *scheduleTestSuite) TestDelete() {
	t := s.T()
	err := schDao.Delete(s.scheduleID)
	require.Nil(t, err)
	schedule, err := schDao.Get(s.scheduleID)
	require.Nil(t, err)
	assert.Nil(t, schedule)
}

func (s *scheduleTestSuite) TestGet() {
	t := s.T()
	schedule, err := schDao.Get(s.scheduleID)
	require.Nil(t, err)
	assert.Equal(t, "pending", schedule.Status)
}

func (s *scheduleTestSuite) TestList() {
	t := s.T()
	// nil query
	schedules, err := schDao.List()
	require.Nil(t, err)
	require.Equal(t, 1, len(schedules))
	assert.Equal(t, s.scheduleID, schedules[0].ID)

	// query by job ID
	schedules, err = schDao.List(&model.ScheduleQuery{
		JobID: "1",
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(schedules))
	assert.Equal(t, s.scheduleID, schedules[0].ID)
}

func TestScheduleDao(t *testing.T) {
	suite.Run(t, &scheduleTestSuite{})
}
