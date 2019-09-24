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
	"time"

	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type taskDaoTestSuite struct {
	suite.Suite
	require *require.Assertions
	assert  *assert.Assertions
	dao     TaskDao
	id      int64
}

func (t *taskDaoTestSuite) SetupSuite() {
	t.require = require.New(t.T())
	t.assert = assert.New(t.T())
	dao.PrepareTestForPostgresSQL()
	t.dao = New()
}

func (t *taskDaoTestSuite) SetupTest() {
	id, err := t.dao.Create(&Task{
		GroupID:        1,
		Status:         job.PendingStatus.String(),
		StatusCode:     job.PendingStatus.Code(),
		StatusRevision: 1,
	})
	t.require.Nil(err)
	t.id = id
}

func (t *taskDaoTestSuite) TearDownTest() {
	err := t.dao.Delete(t.id)
	t.require.Nil(err)
}

func (t *taskDaoTestSuite) TestGet() {
	// get non-existing task
	task, err := t.dao.Get(0)
	t.require.Nil(err)
	t.assert.Nil(task)

	// get the existing task
	task, err = t.dao.Get(t.id)
	t.require.Nil(err)
	t.require.NotNil(task)
	t.assert.Equal(t.id, task.ID)
}

func (t *taskDaoTestSuite) TestUpdate() {
	task, err := t.dao.Get(t.id)
	t.require.Nil(err)
	t.require.NotNil(task)
	t.assert.Equal("", task.JobID)

	err = t.dao.Update(&Task{
		ID:    t.id,
		JobID: "uuid",
	}, "JobID")
	t.require.Nil(err)

	task, err = t.dao.Get(t.id)
	t.require.Nil(err)
	t.require.NotNil(task)
	t.assert.Equal("uuid", task.JobID)
}

func (t *taskDaoTestSuite) TestUpdateStatus() {
	now := time.Now()
	// update the status to success
	err := t.dao.UpdateStatus(t.id, job.SuccessStatus.String(), job.SuccessStatus.Code(), 1, now)
	t.require.Nil(err)
	task, err := t.dao.Get(t.id)
	t.require.Nil(err)
	t.require.NotNil(task)
	t.assert.Equal(job.SuccessStatus.String(), task.Status)
	t.assert.Equal(job.SuccessStatus.Code(), task.StatusCode)
	t.assert.Equal(int64(1), task.StatusRevision)
	t.assert.Equal(now.Unix(), task.EndTime.Unix())

	// update the status to running with the same revision, the updating should take no effect
	err = t.dao.UpdateStatus(t.id, job.RunningStatus.String(), job.RunningStatus.Code(), 1, time.Time{})
	t.require.Nil(err)
	task, err = t.dao.Get(t.id)
	t.require.Nil(err)
	t.require.NotNil(task)
	t.assert.Equal(job.SuccessStatus.String(), task.Status)
	t.assert.Equal(job.SuccessStatus.Code(), task.StatusCode)
	t.assert.Equal(int64(1), task.StatusRevision)
	t.assert.Equal(now.Unix(), task.EndTime.Unix())

	// update the status to running with the different revision
	err = t.dao.UpdateStatus(t.id, job.RunningStatus.String(), job.RunningStatus.Code(), 2, time.Time{})
	t.require.Nil(err)
	task, err = t.dao.Get(t.id)
	t.require.Nil(err)
	t.require.NotNil(task)
	t.assert.Equal(job.RunningStatus.String(), task.Status)
	t.assert.Equal(job.RunningStatus.Code(), task.StatusCode)
	t.assert.Equal(int64(2), task.StatusRevision)
	t.assert.True(task.EndTime.IsZero())
}

func (t *taskDaoTestSuite) TestMethodsOfCheckInData() {
	// create
	now := time.Now()
	id, err := t.dao.CreateCheckInData(&CheckInData{
		TaskID:       t.id,
		Data:         "data",
		CreationTime: now,
		UpdateTime:   now,
	})
	t.require.Nil(err)

	// update
	err = t.dao.UpdateCheckInData(&CheckInData{
		ID:   id,
		Data: "data2",
	}, "Data")
	t.require.Nil(err)

	// list
	checkIds, err := t.dao.ListCheckInData(t.id)
	t.require.Nil(err)
	t.assert.Equal(1, len(checkIds))
	t.assert.Equal(t.id, checkIds[0].TaskID)
	t.assert.Equal("data2", checkIds[0].Data)
	t.assert.Equal(now.Unix(), checkIds[0].CreationTime.Unix())
	t.assert.Equal(now.Unix(), checkIds[0].UpdateTime.Unix())

	// delete
	err = t.dao.Delete(id)
	t.require.Nil(err)
}

func (t *taskDaoTestSuite) TestGetGroupStatus() {
	// group ID: 0
	_, err := t.dao.GetGroupStatus(0)
	t.require.NotNil(err)

	// non-exist group ID
	sgs, err := t.dao.GetGroupStatus(2)
	t.require.Nil(err)
	t.require.Equal(0, len(sgs))

	// valid group ID
	sgs, err = t.dao.GetGroupStatus(1)
	t.require.Nil(err)
	t.require.Equal(1, len(sgs))
	t.assert.Equal(job.PendingStatus.String(), sgs[0].Status)
	t.assert.Equal(int64(1), sgs[0].Count)
}

func (t *taskDaoTestSuite) GetMaxEndTime() {
	// group ID: 0
	_, err := t.dao.GetMaxEndTime(0)
	t.require.NotNil(err)

	// valid group ID
	now := time.Now()
	id2, err := t.dao.Create(&Task{
		GroupID: 1,
		EndTime: now,
	})
	t.require.Nil(err)
	defer func() {
		t.dao.Delete(id2)
	}()
	tm, err := t.dao.GetMaxEndTime(1)
	t.require.Nil(err)
	t.Equal(now, tm)
}

func TestTaskDaoTestSuite(t *testing.T) {
	suite.Run(t, &taskDaoTestSuite{})
}
