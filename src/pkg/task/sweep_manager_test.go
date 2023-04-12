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

package task

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/task/dao"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/goharbor/harbor/src/testing/mock"
)

type sweepManagerTestSuite struct {
	htesting.Suite
	execDao *mockExecutionDAO
	mgr     *sweepManager
}

func TestSweepManager(t *testing.T) {
	suite.Run(t, &sweepManagerTestSuite{})
}

func (suite *sweepManagerTestSuite) SetupSuite() {
	suite.Suite.SetupSuite()
	suite.execDao = &mockExecutionDAO{}
	suite.mgr = &sweepManager{execDAO: suite.execDao}
}

func (suite *sweepManagerTestSuite) TestGetCandidateMaxStartTime() {
	// test error case
	suite.execDao.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("failed to list executions")).Once()
	startTime, err := suite.mgr.getCandidateMaxStartTime(context.TODO(), "WEBHOOK", 1, 10)
	suite.Error(err, "should got error")
	suite.Nil(startTime)
	// test normal case
	now := time.Now()
	execs := []*dao.Execution{{ID: 1, StartTime: now}}
	suite.execDao.On("List", mock.Anything, mock.Anything).Return(execs, nil)
	startTime, err = suite.mgr.getCandidateMaxStartTime(context.TODO(), "WEBHOOK", 1, 10)
	suite.NoError(err, "should not got error")
	suite.Equal(now.String(), startTime.String())
}

func (suite *sweepManagerTestSuite) Test_sweepManager_FixDanglingStateExecution() {
	err := suite.mgr.FixDanglingStateExecution(suite.Context())
	suite.NoError(err, "should not got error")
}
