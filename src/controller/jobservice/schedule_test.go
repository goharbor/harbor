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
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/controller/purge"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/queuestatus"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/testing/mock"
	queueStatusMock "github.com/goharbor/harbor/src/testing/pkg/queuestatus"
	testingScheduler "github.com/goharbor/harbor/src/testing/pkg/scheduler"
)

type ScheduleTestSuite struct {
	suite.Suite
	scheduler      *testingScheduler.Scheduler
	ctl            SchedulerController
	queueStatusMgr queuestatus.Manager
}

func (s *ScheduleTestSuite) SetupSuite() {
	s.scheduler = &testingScheduler.Scheduler{}
	s.queueStatusMgr = &queueStatusMock.Manager{}
	s.ctl = &schedulerController{
		schedulerMgr:   s.scheduler,
		queueStatusMgr: s.queueStatusMgr,
	}
}

func (s *ScheduleTestSuite) TestCreateSchedule() {
	s.scheduler.On("Schedule", mock.Anything, mock.Anything, mock.Anything, mock.Anything,
		mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)

	dataMap := make(map[string]interface{})
	p := purge.JobPolicy{}
	id, err := s.ctl.Create(nil, job.PurgeAuditVendorType, "Daily", "* * * * * *", purge.SchedulerCallback, p, dataMap)
	s.Nil(err)
	s.Equal(int64(1), id)
}

func (s *ScheduleTestSuite) TestDeleteSchedule() {
	s.scheduler.On("UnScheduleByVendor", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.Nil(s.ctl.Delete(nil, job.PurgeAuditVendorType))
}

func (s *ScheduleTestSuite) TestGetSchedule() {
	s.scheduler.On("ListSchedules", mock.Anything, mock.Anything).Return([]*scheduler.Schedule{
		{
			ID:         1,
			VendorType: job.PurgeAuditVendorType,
		},
	}, nil).Once()

	schedule, err := s.ctl.Get(nil, job.PurgeAuditVendorType)
	s.Nil(err)
	s.Equal(job.PurgeAuditVendorType, schedule.VendorType)
}

func (s *ScheduleTestSuite) TestListSchedule() {
	mock.OnAnything(s.scheduler, "ListSchedules").Return([]*scheduler.Schedule{
		{ID: 1, VendorType: "GARBAGE_COLLECTION", CRON: "0 0 0 * * *", ExtraAttrs: map[string]interface{}{"args": "sample args"}}}, nil).Once()
	schedules, err := s.scheduler.ListSchedules(nil, nil)
	s.Assert().Nil(err)
	s.Assert().Equal(1, len(schedules))
	s.Assert().Equal(schedules[0].VendorType, "GARBAGE_COLLECTION")
	s.Assert().Equal(schedules[0].ID, int64(1))
}

func (s *ScheduleTestSuite) TestSchedulerStatus() {
	mock.OnAnything(s.queueStatusMgr, "AllJobTypeStatus").Return(map[string]bool{"SCHEDULER": true}, nil).Once()
	result, err := s.ctl.Paused(context.Background())
	s.Assert().Nil(err)
	s.Assert().True(result)
}

func (s *ScheduleTestSuite) TestCountSchedule() {
	mock.OnAnything(s.scheduler, "CountSchedules").Return(int64(1), nil).Once()
	count, err := s.ctl.Count(context.Background(), nil)
	s.Assert().Nil(err)
	s.Assert().Equal(int64(1), count)
}

func TestScheduleTestSuite(t *testing.T) {
	suite.Run(t, &ScheduleTestSuite{})
}
