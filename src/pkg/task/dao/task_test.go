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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	htesting "github.com/goharbor/harbor/src/testing"
)

type taskDAOTestSuite struct {
	htesting.Suite
	ctx          context.Context
	taskDAO      *taskDAO
	executionDAO *executionDAO
	executionID  int64
	taskID       int64
}

func (t *taskDAOTestSuite) SetupSuite() {
	t.Suite.SetupSuite()
	t.ctx = orm.Context()
	t.taskDAO = &taskDAO{}
	t.executionDAO = &executionDAO{}
}

func (t *taskDAOTestSuite) SetupTest() {
	id, err := t.executionDAO.Create(t.ctx, &Execution{
		VendorType: "test",
		Trigger:    "test",
		ExtraAttrs: "{}",
	})
	t.Require().Nil(err)
	t.executionID = id
	id, err = t.taskDAO.Create(t.ctx, &Task{
		ExecutionID: t.executionID,
		Status:      "success",
		StatusCode:  1,
		ExtraAttrs:  `{"key":"value"}`,
	})
	t.Require().Nil(err)
	t.taskID = id
}

func (t *taskDAOTestSuite) TearDownTest() {
	err := t.taskDAO.Delete(t.ctx, t.taskID)
	t.Nil(err)

	err = t.executionDAO.Delete(t.ctx, t.executionID)
	t.Nil(err)
}

func (t *taskDAOTestSuite) TestCount() {
	count, err := t.taskDAO.Count(t.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID":    t.executionID,
			"ExtraAttrs.key": "value",
		},
	})
	t.Require().Nil(err)
	t.Equal(int64(1), count)

	count, err = t.taskDAO.Count(t.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID":    t.executionID,
			"ExtraAttrs.key": "incorrect-value",
		},
	})
	t.Require().Nil(err)
	t.Equal(int64(0), count)
}

func (t *taskDAOTestSuite) TestList() {
	tasks, err := t.taskDAO.List(t.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID":    t.executionID,
			"ExtraAttrs.key": "value",
		},
	})
	t.Require().Nil(err)
	t.Require().Len(tasks, 1)
	t.Equal(t.taskID, tasks[0].ID)

	tasks, err = t.taskDAO.List(t.ctx, &q.Query{
		Keywords: map[string]interface{}{
			"ExecutionID":    t.executionID,
			"ExtraAttrs.key": "incorrect-value",
		},
	})
	t.Require().Nil(err)
	t.Require().Len(tasks, 0)
}

func (t *taskDAOTestSuite) TestListScanTasksByReportUUID() {
	reportUUID := `7f20b1b9-6117-4a2e-820b-e4cc0401f15e`
	// should not exist if non set
	tasks, err := t.taskDAO.ListScanTasksByReportUUID(t.ctx, reportUUID)
	t.Require().Nil(err)
	t.Require().Len(tasks, 0)
	// create one with report uuid
	taskID, err := t.taskDAO.Create(t.ctx, &Task{
		ExecutionID: t.executionID,
		Status:      "success",
		StatusCode:  1,
		ExtraAttrs:  fmt.Sprintf(`{"report_uuids": ["%s"]}`, reportUUID),
	})
	t.Require().Nil(err)
	defer t.taskDAO.Delete(t.ctx, taskID)
	// should exist as created
	tasks, err = t.taskDAO.ListScanTasksByReportUUID(t.ctx, reportUUID)
	t.Require().Nil(err)
	t.Require().Len(tasks, 1)
	t.Equal(taskID, tasks[0].ID)
}

func (t *taskDAOTestSuite) TestGet() {
	// not exist
	_, err := t.taskDAO.Get(t.ctx, 10000)
	t.Require().NotNil(err)
	t.True(errors.IsNotFoundErr(err))

	// exist
	task, err := t.taskDAO.Get(t.ctx, t.taskID)
	t.Require().Nil(err)
	t.NotNil(task)
}

func (t *taskDAOTestSuite) TestCreate() {
	// reference the non-existing execution
	_, err := t.taskDAO.Create(t.ctx, &Task{
		ExecutionID: 10000,
		Status:      "success",
		StatusCode:  1,
		ExtraAttrs:  "{}",
	})
	t.Require().NotNil(err)
	t.True(errors.IsErr(err, errors.ViolateForeignKeyConstraintCode))

	// reference the existing execution is covered by SetupTest
}

func (t *taskDAOTestSuite) TestUpdate() {
	// not exist
	err := t.taskDAO.Update(t.ctx, &Task{ID: 10000}, "Status")
	t.Require().NotNil(err)
	t.True(errors.IsNotFoundErr(err))

	// exist
	err = t.taskDAO.Update(t.ctx, &Task{
		ID:     t.taskID,
		Status: "failed",
	}, "Status")
	t.Require().Nil(err)
	task, err := t.taskDAO.Get(t.ctx, t.taskID)
	t.Require().Nil(err)
	t.Equal("failed", task.Status)
}

func (t *taskDAOTestSuite) TestUpdateStatus() {
	// update status to running
	status := job.RunningStatus.String()
	statusRevision := time.Now().Unix()
	err := t.taskDAO.UpdateStatus(t.ctx, t.taskID, status, statusRevision)
	t.Require().Nil(err)

	task, err := t.taskDAO.Get(t.ctx, t.taskID)
	t.Require().Nil(err)
	t.Equal(int32(1), task.RunCount)
	t.True(time.Unix(statusRevision, 0).Equal(task.StartTime))
	t.Equal(status, task.Status)
	t.Equal(job.RunningStatus.Code(), task.StatusCode)
	t.Equal(statusRevision, task.StatusRevision)
	t.NotEqual(time.Time{}, task.UpdateTime)
	t.Equal(time.Time{}, task.EndTime)

	// update status to success
	status = job.SuccessStatus.String()
	err = t.taskDAO.UpdateStatus(t.ctx, t.taskID, status, statusRevision)
	t.Require().Nil(err)

	task, err = t.taskDAO.Get(t.ctx, t.taskID)
	t.Require().Nil(err)
	t.Equal(int32(1), task.RunCount)
	t.True(time.Unix(statusRevision, 0).Equal(task.StartTime))
	t.Equal(status, task.Status)
	t.Equal(job.SuccessStatus.Code(), task.StatusCode)
	t.Equal(statusRevision, task.StatusRevision)
	t.NotEqual(time.Time{}, task.EndTime)

	// update status to running again with different revision
	status = job.RunningStatus.String()
	statusRevision = time.Now().Add(1 * time.Second).Unix()
	err = t.taskDAO.UpdateStatus(t.ctx, t.taskID, status, statusRevision)
	t.Require().Nil(err)

	task, err = t.taskDAO.Get(t.ctx, t.taskID)
	t.Require().Nil(err)
	t.Equal(int32(2), task.RunCount)
	t.True(time.Unix(statusRevision, 0).Equal(task.StartTime))
	t.Equal(status, task.Status)
	t.Equal(job.RunningStatus.Code(), task.StatusCode)
	t.Equal(statusRevision, task.StatusRevision)
	t.Equal(time.Time{}, task.EndTime)
}

func (t *taskDAOTestSuite) TestDelete() {
	// not exist
	err := t.taskDAO.Delete(t.ctx, 10000)
	t.Require().NotNil(err)
	t.True(errors.IsNotFoundErr(err))

	// happy pass is covered by TearDownTest
}

func (t *taskDAOTestSuite) TestListStatusCount() {
	scs, err := t.taskDAO.ListStatusCount(t.ctx, t.executionID)
	t.Require().Nil(err)
	t.Require().Len(scs, 1)
	t.Equal("success", scs[0].Status)
	t.Equal(int64(1), scs[0].Count)
}

func (t *taskDAOTestSuite) TestGetMaxEndTime() {
	now := time.Now()
	err := t.taskDAO.Update(t.ctx, &Task{
		ID:      t.taskID,
		EndTime: now,
	}, "EndTime")
	t.Require().Nil(err)
	endTime, err := t.taskDAO.GetMaxEndTime(t.ctx, t.executionID)
	t.Require().Nil(err)
	t.Equal(now.Unix(), endTime.Unix())
}

func (t *taskDAOTestSuite) TestUpdateStatusInBatch() {
	jobIDs := make([]string, 0)
	taskIDs := make([]int64, 0)
	for i := 0; i < 300; i++ {
		jobID := fmt.Sprintf("job-%d", i)
		tid, err := t.taskDAO.Create(t.ctx, &Task{
			JobID:       jobID,
			ExecutionID: t.executionID,
			Status:      "Pending",
			StatusCode:  1,
			ExtraAttrs:  "{}",
		})
		t.Require().Nil(err)
		jobIDs = append(jobIDs, jobID)
		taskIDs = append(taskIDs, tid)
	}

	err := t.taskDAO.UpdateStatusInBatch(t.ctx, jobIDs, "Stopped", 10)
	t.Require().Nil(err)
	for i := 0; i < 300; i++ {
		tasks, err := t.taskDAO.List(t.ctx, &q.Query{
			Keywords: q.KeyWords{"job_id": jobIDs[i]}})
		t.Require().Nil(err)
		t.Require().Len(tasks, 1)
		t.Equal("Stopped", tasks[0].Status)
	}
	for _, taskID := range taskIDs {
		t.taskDAO.Delete(t.ctx, taskID)
	}
}

func (t *taskDAOTestSuite) TestExecutionIDsByVendorAndStatus() {
	tid, err := t.taskDAO.Create(t.ctx, &Task{
		JobID:       "job123",
		ExecutionID: t.executionID,
		Status:      "Pending",
		StatusCode:  1,
		ExtraAttrs:  "{}",
		VendorType:  "MYREPLICATION",
	})
	t.Require().Nil(err)
	exeIDs, err := t.taskDAO.ExecutionIDsByVendorAndStatus(t.ctx, "MYREPLICATION", "Pending")
	t.Require().Nil(err)
	t.Require().Len(exeIDs, 1)
	defer t.taskDAO.Delete(t.ctx, tid)
}

func TestIsValidUUID(t *testing.T) {
	tests := []struct {
		name     string
		uuid     string
		expected bool
	}{
		{"Valid UUID", "7f20b1b9-6117-4a2e-820b-e4cc0401f15f", true},
		{"Invalid UUID - Short", "7f20b1b9-6117-4a2e-820b", false},
		{"Invalid UUID - Long", "7f20b1b9-6117-4a2e-820b-e4cc0401f15f-extra", false},
		{"Invalid UUID - Invalid Characters", "7f20b1b9-6117-4z2e-820b-e4cc0401f15f", false},
		{"Empty String", "", false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isValidUUID(test.uuid)
			if result != test.expected {
				t.Errorf("Expected isValidUUID(%s) to be %t, got %t", test.uuid, test.expected, result)
			}
		})
	}
}

func TestTaskDAOSuite(t *testing.T) {
	suite.Run(t, &taskDAOTestSuite{})
}
