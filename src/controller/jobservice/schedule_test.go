//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package jobservice

import (
	"github.com/goharbor/harbor/src/controller/purge"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/testing/mock"
	testingScheduler "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	"github.com/stretchr/testify/suite"
	"testing"
)

type ScheduleTestSuite struct {
	suite.Suite
	scheduler *testingScheduler.Scheduler
	ctl       SchedulerController
}

func (s *ScheduleTestSuite) SetupSuite() {
	s.scheduler = &testingScheduler.Scheduler{}
	s.ctl = &schedulerController{
		schedulerMgr: s.scheduler,
	}
}

func (s *ScheduleTestSuite) TestCreateSchedule() {
	s.scheduler.On("Schedule", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	dataMap := make(map[string]interface{})
	p := purge.JobPolicy{}
	id, err := s.ctl.Create(nil, purge.VendorType, "Daily", "* * * * * *", purge.SchedulerCallback, p, dataMap)
	s.Nil(err)
	s.Equal(int64(1), id)
}

func (s *ScheduleTestSuite) TestDeleteSchedule() {
	s.scheduler.On("UnScheduleByVendor", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.Nil(s.ctl.Delete(nil, purge.VendorType))
}

func (s *ScheduleTestSuite) TestGetSchedule() {
	s.scheduler.On("ListSchedules", mock.Anything, mock.Anything).Return([]*scheduler.Schedule{
		{
			ID:         1,
			VendorType: purge.VendorType,
		},
	}, nil)

	schedule, err := s.ctl.Get(nil, purge.VendorType)
	s.Nil(err)
	s.Equal(purge.VendorType, schedule.VendorType)
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, &ScheduleTestSuite{})
}
