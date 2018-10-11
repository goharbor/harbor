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

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddAdminJob(t *testing.T) {
	job := &models.AdminJob{
		Name: "job",
		Kind: "jobKind",
	}

	job0 := &models.AdminJob{
		Name: "GC",
		Kind: "testKind",
	}

	// add
	id, err := AddAdminJob(job0)
	require.Nil(t, err)
	job0.ID = id

	// get
	job1, err := GetAdminJob(id)
	require.Nil(t, err)
	assert.Equal(t, job1.ID, job0.ID)
	assert.Equal(t, job1.Name, job0.Name)

	// update status
	err = UpdateAdminJobStatus(id, "testStatus")
	require.Nil(t, err)
	job2, err := GetAdminJob(id)
	assert.Equal(t, job2.Status, "testStatus")

	// set uuid
	err = SetAdminJobUUID(id, "f5ef34f4cb3588d663176132")
	require.Nil(t, err)
	job3, err := GetAdminJob(id)
	require.Nil(t, err)
	assert.Equal(t, job3.UUID, "f5ef34f4cb3588d663176132")

	// get admin jobs
	_, err = AddAdminJob(job)
	require.Nil(t, err)
	query := &models.AdminJobQuery{
		Name: "job",
	}
	jobs, err := GetAdminJobs(query)
	assert.Equal(t, len(jobs), 1)

	// get top 10
	_, err = AddAdminJob(job)
	require.Nil(t, err)

	jobs, _ = GetTop10AdminJobsOfName("job")
	assert.Equal(t, len(jobs), 2)
}
