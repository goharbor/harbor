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

package purge

import (
	"github.com/goharbor/harbor/src/pkg/task"
	testingTask "github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
)

type PurgeControllerTestSuite struct {
	suite.Suite
	taskMgr *testingTask.Manager
	exeMgr  *testingTask.ExecutionManager
	Ctl     Controller
}

func (p *PurgeControllerTestSuite) SetupSuite() {
	p.taskMgr = &testingTask.Manager{}
	p.exeMgr = &testingTask.ExecutionManager{}
	p.Ctl = &controller{
		taskMgr: p.taskMgr,
		exeMgr:  p.exeMgr,
	}
}

func (p *PurgeControllerTestSuite) TestStart() {
	p.exeMgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	p.taskMgr.On("Create", mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	policy := JobPolicy{}
	id, err := p.Ctl.Start(nil, policy, task.ExecutionTriggerManual)
	p.Nil(err)
	p.Equal(int64(1), id)
}

func (p *PurgeControllerTestSuite) TearDownSuite() {
}

func TestPurgeControllerTestSuite(t *testing.T) {
	suite.Run(t, &PurgeControllerTestSuite{})
}
