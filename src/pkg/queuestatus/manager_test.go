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

package queuestatus

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/pkg/queuestatus/model"
	htesting "github.com/goharbor/harbor/src/testing"
)

type ManagerTestSuite struct {
	htesting.Suite
	mgr Manager
}

func (s *ManagerTestSuite) SetupSuite() {
	s.Suite.SetupSuite()
	s.Suite.ClearTables = []string{"queue_status"}
	s.mgr = newManager()
}

func (s *ManagerTestSuite) TestAllJobTypeStatus() {
	ctx := s.Context()
	_, err := s.mgr.CreateOrUpdate(ctx, &model.JobQueueStatus{JobType: "GARBAGE_COLLECTION", Paused: true})
	s.Nil(err)
	_, err = s.mgr.CreateOrUpdate(ctx, &model.JobQueueStatus{JobType: "REPLICATION", Paused: false})
	s.Nil(err)
	resultMap, err := s.mgr.AllJobTypeStatus(ctx)
	s.Nil(err)
	s.Equal(2, len(resultMap))
	s.True(resultMap["GARBAGE_COLLECTION"])
	s.False(resultMap["REPLICATION"])
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, &ManagerTestSuite{})
}
