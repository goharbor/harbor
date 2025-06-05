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
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/task"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
)

type schedulerTestSuite struct {
	suite.Suite
	ctx       context.Context
	scheduler *scheduler
	dao       *mockDAO
	execMgr   *tasktesting.ExecutionManager
	taskMgr   *tasktesting.Manager
}

func (s *schedulerTestSuite) SetupTest() {
	registry = map[string]CallbackFunc{}
	err := RegisterCallbackFunc("callback", func(context.Context, string) error { return nil })
	s.Require().Nil(err)

	s.ctx = orm.NewContext(nil, &ormtesting.FakeOrmer{})
	s.dao = &mockDAO{}
	s.execMgr = &tasktesting.ExecutionManager{}
	s.taskMgr = &tasktesting.Manager{}

	s.scheduler = &scheduler{
		dao:     s.dao,
		execMgr: s.execMgr,
		taskMgr: s.taskMgr,
	}
}

func (s *schedulerTestSuite) TestSchedule() {
	// empty vendor type
	extras := make(map[string]any)
	id, err := s.scheduler.Schedule(s.ctx, "", 0, "", "0 * * * * *", "callback", nil, extras)
	s.NotNil(err)

	// invalid cron
	id, err = s.scheduler.Schedule(s.ctx, "vendor", 1, "", "", "callback", nil, extras)
	s.NotNil(err)

	// callback function not exist
	id, err = s.scheduler.Schedule(s.ctx, "vendor", 1, "", "0 * * * * *", "not-exist", nil, extras)
	s.NotNil(err)

	// failed to submit to jobservice
	s.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Task{
		ID:          1,
		ExecutionID: 1,
		Status:      job.ErrorStatus.String(),
	}, nil)
	s.taskMgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	_, err = s.scheduler.Schedule(s.ctx, "vendor", 1, "", "0 * * * * *", "callback", "param", extras)
	s.Require().NotNil(err)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
	s.taskMgr.AssertExpectations(s.T())

	// reset mocks
	s.SetupTest()

	// pass
	s.execMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.dao.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	s.taskMgr.On("Get", mock.Anything, mock.Anything).Return(&task.Task{
		ID:          1,
		ExecutionID: 1,
		Status:      job.SuccessStatus.String(),
	}, nil)
	id, err = s.scheduler.Schedule(s.ctx, "vendor", 1, "", "0 * * * * *", "callback", "param", extras)
	s.Require().Nil(err)
	s.Equal(int64(1), id)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
	s.taskMgr.AssertExpectations(s.T())
}

func (s *schedulerTestSuite) TestUnScheduleByID() {
	// the execution isn't stopped
	s.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID: 1,
		},
	}, nil)
	s.execMgr.On("StopAndWait", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("error"))
	err := s.scheduler.UnScheduleByID(nil, 1)
	s.NotNil(err)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())

	// reset mocks
	s.SetupTest()

	// pass
	s.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID: 1,
		},
	}, nil)
	s.execMgr.On("StopAndWait", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.dao.On("Delete", mock.Anything, mock.Anything).Return(nil)
	s.execMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err = s.scheduler.UnScheduleByID(nil, 1)
	s.Nil(err)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
}

func (s *schedulerTestSuite) TestUnScheduleByVendor() {
	s.dao.On("List", mock.Anything, mock.Anything).Return([]*schedule{
		{
			ID: 1,
		},
	}, nil)
	s.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID: 1,
		},
	}, nil)
	s.execMgr.On("StopAndWait", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	s.dao.On("Delete", mock.Anything, mock.Anything).Return(nil)
	s.execMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err := s.scheduler.UnScheduleByVendor(nil, "vendor", 1)
	s.Nil(err)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
}

func (s *schedulerTestSuite) TestGetSchedule() {
	// no execution for the schedule
	s.dao.On("Get", mock.Anything, mock.Anything).Return(&schedule{
		ID:         1,
		VendorType: "vendor",
		VendorID:   1,
		CRON:       "0 * * * * *",
	}, nil)
	s.execMgr.On("List", mock.Anything, mock.Anything).Return(nil, nil)
	schd, err := s.scheduler.GetSchedule(s.ctx, 1)
	s.Require().Nil(err)
	s.Equal("0 * * * * *", schd.CRON)
	s.Equal(job.ErrorStatus.String(), schd.Status)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())

	// reset mocks
	s.SetupTest()

	// pass
	s.dao.On("Get", mock.Anything, mock.Anything).Return(&schedule{
		ID:         1,
		VendorType: "vendor",
		VendorID:   1,
		CRON:       "0 * * * * *",
	}, nil)
	s.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID:     1,
			Status: job.SuccessStatus.String(),
		},
	}, nil)
	schd, err = s.scheduler.GetSchedule(s.ctx, 1)
	s.Require().Nil(err)
	s.Equal("0 * * * * *", schd.CRON)
	s.Equal(job.SuccessStatus.String(), schd.Status)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
}

func (s *schedulerTestSuite) TestListSchedules() {
	s.dao.On("List", mock.Anything, mock.Anything).Return([]*schedule{
		{
			ID:         1,
			VendorType: "vendor",
			VendorID:   1,
			CRON:       "0 * * * * *",
		},
	}, nil)
	s.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task.Execution{
		{
			ID:     1,
			Status: job.SuccessStatus.String(),
		},
	}, nil)
	schds, err := s.scheduler.ListSchedules(nil, nil)
	s.Require().Nil(err)
	s.Require().Len(schds, 1)
	s.Equal("0 * * * * *", schds[0].CRON)
	s.Equal(job.SuccessStatus.String(), schds[0].Status)
	s.dao.AssertExpectations(s.T())
	s.execMgr.AssertExpectations(s.T())
}

func TestScheduler(t *testing.T) {
	s := &schedulerTestSuite{}
	suite.Run(t, s)
}
