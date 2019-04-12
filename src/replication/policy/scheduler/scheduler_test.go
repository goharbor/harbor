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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/dao"
	rep_models "github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TODO share the faked implementation in a separated common package?
// TODO can we use a mock framework?

var (
	uuid           = "uuid"
	policyID int64 = 100
)

type fakedJobserviceClient struct {
	jobData *models.JobData
	stopped bool
}

func (f *fakedJobserviceClient) SubmitJob(jobData *models.JobData) (string, error) {
	f.jobData = jobData
	return uuid, nil
}
func (f *fakedJobserviceClient) GetJobLog(uuid string) ([]byte, error) {
	f.stopped = true
	return nil, nil
}
func (f *fakedJobserviceClient) PostAction(uuid, action string) error {
	f.stopped = true
	return nil
}

type fakedScheduleJobDAO struct {
	idCounter int64
	sjs       map[int64]*rep_models.ScheduleJob
}

func (f *fakedScheduleJobDAO) Add(sj *rep_models.ScheduleJob) (int64, error) {
	if f.sjs == nil {
		f.sjs = make(map[int64]*rep_models.ScheduleJob)
	}
	id := f.idCounter + 1
	sj.ID = id
	f.sjs[id] = sj
	return id, nil
}
func (f *fakedScheduleJobDAO) Get(id int64) (*rep_models.ScheduleJob, error) {
	if f.sjs == nil {
		return nil, nil
	}
	return f.sjs[id], nil
}
func (f *fakedScheduleJobDAO) Update(sj *rep_models.ScheduleJob, props ...string) error {
	err := fmt.Errorf("schedule job %d not found", sj.ID)
	if f.sjs == nil {
		return err
	}
	j, exist := f.sjs[sj.ID]
	if !exist {
		return err
	}
	if len(props) == 0 {
		f.sjs[sj.ID] = sj
		return nil
	}

	for _, prop := range props {
		switch prop {
		case "PolicyID":
			j.PolicyID = sj.PolicyID
		case "JobID":
			j.JobID = sj.JobID
		case "Status":
			j.Status = sj.Status
		case "UpdateTime":
			j.UpdateTime = sj.UpdateTime
		}
	}
	return nil
}
func (f *fakedScheduleJobDAO) Delete(id int64) error {
	if f.sjs == nil {
		return nil
	}
	delete(f.sjs, id)
	return nil
}
func (f *fakedScheduleJobDAO) List(query ...*rep_models.ScheduleJobQuery) ([]*rep_models.ScheduleJob, error) {
	var policyID int64
	if len(query) > 0 {
		policyID = query[0].PolicyID
	}
	sjs := []*rep_models.ScheduleJob{}
	for _, sj := range f.sjs {
		if policyID == 0 {
			sjs = append(sjs, sj)
			continue
		}
		if sj.PolicyID == policyID {
			sjs = append(sjs, sj)
		}
	}
	return sjs, nil
}

func TestSchedule(t *testing.T) {
	config.Config = &config.Configuration{}
	dao.ScheduleJob = &fakedScheduleJobDAO{}
	js := &fakedJobserviceClient{}
	scheduler := NewScheduler(js)
	err := scheduler.Schedule(policyID, "1 * * * *")
	require.Nil(t, err)

	sjs, err := dao.ScheduleJob.List(&rep_models.ScheduleJobQuery{
		PolicyID: policyID,
	})
	require.Nil(t, err)
	require.Equal(t, 1, len(sjs))
	assert.Equal(t, uuid, sjs[0].JobID)

	policyID, ok := js.jobData.Parameters["policy_id"].(int64)
	require.True(t, ok)
	assert.Equal(t, policyID, policyID)
}

func TestUnschedule(t *testing.T) {
	config.Config = &config.Configuration{}
	dao.ScheduleJob = &fakedScheduleJobDAO{}
	_, err := dao.ScheduleJob.Add(&rep_models.ScheduleJob{
		PolicyID: policyID,
	})
	require.Nil(t, err)
	js := &fakedJobserviceClient{}
	scheduler := NewScheduler(js)
	err = scheduler.Unschedule(policyID)
	require.Nil(t, err)

	sjs, err := dao.ScheduleJob.List(&rep_models.ScheduleJobQuery{
		PolicyID: policyID,
	})
	require.Nil(t, err)
	require.Equal(t, 0, len(sjs))

	assert.True(t, js.stopped)
}
