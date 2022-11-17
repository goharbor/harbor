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

package dao

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/queuestatus/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type DaoTestSuite struct {
	htesting.Suite
	dao DAO
}

func (s *DaoTestSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.Suite.ClearTables = []string{"queue_status"}
	s.dao = New()
}

func (s *DaoTestSuite) TestCRUDQueueStatus() {
	ctx := s.Context()
	jobType := "GARBAGE_COLLECTION"
	queueStatus := model.JobQueueStatus{
		JobType: jobType,
	}
	id, err := s.dao.InsertOrUpdate(ctx, &queueStatus)
	s.Nil(err)
	s.True(id > 0)

	id2, err := s.dao.InsertOrUpdate(ctx, &queueStatus)
	s.Nil(err)
	s.Equal(id, id2)

	qs, err2 := s.dao.GetByJobType(ctx, jobType)
	s.Nil(err2)
	s.Equal("GARBAGE_COLLECTION", qs.JobType)
	s.False(qs.Paused)

	err3 := s.dao.UpdateStatus(ctx, jobType, true)
	s.Nil(err3)

	qs2, err4 := s.dao.GetByJobType(ctx, jobType)
	s.Nil(err4)
	s.Equal(jobType, qs2.JobType)
	s.True(qs2.Paused)

	qList, err := s.dao.Query(ctx, q.New(q.KeyWords{"job_type": jobType}))
	s.Nil(err)
	s.Equal(1, len(qList))

}

func TestDaoTestSuite(t *testing.T) {
	suite.Run(t, &DaoTestSuite{})
}
