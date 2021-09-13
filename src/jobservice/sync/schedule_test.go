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

package sync

import (
	"context"
	"sync"
	"testing"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/env"
	"github.com/goharbor/harbor/src/jobservice/mgt"
	"github.com/goharbor/harbor/src/jobservice/period"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/scheduler"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	ts "github.com/goharbor/harbor/src/testing/pkg/scheduler"
	tt "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
)

// WorkerTestSuite is test suite for testing sync.Worker.
type WorkerTestSuite struct {
	suite.Suite

	worker *Worker
}

// TestWorker is the entry method of WorkerTestSuite.
func TestWorker(t *testing.T) {
	suite.Run(t, &WorkerTestSuite{})
}

// SetupSuite sets up suite.
func (suite *WorkerTestSuite) SetupSuite() {
	sysContext := context.TODO()

	dao.PrepareTestForPostgresSQL()

	getPolicies := func() ([]*period.Policy, error) {
		return []*period.Policy{
			// Dirty data in js datastore.
			{
				ID:         "8ff2aabb977077b84b4d5f1b",
				JobName:    scheduler.JobNameScheduler,
				CronSpec:   "0 0 0 * * 0",
				WebHookURL: "http://core:8080/service/notifications/tasks/250",
				NumericID:  1630667250,
			},
		}, nil
	}

	// Mock methods
	//
	tss := &ts.Scheduler{}
	tss.On("ListSchedules", mock.Anything, mock.Anything).Return([]*scheduler.Schedule{
		{
			ID:   550,
			CRON: "0 0 0 * * *",
		},
	}, nil)

	// The missing schedule in database.
	tte := &tt.ExecutionManager{}
	tte.On("List", mock.Anything, &q.Query{
		Keywords: map[string]interface{}{
			"vendor_type": scheduler.JobNameScheduler,
			"vendor_id":   (int64)(550),
		},
	}).Return([]*task.Execution{
		{
			ID: 1550,
		},
	}, nil)

	ttm := &tt.Manager{}
	ttm.On("List", mock.Anything, &q.Query{
		Keywords: map[string]interface{}{
			"execution_id": (int64)(1550),
		},
	}).Return([]*task.Task{
		{
			ID:          2550,
			ExecutionID: 1550,
			JobID:       "f754ccdd123664b2acb971d9",
		},
	}, nil)

	pms := &period.MockScheduler{}
	pms.On("Schedule", &period.Policy{
		ID:         "f754ccdd123664b2acb971d9",
		JobName:    scheduler.JobNameScheduler,
		CronSpec:   "0 0 0 * * *",
		WebHookURL: "http://core:8080/service/notifications/tasks/2550",
	}).Return((int64)(1630667500), nil)
	pms.On("UnSchedule", "8ff2aabb977077b84b4d5f1b").Return(nil)

	mmm := &mgt.MockManager{}
	mmm.On("SaveJob", mock.Anything).Return(nil)

	suite.worker = New(3).
		WithContext(&env.Context{
			SystemContext: sysContext,
			WG:            &sync.WaitGroup{},
			ErrorChan:     make(chan error, 1),
		}).UseCoreScheduler(tss).
		UseCoreExecutionManager(tte).
		UseCoreTaskManager(ttm).
		UseScheduler(pms).
		UseManager(mmm).
		WithCoreInternalAddr("http://core:8080").
		WithPolicyLoader(getPolicies)
}

// TestStart test Start().
func (suite *WorkerTestSuite) TestStart() {
	err := suite.worker.Start()
	suite.NoError(err, "start worker")
}

// TestRun test Run().
func (suite *WorkerTestSuite) TestRun() {
	err := suite.worker.Run(context.TODO())
	suite.NoError(err, "run worker")
}
