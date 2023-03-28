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

package handler

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/goharbor/harbor/src/pkg/notification"
	policyModel "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	taskModel "github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	"github.com/goharbor/harbor/src/testing/controller/task"
	"github.com/goharbor/harbor/src/testing/controller/webhook"
	"github.com/goharbor/harbor/src/testing/mock"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
)

type WebhookTestSuite struct {
	htesting.Suite
	webhookCtl *webhook.Controller
	execCtl    *task.ExecutionController
	taskCtl    *task.Controller
}

func (suite *WebhookTestSuite) SetupSuite() {
	suite.webhookCtl = &webhook.Controller{}
	suite.execCtl = &task.ExecutionController{}
	suite.taskCtl = &task.Controller{}
	suite.Config = &restapi.Config{
		WebhookAPI: &webhookAPI{
			webhookCtl: suite.webhookCtl,
			execCtl:    suite.execCtl,
			taskCtl:    suite.taskCtl,
		},
	}

	suite.Suite.SetupSuite()
	notification.Init()
}

func (suite *WebhookTestSuite) TestListWebhookPoliciesOfProject() {
	projectID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("CountPolicies", mock.Anything, mock.Anything).Return(int64(1), nil)
	suite.webhookCtl.On("ListPolicies", mock.Anything, mock.Anything).Return([]*policyModel.Policy{{ID: 1, ProjectID: projectID}}, nil)

	url := fmt.Sprintf("/projects/%d/webhook/policies", projectID)
	var body []*policyModel.Policy
	resp, err := suite.GetJSON(url, &body)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
	suite.Len(body, 1)
	suite.Equal(projectID, body[0].ProjectID)
}

func (suite *WebhookTestSuite) TestCreateWebhookPolicyOfProject() {
	projectID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("CreatePolicy", mock.Anything, mock.Anything).Return(int64(1), nil)
	url := fmt.Sprintf("/projects/%d/webhook/policies", projectID)
	{
		// invalid event type should got 400
		resp, err := suite.PostJSON(url, &models.WebhookPolicy{EventTypes: []string{"INVALID"}})
		suite.NoError(err)
		suite.Equal(400, resp.StatusCode)
	}

	{
		// invalid target type should got 400
		resp, err := suite.PostJSON(url, &models.WebhookPolicy{EventTypes: []string{"PUSH_ARTIFACT"}, Targets: []*models.WebhookTargetObject{{Type: "invalid"}}})
		suite.NoError(err)
		suite.Equal(400, resp.StatusCode)
	}

	{
		// valid policy should got 200
		resp, err := suite.PostJSON(url, &models.WebhookPolicy{EventTypes: []string{"PUSH_ARTIFACT"}, Targets: []*models.WebhookTargetObject{{Type: "http", Address: "http://127.0.0.1"}}})
		suite.NoError(err)
		suite.Equal(201, resp.StatusCode)
	}
}

func (suite *WebhookTestSuite) TestUpdateWebhookPolicyOfProject() {
	projectID := int64(1)
	policyID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("GetPolicy", mock.Anything, mock.Anything).Return(&policyModel.Policy{ID: policyID, ProjectID: projectID}, nil)
	suite.webhookCtl.On("UpdatePolicy", mock.Anything, mock.Anything).Return(nil)
	url := fmt.Sprintf("/projects/%d/webhook/policies/%d", projectID, policyID)
	{
		// invalid event type should got 400
		resp, err := suite.PutJSON(url, &models.WebhookPolicy{EventTypes: []string{"INVALID"}})
		suite.NoError(err)
		suite.Equal(400, resp.StatusCode)
	}

	{
		// invalid target type should got 400
		resp, err := suite.PutJSON(url, &models.WebhookPolicy{EventTypes: []string{"PUSH_ARTIFACT"}, Targets: []*models.WebhookTargetObject{{Type: "invalid"}}})
		suite.NoError(err)
		suite.Equal(400, resp.StatusCode)
	}

	{
		// valid policy should got 200
		resp, err := suite.PutJSON(url, &models.WebhookPolicy{EventTypes: []string{"PUSH_ARTIFACT"}, Targets: []*models.WebhookTargetObject{{Type: "http", Address: "http://127.0.0.1"}}})
		suite.NoError(err)
		suite.Equal(200, resp.StatusCode)
	}
}

func (suite *WebhookTestSuite) TestDeleteWebhookPolicyOfProject() {
	projectID := int64(1)
	policyID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("GetPolicy", mock.Anything, mock.Anything).Return(&policyModel.Policy{ID: policyID, ProjectID: projectID}, nil)
	suite.webhookCtl.On("DeletePolicy", mock.Anything, mock.Anything).Return(nil)
	url := fmt.Sprintf("/projects/%d/webhook/policies/%d", projectID, policyID)
	resp, err := suite.Delete(url)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
}

func (suite *WebhookTestSuite) TestGetWebhookPolicyOfProject() {
	projectID := int64(1)
	policyID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("GetPolicy", mock.Anything, mock.Anything).Return(&policyModel.Policy{ID: policyID, ProjectID: projectID}, nil)
	url := fmt.Sprintf("/projects/%d/webhook/policies/%d", projectID, policyID)
	var body *models.WebhookPolicy
	resp, err := suite.GetJSON(url, &body)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
	suite.Equal(projectID, body.ProjectID)
}

func (suite *WebhookTestSuite) TestListExecutionsOfWebhookPolicy() {
	projectID := int64(1)
	policyID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("GetPolicy", mock.Anything, mock.Anything).Return(&policyModel.Policy{ID: policyID, ProjectID: projectID}, nil)
	suite.webhookCtl.On("CountExecutions", mock.Anything, policyID, mock.Anything).Return(int64(1), nil)
	suite.webhookCtl.On("ListExecutions", mock.Anything, policyID, mock.Anything).Return([]*taskModel.Execution{{ID: 1, VendorID: policyID}}, nil)
	url := fmt.Sprintf("/projects/%d/webhook/policies/%d/executions", projectID, policyID)
	var body []*taskModel.Execution
	resp, err := suite.GetJSON(url, &body)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
	suite.Equal(policyID, body[0].VendorID)
}

func (suite *WebhookTestSuite) TestListTasksOfWebhookExecution() {
	projectID := int64(1)
	policyID := int64(1)
	execID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("GetPolicy", mock.Anything, mock.Anything).Return(&policyModel.Policy{ID: policyID, ProjectID: projectID}, nil)
	suite.execCtl.On("Get", mock.Anything, mock.Anything).Return(&taskModel.Execution{ID: execID, VendorID: projectID, VendorType: "WEBHOOK"}, nil)
	suite.webhookCtl.On("CountTasks", mock.Anything, policyID, mock.Anything).Return(int64(1), nil)
	suite.webhookCtl.On("ListTasks", mock.Anything, policyID, mock.Anything).Return([]*taskModel.Task{{ID: 1, ExecutionID: execID}}, nil)
	url := fmt.Sprintf("/projects/%d/webhook/policies/%d/executions/%d/tasks", projectID, policyID, execID)
	var body []*taskModel.Task
	resp, err := suite.GetJSON(url, &body)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
	suite.Equal(execID, body[0].ExecutionID)
}

func (suite *WebhookTestSuite) TestGetLogsOfWebhookTask() {
	projectID := int64(1)
	policyID := int64(1)
	execID := int64(1)
	taskID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("GetPolicy", mock.Anything, mock.Anything).Return(&policyModel.Policy{ID: policyID, ProjectID: projectID}, nil)
	suite.execCtl.On("Get", mock.Anything, mock.Anything).Return(&taskModel.Execution{ID: execID, VendorID: projectID, VendorType: "WEBHOOK"}, nil)
	suite.taskCtl.On("Get", mock.Anything, mock.Anything).Return(&taskModel.Task{ID: taskID, ExecutionID: execID}, nil)
	suite.webhookCtl.On("GetTaskLog", mock.Anything, taskID).Return([]byte("logs..."), nil)
	url := fmt.Sprintf("/projects/%d/webhook/policies/%d/executions/%d/tasks/%d/log", projectID, policyID, execID, taskID)
	resp, err := suite.Get(url)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()
	suite.NoError(err)
	suite.Equal("logs...", string(data))
}

func (suite *WebhookTestSuite) TestLastTrigger() {
	projectID := int64(1)
	policyID := int64(1)
	now := time.Now()
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.webhookCtl.On("ListPolicies", mock.Anything, mock.Anything).Return([]*policyModel.Policy{{ID: policyID, ProjectID: projectID, EventTypes: []string{"PUSH_ARTIFACT"}}}, nil)
	suite.webhookCtl.On("GetLastTriggerTime", mock.Anything, mock.Anything, policyID).Return(now, nil)
	url := fmt.Sprintf("/projects/%d/webhook/lasttrigger", projectID)
	var body []*models.WebhookLastTrigger
	resp, err := suite.GetJSON(url, &body)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
	suite.Len(body, 1)
	suite.Equal(strfmt.DateTime(now).String(), body[0].LastTriggerTime.String())
}

func (suite *WebhookTestSuite) TestGetSupportedEventTypes() {
	projectID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	url := fmt.Sprintf("/projects/%d/webhook/events", projectID)
	var body *models.SupportedWebhookEventTypes
	resp, err := suite.GetJSON(url, &body)
	suite.NoError(err)
	suite.Equal(200, resp.StatusCode)
	suite.Len(body.EventType, len(notification.GetSupportedEventTypes()))
	suite.Len(body.NotifyType, len(notification.GetSupportedNotifyTypes()))
}

func TestWebhookTestSuite(t *testing.T) {
	suite.Run(t, &WebhookTestSuite{})
}
