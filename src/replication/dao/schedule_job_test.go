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

	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sjID int64

func TestAddScheduleJob(t *testing.T) {
	sj := &models.ScheduleJob{
		PolicyID: 1,
		JobID:    "uuid",
		Status:   "running",
	}
	id, err := ScheduleJob.Add(sj)
	require.Nil(t, err)
	sjID = id
}

func TestUpdateScheduleJob(t *testing.T) {
	err := ScheduleJob.Update(&models.ScheduleJob{
		ID:     sjID,
		Status: "success",
	}, "Status")
	require.Nil(t, err)
}

func TestGetScheduleJob(t *testing.T) {
	sj, err := ScheduleJob.Get(sjID)
	require.Nil(t, err)
	assert.Equal(t, int64(1), sj.PolicyID)
	assert.Equal(t, "success", sj.Status)
}

func TestListScheduleJobs(t *testing.T) {
	// nil query
	sjs, err := ScheduleJob.List()
	require.Nil(t, err)
	assert.Equal(t, 1, len(sjs))

	// query
	sjs, err = ScheduleJob.List(&models.ScheduleJobQuery{
		PolicyID: 1,
	})
	require.Nil(t, err)
	assert.Equal(t, 1, len(sjs))

	// query
	sjs, err = ScheduleJob.List(&models.ScheduleJobQuery{
		PolicyID: 2,
	})
	require.Nil(t, err)
	assert.Equal(t, 0, len(sjs))
}

func TestDeleteScheduleJob(t *testing.T) {
	err := ScheduleJob.Delete(sjID)
	require.Nil(t, err)

	sj, err := ScheduleJob.Get(sjID)
	require.Nil(t, err)
	assert.Nil(t, sj)
}
