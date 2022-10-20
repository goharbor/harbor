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

package jobmonitor

import (
	"testing"
	"time"

	"github.com/goharbor/harbor/src/pkg/queuestatus"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	queueStatusMock "github.com/goharbor/harbor/src/testing/pkg/queuestatus"

	"github.com/gocraft/work"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/jobmonitor"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	monitorMock "github.com/goharbor/harbor/src/testing/pkg/jobmonitor"
	taskMock "github.com/goharbor/harbor/src/testing/pkg/task"
)

type JobServiceMonitorTestSuite struct {
	suite.Suite
	jmClient           jobmonitor.JobServiceMonitorClient
	poolManager        jobmonitor.PoolManager
	workerManager      jobmonitor.WorkerManager
	monitController    MonitorController
	taskManager        task.Manager
	queueStatusManager queuestatus.Manager
	sch                scheduler.Scheduler
	redisClient        jobmonitor.RedisClient
}

func (s *JobServiceMonitorTestSuite) SetupSuite() {
	s.jmClient = &monitorMock.JobServiceMonitorClient{}
	s.poolManager = &monitorMock.PoolManager{}
	s.workerManager = jobmonitor.NewWorkerManager()
	s.taskManager = &taskMock.Manager{}
	s.redisClient = &monitorMock.RedisClient{}
	s.queueStatusManager = &queueStatusMock.Manager{}
	s.monitController = &monitorController{
		poolManager:        s.poolManager,
		workerManager:      s.workerManager,
		taskManager:        s.taskManager,
		queueStatusManager: s.queueStatusManager,
		monitorClient: func() (jobmonitor.JobServiceMonitorClient, error) {
			return s.jmClient, nil
		},
		jobServiceRedisClient: func() (jobmonitor.RedisClient, error) {
			return s.redisClient, nil
		},
	}
}

func (s *JobServiceMonitorTestSuite) TearDownSuite() {
}

func (s *JobServiceMonitorTestSuite) TestListPool() {
	mock.OnAnything(s.poolManager, "List").Return([]*jobmonitor.WorkerPool{
		{
			ID: "1", PID: 1, StartAt: time.Now().Unix(), Concurrency: 10,
		},
	}, nil)
	pools, err := s.poolManager.List(nil, s.jmClient)
	s.Assert().Nil(err)
	s.Assert().Equal(1, len(pools))
}

func (s *JobServiceMonitorTestSuite) TestListWorker() {
	mock.OnAnything(s.jmClient, "WorkerObservations").Return([]*work.WorkerObservation{
		{WorkerID: "abc", IsBusy: true, JobName: "test", JobID: "1", ArgsJSON: "{\"sample\":\"sample args\"}"},
	}, nil)
	mock.OnAnything(s.jmClient, "WorkerPoolHeartbeats").Return([]*work.WorkerPoolHeartbeat{
		{WorkerPoolID: "1", Pid: 1, StartedAt: time.Now().Unix(), Concurrency: 10, WorkerIDs: []string{"abc"}},
	}, nil)
	workers, err := s.monitController.ListWorkers(nil, "1")
	s.Assert().Nil(err)
	s.Assert().Equal(1, len(workers))
}

func (s *JobServiceMonitorTestSuite) TestStopRunningJob() {
	mock.OnAnything(s.jmClient, "WorkerObservations").Return([]*work.WorkerObservation{
		{WorkerID: "abc", IsBusy: true, JobName: "test", JobID: "1", ArgsJSON: "{\"sample\":\"sample args\"}"},
	}, nil)
	mock.OnAnything(s.jmClient, "WorkerPoolHeartbeats").Return([]*work.WorkerPoolHeartbeat{
		{WorkerPoolID: "1", Pid: 1, StartedAt: time.Now().Unix(), Concurrency: 10, WorkerIDs: []string{"abc"}},
	}, nil)
	mock.OnAnything(s.taskManager, "List").Return([]*task.Task{{ID: 1, VendorType: "GARBAGE_COLLECTION"}}, nil)
	mock.OnAnything(s.taskManager, "Stop").Return(nil)
	err := s.monitController.StopRunningJobs(nil, "1")
	s.Assert().Nil(err)
}

func (s *JobServiceMonitorTestSuite) TestListQueue() {
	mock.OnAnything(s.jmClient, "Queues").Return([]*work.Queue{
		{JobName: "GARBAGE_COLLECTION", Count: 100, Latency: 10000}}, nil)
	mock.OnAnything(s.queueStatusManager, "AllJobTypeStatus").Return(map[string]bool{"GARBAGE_COLLECTION": false}, nil).Once()
	queues, err := s.monitController.ListQueues(nil)
	s.Assert().Nil(err)
	s.Assert().Equal(1, len(queues))
	s.Assert().Equal("GARBAGE_COLLECTION", queues[0].JobType)
	s.Assert().False(queues[0].Paused)
}

func (s *JobServiceMonitorTestSuite) TestPauseJob() {
	mock.OnAnything(s.redisClient, "PauseJob").Return(nil).Once()
	err := s.monitController.PauseJobQueues(nil, "GARBAGE_COLLECTION")
	s.Assert().Nil(err)
	mock.OnAnything(s.redisClient, "UnpauseJob").Return(nil).Once()
	err = s.monitController.ResumeJobQueues(nil, "GARBAGE_COLLECTION")
	s.Assert().Nil(err)
}

func TestJobServiceMonitorTestSuite(t *testing.T) {
	suite.Run(t, &JobServiceMonitorTestSuite{})
}
