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

package hook

import (
	"testing"

	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/task/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type handlerTestingSuite struct {
	suite.Suite
	t               *testing.T
	assert          *assert.Assertions
	require         *require.Assertions
	mockTaskManager *test.MockTaskManager
}

func (h *handlerTestingSuite) SetupSuite() {
	h.t = h.T()
	h.assert = assert.New(h.t)
	h.require = require.New(h.t)
	h.mockTaskManager = &test.MockTaskManager{}
	Hdlr = &handler{
		mgr: h.mockTaskManager,
	}
}

// handle check in data
func (h *handlerTestingSuite) TestHandleHookWithCheckInData() {
	h.mockTaskManager.On("AppendCheckInData", mock.Anything).Return(nil)
	err := Hdlr.Handle(1, &job.StatusChange{
		CheckIn: "check_in",
	})
	h.mockTaskManager.AssertCalled(h.t, "AppendCheckInData", mock.Anything)
	h.require.Nil(err)
}

// handle status updating data
func (h *handlerTestingSuite) TestHandleHookWithStatusUpdateData() {
	h.mockTaskManager.On("UpdateStatus", mock.Anything).Return(nil)
	err := Hdlr.Handle(1, &job.StatusChange{
		Status: job.RunningStatus.String(),
	})
	h.mockTaskManager.AssertCalled(h.t, "UpdateStatus", mock.Anything)
	h.require.Nil(err)
}

func TestControllerTestingSuite(t *testing.T) {
	suite.Run(t, &handlerTestingSuite{})
}
