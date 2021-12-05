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

package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	taskdao "github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	scantesting "github.com/goharbor/harbor/src/testing/controller/scan"
	scannertesting "github.com/goharbor/harbor/src/testing/controller/scanner"
	ormtesting "github.com/goharbor/harbor/src/testing/lib/orm"
	"github.com/goharbor/harbor/src/testing/mock"
	schedulertesting "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	tasktesting "github.com/goharbor/harbor/src/testing/pkg/task"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
)

type ScanAllTestSuite struct {
	htesting.Suite

	execMgr    *tasktesting.ExecutionManager
	scanCtl    *scantesting.Controller
	scannerCtl *scannertesting.Controller
	scheduler  *schedulertesting.Scheduler

	execution *task.Execution
	schedule  *scheduler.Schedule
}

func (suite *ScanAllTestSuite) SetupSuite() {
	suite.execution = &task.Execution{
		Status: "Running",
		Metrics: &taskdao.Metrics{
			TaskCount:          10,
			SuccessTaskCount:   5,
			ErrorTaskCount:     0,
			PendingTaskCount:   4,
			RunningTaskCount:   1,
			ScheduledTaskCount: 0,
			StoppedTaskCount:   0,
		},
	}

	suite.schedule = &scheduler.Schedule{
		ID:           1,
		VendorType:   "vendor_type",
		CRONType:     "Daily",
		CRON:         "0 0 0 * * *",
		Status:       "Running",
		CreationTime: time.Now(),
		UpdateTime:   time.Now(),
	}

	suite.execMgr = &tasktesting.ExecutionManager{}
	suite.scanCtl = &scantesting.Controller{}
	suite.scannerCtl = &scannertesting.Controller{}
	suite.scheduler = &schedulertesting.Scheduler{}

	suite.Config = &restapi.Config{
		ScanAllAPI: &scanAllAPI{
			execMgr:    suite.execMgr,
			scanCtl:    suite.scanCtl,
			scannerCtl: suite.scannerCtl,
			scheduler:  suite.scheduler,
			makeCtx:    func() context.Context { return orm.NewContext(nil, &ormtesting.FakeOrmer{}) },
		},
	}

	suite.Suite.SetupSuite()
}

func (suite *ScanAllTestSuite) TestAuthorization() {
	newBody := func(body interface{}) io.Reader {
		if body == nil {
			return nil
		}

		buf, err := json.Marshal(body)
		suite.Require().NoError(err)
		return bytes.NewBuffer(buf)
	}

	schedule := models.Schedule{
		Schedule: &models.ScheduleObj{Type: "Manual"},
	}

	reqs := []struct {
		method string
		url    string
		body   interface{}
	}{
		{http.MethodGet, "/scans/all/metrics", nil},
		{http.MethodGet, "/scans/schedule/metrics", nil},
		{http.MethodGet, "/system/scanAll/schedule", nil},
		{http.MethodPut, "/system/scanAll/schedule", schedule},
		{http.MethodPost, "/system/scanAll/schedule", schedule},
	}
	for _, req := range reqs {
		{
			// authorized required
			suite.Security.On("IsAuthenticated").Return(false).Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(401, res.StatusCode)
		}

		{
			// system admin required
			suite.Security.On("IsAuthenticated").Return(true).Once()
			suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(false).Once()
			suite.Security.On("GetUsername").Return("username").Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(403, res.StatusCode)
		}

		{
			// default scanner required
			suite.Security.On("IsAuthenticated").Return(true).Once()
			suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
			mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return(nil, nil).Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(412, res.StatusCode)
		}

		{
			// default scanner required failed
			suite.Security.On("IsAuthenticated").Return(true).Once()
			suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Once()
			mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return(nil, fmt.Errorf("failed")).Once()

			res, err := suite.DoReq(req.method, req.url, newBody(req.body))
			suite.NoError(err)
			suite.Equal(500, res.StatusCode)
		}
	}
}

func (suite *ScanAllTestSuite) TestGetLatestScanAllMetrics() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)
	mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{{ID: int64(1)}}, nil).Times(times)

	{
		// get scan all execution failed
		mock.OnAnything(suite.execMgr, "List").Return(nil, fmt.Errorf("failed to list executions")).Once()

		res, err := suite.Get("/scans/all/metrics")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scan all execution not found
		mock.OnAnything(suite.execMgr, "List").Return(nil, nil).Once()

		var stats map[string]interface{}
		res, err := suite.GetJSON("/scans/all/metrics", &stats)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Contains(stats, "ongoing")
	}

	{
		// scan all execution found
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{suite.execution}, nil).Once()

		var stats models.Stats
		res, err := suite.GetJSON("/scans/all/metrics", &stats)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.True(stats.Ongoing)
	}
}

func (suite *ScanAllTestSuite) TestGetLatestScheduledScanAllMetrics() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)
	mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{{ID: int64(1)}}, nil).Times(times)

	{
		// get scan all execution failed
		mock.OnAnything(suite.execMgr, "List").Return(nil, fmt.Errorf("failed to list executions")).Once()

		res, err := suite.Get("/scans/schedule/metrics")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// scan all execution not found
		mock.OnAnything(suite.execMgr, "List").Return(nil, nil).Once()

		var stats map[string]interface{}
		res, err := suite.GetJSON("/scans/schedule/metrics", &stats)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Contains(stats, "ongoing")
	}

	{
		// scan all execution found
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{suite.execution}, nil).Once()

		var stats models.Stats
		res, err := suite.GetJSON("/scans/schedule/metrics", &stats)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.True(stats.Ongoing)
	}
}

func (suite *ScanAllTestSuite) TestStopScanAll() {
	times := 3
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)
	mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{{ID: int64(1)}}, nil).Times(times)

	{
		// create stop scan all but get latest scan all execution failed
		mock.OnAnything(suite.execMgr, "List").Return(nil, fmt.Errorf("list executions failed")).Once()

		res, err := suite.Post("/system/scanAll/stop", nil)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// create stop scan all but no latest scan all execution
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{}, nil).Once()

		res, err := suite.Post("/system/scanAll/stop", nil)
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		// successfully stop scan all
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{suite.execution}, nil).Once()
		mock.OnAnything(suite.execMgr, "Stop").Return(nil).Once()

		res, err := suite.Post("/system/scanAll/stop", nil)
		suite.NoError(err)
		suite.Equal(202, res.StatusCode)
	}
}

func (suite *ScanAllTestSuite) TestCreateScanAllSchedule() {
	times := 11
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)
	mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{{ID: int64(1)}}, nil).Times(times)

	{
		// create scan all schedule no body
		res, err := suite.Post("/system/scanAll/schedule", nil)
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// create scan all schedule with bad body
		res, err := suite.Post("/system/scanAll/schedule", bytes.NewBuffer([]byte("bad body")))
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// create scan all schedule with ScheduleNone
		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleNone}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(201, res.StatusCode)
	}

	{
		// create scan all schedule with ScheduleManual but get latest scan all execution failed
		mock.OnAnything(suite.execMgr, "List").Return(nil, fmt.Errorf("list executions failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleManual}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// create scan all schedule with ScheduleManual but a previous scan all job aleady exits
		mock.OnAnything(suite.execMgr, "List").Return([]*task.Execution{suite.execution}, nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleManual}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(409, res.StatusCode)
	}

	{
		// create scan all schedule with ScheduleManual no previous scan all job exits
		mock.OnAnything(suite.execMgr, "List").Return(nil, nil).Once()
		mock.OnAnything(suite.scanCtl, "ScanAll").Return(int64(1), nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleManual}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(201, res.StatusCode)
	}

	{
		// create scan all schedule with ScheduleManual but scan all failed
		mock.OnAnything(suite.execMgr, "List").Return(nil, nil).Once()
		mock.OnAnything(suite.scanCtl, "ScanAll").Return(int64(0), fmt.Errorf("scan all failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleManual}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// create scan all schedule with periodic but get latest schedule failed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return(nil, fmt.Errorf("get schedule failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleDaily, Cron: "0 0 0 * * *"}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// create scan all schedule with periodic but schedule areadly exists
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleDaily, Cron: "0 0 0 * * *"}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(412, res.StatusCode)
	}

	{
		// create scan all schedule with periodic but create schedule failed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return(nil, nil).Once()
		mock.OnAnything(suite.scheduler, "Schedule").Return(int64(0), fmt.Errorf("create schedule failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleDaily, Cron: "0 0 0 * * *"}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// create scan all schedule with periodic
		mock.OnAnything(suite.scheduler, "ListSchedules").Return(nil, nil).Once()
		mock.OnAnything(suite.scheduler, "Schedule").Return(int64(1), nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleDaily, Cron: "0 0 0 * * *"}}
		res, err := suite.PostJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(201, res.StatusCode)
	}
}

func (suite *ScanAllTestSuite) TestUpdateScanAllSchedule() {
	times := 11
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)
	mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{{ID: int64(1)}}, nil).Times(times)

	{
		// update scan all schedule no body
		res, err := suite.Put("/system/scanAll/schedule", nil)
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// update scan all schedule with bad body
		res, err := suite.Put("/system/scanAll/schedule", bytes.NewBuffer([]byte("bad body")))
		suite.NoError(err)
		suite.Equal(422, res.StatusCode)
	}

	{
		// update scan all schedule with ScheduleManual
		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleManual}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(400, res.StatusCode)
	}

	{
		// update scan all schedule but get schedule failed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return(nil, fmt.Errorf("get schedule failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleNone}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// update scan all schedule with ScheduleNone when no schedule found
		mock.OnAnything(suite.scheduler, "ListSchedules").Return(nil, nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleNone}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// update scan all schedule with ScheduleNone and unschedule failed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()
		mock.OnAnything(suite.scheduler, "UnScheduleByID").Return(fmt.Errorf("unschedule failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleNone}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// update scan all schedule with ScheduleNone successfully
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()
		mock.OnAnything(suite.scheduler, "UnScheduleByID").Return(nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleNone}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// update scan all schedule with periodic but schedule not changed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleDaily, Cron: "0 0 0 * * *"}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// update scan all schedule with periodic and schedule changed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()
		mock.OnAnything(suite.scheduler, "UnScheduleByID").Return(nil).Once()
		mock.OnAnything(suite.scheduler, "Schedule").Return(int64(1), nil).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleCustom, Cron: "0 1 0 * * *"}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// update scan all schedule with periodic and schedule changed, but unschedule old schedule failed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()
		mock.OnAnything(suite.scheduler, "UnScheduleByID").Return(fmt.Errorf("unschedule failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleCustom, Cron: "0 1 0 * * *"}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// update scan all schedule with periodic and schedule changed, but creat new schedule failed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()
		mock.OnAnything(suite.scheduler, "UnScheduleByID").Return(nil).Once()
		mock.OnAnything(suite.scheduler, "Schedule").Return(int64(0), fmt.Errorf("create schedule failed")).Once()

		body := models.Schedule{Schedule: &models.ScheduleObj{Type: ScheduleCustom, Cron: "0 1 0 * * *"}}
		res, err := suite.PutJSON("/system/scanAll/schedule", body)
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}
}

func (suite *ScanAllTestSuite) TestGetScanAllSchedule() {
	times := 4
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true).Times(times)
	mock.OnAnything(suite.scannerCtl, "ListRegistrations").Return([]*scanner.Registration{{ID: int64(1)}}, nil).Times(times)

	{
		// get schedule failed
		mock.OnAnything(suite.scheduler, "ListSchedules").Return(nil, fmt.Errorf("get schedule failed")).Once()

		res, err := suite.Get("/system/scanAll/schedule")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}

	{
		// schedule not found
		mock.OnAnything(suite.scheduler, "ListSchedules").Return(nil, nil).Once()

		res, err := suite.Get("/system/scanAll/schedule")
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
	}

	{
		// schedule found
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule}, nil).Once()

		var schedule models.Schedule
		res, err := suite.GetJSON("/system/scanAll/schedule", &schedule)
		suite.NoError(err)
		suite.Equal(200, res.StatusCode)
		suite.Equal(suite.schedule.CRONType, schedule.Schedule.Type)
		suite.Equal(suite.schedule.CRON, schedule.Schedule.Cron)
	}

	{
		// schedule found more than one
		mock.OnAnything(suite.scheduler, "ListSchedules").Return([]*scheduler.Schedule{suite.schedule, suite.schedule}, nil).Once()

		res, err := suite.Get("/system/scanAll/schedule")
		suite.NoError(err)
		suite.Equal(500, res.StatusCode)
	}
}

func TestScanAllTestSuite(t *testing.T) {
	suite.Run(t, &ScanAllTestSuite{})
}
