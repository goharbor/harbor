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

package webhook

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/policy/model"
	task_model "github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/notification/policy"
	"github.com/goharbor/harbor/src/testing/pkg/task"
	"github.com/stretchr/testify/suite"
)

type controllerTestSuite struct {
	suite.Suite
	ctl       *controller
	policyMgr *policy.Manager
	taskMgr   *task.Manager
	execMgr   *task.ExecutionManager
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &controllerTestSuite{})
}

func (c *controllerTestSuite) SetupTest() {
	c.policyMgr = &policy.Manager{}
	c.taskMgr = &task.Manager{}
	c.execMgr = &task.ExecutionManager{}
	c.ctl = &controller{
		policyMgr: c.policyMgr,
		taskMgr:   c.taskMgr,
		execMgr:   c.execMgr,
	}
}

func (c *controllerTestSuite) TestCreatePolicy() {
	c.policyMgr.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)
	id, err := c.ctl.CreatePolicy(context.TODO(), &model.Policy{Name: "test-policy"})
	c.NoError(err)
	c.Equal(int64(1), id)
}

func (c *controllerTestSuite) TestListPolicies() {
	c.policyMgr.On("List", mock.Anything, mock.Anything).Return([]*model.Policy{{Name: "test-policy-1"}, {Name: "test-policy-2"}}, nil)
	policies, err := c.ctl.ListPolicies(context.TODO(), q.MustClone(nil))
	c.NoError(err)
	c.Len(policies, 2)
}

func (c *controllerTestSuite) TestCountPolicies() {
	c.policyMgr.On("Count", mock.Anything, mock.Anything).Return(int64(3), nil)
	cnt, err := c.ctl.CountPolicies(context.TODO(), q.MustClone(nil))
	c.NoError(err)
	c.Equal(int64(3), cnt)
}

func (c *controllerTestSuite) TestGetPolicy() {
	c.policyMgr.On("Get", mock.Anything, mock.Anything).Return(&model.Policy{Name: "test-policy"}, nil)
	p, err := c.ctl.GetPolicy(context.TODO(), 1)
	c.NoError(err)
	c.Equal("test-policy", p.Name)
}

func (c *controllerTestSuite) TestUpdatePolicy() {
	c.policyMgr.On("Update", mock.Anything, mock.Anything).Return(nil)
	err := c.ctl.UpdatePolicy(context.TODO(), &model.Policy{})
	c.NoError(err)
}

func (c *controllerTestSuite) TestDeletePolicy() {
	delExecErr := errors.New("delete executions error")
	// failed to delete policy due to webhook executions deletion error
	c.execMgr.On("DeleteByVendor", mock.Anything, "WEBHOOK", mock.Anything).Return(delExecErr).Once()
	err := c.ctl.DeletePolicy(context.TODO(), 1)
	c.ErrorIs(err, delExecErr)

	// failed to delete policy due to slack executions deletion error
	c.execMgr.On("DeleteByVendor", mock.Anything, "WEBHOOK", mock.Anything).Return(nil).Once()
	c.execMgr.On("DeleteByVendor", mock.Anything, "SLACK", mock.Anything).Return(delExecErr).Once()
	err = c.ctl.DeletePolicy(context.TODO(), 1)
	c.ErrorIs(err, delExecErr)

	// failed to delete policy due to teams executions deletion error
	c.execMgr.On("DeleteByVendor", mock.Anything, "WEBHOOK", mock.Anything).Return(nil).Once()
	c.execMgr.On("DeleteByVendor", mock.Anything, "TEAMS", mock.Anything).Return(delExecErr).Once()
	err = c.ctl.DeletePolicy(context.TODO(), 1)
	c.ErrorIs(err, delExecErr)

	// successfully deletion for all
	c.execMgr.On("DeleteByVendor", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	c.policyMgr.On("Delete", mock.Anything, mock.Anything).Return(nil)
	err = c.ctl.DeletePolicy(context.TODO(), 1)
	c.NoError(err)
}

func (c *controllerTestSuite) TestGetRelatedPolices() {
	c.policyMgr.On("GetRelatedPolices", mock.Anything, mock.Anything, mock.Anything).Return([]*model.Policy{{Name: "test-policy-1"}, {Name: "test-policy-2"}}, nil)
	policies, err := c.ctl.GetRelatedPolices(context.TODO(), 1, "mock")
	c.NoError(err)
	c.Len(policies, 2)
}

func (c *controllerTestSuite) TestCountExecutions() {
	c.execMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	cnt, err := c.ctl.CountExecutions(context.TODO(), 1, q.MustClone(nil))
	c.NoError(err)
	c.Equal(int64(1), cnt)
}

func (c *controllerTestSuite) TestListExecutions() {
	c.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task_model.Execution{{ID: 1, VendorType: "WEBHOOK", VendorID: 1}}, nil)
	execs, err := c.ctl.ListExecutions(context.TODO(), 1, q.MustClone(nil))
	c.NoError(err)
	c.Len(execs, 1)
}

func (c *controllerTestSuite) TestCountTasks() {
	c.taskMgr.On("Count", mock.Anything, mock.Anything).Return(int64(1), nil)
	cnt, err := c.ctl.CountTasks(context.TODO(), 1, q.MustClone(nil))
	c.NoError(err)
	c.Equal(int64(1), cnt)
}

func (c *controllerTestSuite) TestListTasks() {
	c.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task_model.Task{{ID: 1, ExecutionID: 1}}, nil)
	tasks, err := c.ctl.ListTasks(context.TODO(), 1, q.MustClone(nil))
	c.NoError(err)
	c.Len(tasks, 1)
}

func (c *controllerTestSuite) TestGetTask() {
	c.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task_model.Task{{ID: 1, ExecutionID: 1}}, nil)
	task, err := c.ctl.GetTask(context.TODO(), 1)
	c.NoError(err)
	c.Equal(int64(1), task.ID)
	c.Equal(int64(1), task.ExecutionID)
}

func (c *controllerTestSuite) TestGetTaskLog() {
	c.taskMgr.On("List", mock.Anything, mock.Anything).Return([]*task_model.Task{{ID: 1, ExecutionID: 1}}, nil)
	c.taskMgr.On("GetLog", mock.Anything, mock.Anything).Return([]byte("test logs"), nil)
	logs, err := c.ctl.GetTaskLog(context.TODO(), 1)
	c.NoError(err)
	c.Equal("test logs", string(logs))
}

func (c *controllerTestSuite) TestGetLastTriggerTime() {
	now := time.Now()
	c.execMgr.On("List", mock.Anything, mock.Anything).Return([]*task_model.Execution{{ID: 1, StartTime: now}}, nil)
	time, err := c.ctl.GetLastTriggerTime(context.TODO(), "mock", 1)
	c.NoError(err)
	c.Equal(now, time)
}
