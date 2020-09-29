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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// AdminJobSuite is a test suite for testing admin job
type AdminJobSuite struct {
	suite.Suite

	job0 *models.AdminJob
	ids  []int64
}

// TestAdminJob is the entry point of AdminJobSuite
func TestAdminJob(t *testing.T) {
	suite.Run(t, &AdminJobSuite{})
}

// SetupSuite prepares testing env for the suite
func (suite *AdminJobSuite) SetupSuite() {
	job := &models.AdminJob{
		Name: "job",
		Kind: "jobKind",
	}

	job0 := &models.AdminJob{
		Name:       "GC",
		Kind:       "testKind",
		Parameters: "{test:test}",
	}

	suite.ids = make([]int64, 0)

	// add
	id, err := AddAdminJob(job0)
	require.NoError(suite.T(), err)
	job0.ID = id
	suite.job0 = job0
	suite.ids = append(suite.ids, id)

	id1, err := AddAdminJob(job)
	require.NoError(suite.T(), err)
	suite.ids = append(suite.ids, id1)
}

// TearDownSuite cleans testing env
func (suite *AdminJobSuite) TearDownSuite() {
	for _, id := range suite.ids {
		err := DeleteAdminJob(id)
		suite.NoError(err, fmt.Sprintf("clear admin job: %d", id))
	}
}

// TestAdminJobBase ...
func (suite *AdminJobSuite) TestAdminJobBase() {
	// get
	job1, err := GetAdminJob(suite.job0.ID)
	require.Nil(suite.T(), err)
	suite.Equal(job1.ID, suite.job0.ID)
	suite.Equal(job1.Name, suite.job0.Name)
	suite.Equal(job1.Parameters, suite.job0.Parameters)

	// set uuid
	err = SetAdminJobUUID(suite.job0.ID, "f5ef34f4cb3588d663176132")
	require.Nil(suite.T(), err)
	job3, err := GetAdminJob(suite.job0.ID)
	require.Nil(suite.T(), err)
	suite.Equal(job3.UUID, "f5ef34f4cb3588d663176132")

	// get admin jobs
	query := &models.AdminJobQuery{
		Name: "job",
	}
	jobs, err := GetAdminJobs(query)
	suite.Equal(len(jobs), 1)

	// get top 10
	jobs, _ = GetTop10AdminJobsOfName("job")
	suite.Equal(len(jobs), 1)
}

// TestAdminJobUpdateStatus ...
func (suite *AdminJobSuite) TestAdminJobUpdateStatus() {
	// update status
	err := UpdateAdminJobStatus(suite.job0.ID, "testStatus", 1, 10000)
	require.Nil(suite.T(), err)

	job2, err := GetAdminJob(suite.job0.ID)
	require.Nil(suite.T(), err)
	suite.Equal(job2.Status, "testStatus")

	// Update status with same rev
	err = UpdateAdminJobStatus(suite.job0.ID, "testStatus3", 3, 10000)
	require.Nil(suite.T(), err)

	job3, err := GetAdminJob(suite.job0.ID)
	require.Nil(suite.T(), err)
	suite.Equal(job3.Status, "testStatus3")

	// Update status with same rev, previous status
	err = UpdateAdminJobStatus(suite.job0.ID, "testStatus2", 2, 10000)
	require.Nil(suite.T(), err)

	job4, err := GetAdminJob(suite.job0.ID)
	require.Nil(suite.T(), err)
	// No status change
	suite.Equal(job4.Status, "testStatus3")

	// Update status with previous rev
	err = UpdateAdminJobStatus(suite.job0.ID, "testStatus4", 4, 9999)
	require.Nil(suite.T(), err)

	job5, err := GetAdminJob(suite.job0.ID)
	require.Nil(suite.T(), err)
	// No status change
	suite.Equal(job5.Status, "testStatus3")

	// Update status with latest rev
	err = UpdateAdminJobStatus(suite.job0.ID, "testStatus", 1, 10001)
	require.Nil(suite.T(), err)

	job6, err := GetAdminJob(suite.job0.ID)
	require.Nil(suite.T(), err)
	suite.Equal(job6.Status, "testStatus")
}
