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

package manager

import (
	"fmt"
	"time"

	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	"github.com/goharbor/harbor/src/pkg/task/model"
)

// Manager defines the inferface to interactive with the Task data model
type Manager interface {
	Create(task *model.Task) (int64, error)
	Get(id int64) (*model.Task, error)
	Update(task *model.Task, cols ...string) error
	UpdateStatus(id int64, status string, statusCode int, statusRevision int64) error
	Delete(id int64) error
	AppendCheckInData(id int64, data string) error
	CalculateTaskGroupStatus(groupID int64) (*model.GroupStatus, error)
}

// New creates a new task manager
func New() Manager {
	return &manager{
		dao: dao.New(),
	}
}

type manager struct {
	dao dao.TaskDao
}

func (m *manager) Create(task *model.Task) (int64, error) {
	t, err := task.To()
	if err != nil {
		return 0, err
	}
	return m.dao.Create(t)
}
func (m *manager) Get(id int64) (*model.Task, error) {
	t, err := m.dao.Get(id)
	if err != nil {
		return nil, err
	}
	if t == nil {
		return nil, nil
	}
	task := &model.Task{}
	if err = task.From(t); err != nil {
		return nil, err
	}
	// populate the check in data
	checkIns, err := m.dao.ListCheckInData(task.ID)
	if err != nil {
		return nil, err
	}
	for _, checkIn := range checkIns {
		task.CheckInData = append(task.CheckInData, checkIn.Data)
	}
	return task, nil
}
func (m *manager) Update(task *model.Task, cols ...string) error {
	t, err := task.To()
	if err != nil {
		return err
	}
	return m.dao.Update(t, cols...)
}
func (m *manager) UpdateStatus(id int64, status string, statusCode int, statusRevision int64) error {
	var tm time.Time
	// when the task is in final status, update the endtime,
	// when the task re-runs again, the endtime should be cleared,
	// so set the endtime to null if the task isn't in final status
	if model.IsFinalStatus(status) {
		tm = time.Now()
	}
	return m.dao.UpdateStatus(id, status, statusCode, statusRevision, tm)
}
func (m *manager) Delete(id int64) error {
	// delete the check in data references the task
	checkIns, err := m.dao.ListCheckInData(id)
	if err != nil {
		return err
	}
	for _, checkIn := range checkIns {
		if err = m.dao.DeleteCheckInData(checkIn.ID); err != nil {
			return err
		}
	}
	// delete the task
	return m.dao.Delete(id)
}

// based on the option, append or override the check in data
func (m *manager) AppendCheckInData(id int64, data string) error {
	task, err := m.Get(id)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task %d not found", id)
	}
	now := time.Now()
	// append the check in data
	if task.Options != nil && task.Options.AppendCheckInData {
		_, err = m.dao.CreateCheckInData(&dao.CheckInData{
			TaskID:       id,
			Data:         data,
			CreationTime: now,
			UpdateTime:   now,
		})
		return err
	}
	// override the check in data
	checkIns, err := m.dao.ListCheckInData(id)
	if err != nil {
		return err
	}
	// there is no check in data record yet, create it
	if len(checkIns) == 0 {
		_, err = m.dao.CreateCheckInData(&dao.CheckInData{
			TaskID:       id,
			Data:         data,
			CreationTime: now,
			UpdateTime:   now,
		})
		return err
	}
	// there is already a check in data record, override it
	return m.dao.UpdateCheckInData(&dao.CheckInData{
		ID:         checkIns[0].ID,
		TaskID:     id,
		Data:       data,
		UpdateTime: now,
	}, "Data", "UpdateTime")
}

func (m *manager) CalculateTaskGroupStatus(groupID int64) (*model.GroupStatus, error) {
	scs, err := m.dao.GetGroupStatus(groupID)
	if err != nil {
		return nil, err
	}
	gs := &model.GroupStatus{
		ID: groupID,
	}
	// the group contains no tasks, returning success as the status
	if len(scs) == 0 {
		gs.Status = job.SuccessStatus.String()
		gs.Total = 0
		return gs, nil
	}

	for _, sg := range scs {
		switch sg.Status {
		// merge pending, scheduled and running as running
		case job.PendingStatus.String(),
			job.ScheduledStatus.String(),
			job.RunningStatus.String():
			gs.Total += sg.Count
			gs.Running += sg.Count
		case job.StoppedStatus.String():
			gs.Total += sg.Count
			gs.Stopped += sg.Count
		case job.ErrorStatus.String():
			gs.Total += sg.Count
			gs.Error += sg.Count
		case job.SuccessStatus.String():
			gs.Total += sg.Count
			gs.Success += sg.Count
		}
	}
	if gs.Running > 0 {
		gs.Status = job.RunningStatus.String()
	} else if gs.Stopped > 0 {
		gs.Status = job.StoppedStatus.String()
	} else if gs.Error > 0 {
		gs.Status = job.ErrorStatus.String()
	} else {
		gs.Status = job.SuccessStatus.String()
	}

	// if the status is a final status, calculate the end time
	if model.IsFinalStatus(gs.Status) {
		t, err := m.dao.GetMaxEndTime(groupID)
		if err != nil {
			log.Errorf("failed to get the max end time: %v", err)
		} else {
			gs.EndTime = t
		}
	}
	return gs, nil
}
