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
	"testing"

	model "github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
)

type executionControllerTestSuite struct {
	suite.Suite
	ctl *executionController
	mgr *task.FakeExecutionManager
}

// TestExecutionControllerTestSuite tests controller.
func TestExecutionControllerTestSuite(t *testing.T) {
	suite.Run(t, &executionControllerTestSuite{})
}

// SetupTest setups the testing env.
func (ec *executionControllerTestSuite) SetupTest() {
	ec.mgr = &task.FakeExecutionManager{}
	ec.ctl = &executionController{
		mgr: ec.mgr,
	}
}

// TestCreate tests create.
func (ec *executionControllerTestSuite) TestCreate() {
	ec.mgr.On("Create", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := ec.ctl.Create(nil, "", 1, "")
	ec.NoError(err)
	ec.Equal(int64(1), id)
}

// TestMarkDown tests mark down.
func (ec *executionControllerTestSuite) TestMarkDone() {
	ec.mgr.On("MarkDone", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := ec.ctl.MarkDone(nil, 1, "")
	ec.NoError(err)
}

// TestMarkError tests mark error.
func (ec *executionControllerTestSuite) TestMarkError() {
	ec.mgr.On("MarkError", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	err := ec.ctl.MarkError(nil, 1, "")
	ec.NoError(err)
}

// TestStop tests stop.
func (ec *executionControllerTestSuite) TestStop() {
	ec.mgr.On("Stop", mock.Anything, mock.Anything).Return(nil)
	err := ec.ctl.Stop(nil, 1)
	ec.NoError(err)
}

// TestDelete tests delete.
func (ec *executionControllerTestSuite) TestDelete() {
	ec.mgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err := ec.ctl.Delete(nil, 1)
	ec.NoError(err)
}

// TestGet tests get.
func (ec *executionControllerTestSuite) TestGet() {
	ec.mgr.On("Get", mock.Anything, mock.Anything).Return(&model.Execution{ID: 1}, nil)
	e, err := ec.ctl.Get(nil, 1)
	ec.NoError(err)
	ec.Equal(int64(1), e.ID)
}

// TestList tests list.
func (ec *executionControllerTestSuite) TestList() {
	ec.mgr.On("List", mock.Anything, mock.Anything).Return([]*model.Execution{
		{ID: 1},
		{ID: 2},
	}, nil)
	es, err := ec.ctl.List(nil, nil)
	ec.NoError(err)
	ec.Len(es, 2)
}
