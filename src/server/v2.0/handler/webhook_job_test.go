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
	"testing"

	policyModel "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	projectModel "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/task"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	"github.com/goharbor/harbor/src/testing/controller/webhook"
	"github.com/goharbor/harbor/src/testing/mock"
	"github.com/goharbor/harbor/src/testing/pkg/project"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
	"github.com/stretchr/testify/suite"
)

type WebhookJobTestSuite struct {
	htesting.Suite
	webhookCtl *webhook.Controller
	projectMgr *project.Manager
}

func (suite *WebhookJobTestSuite) SetupSuite() {
	suite.webhookCtl = &webhook.Controller{}
	suite.projectMgr = &project.Manager{}
	suite.Config = &restapi.Config{
		WebhookjobAPI: &webhookJobAPI{
			webhookCtl: suite.webhookCtl,
			projectMgr: suite.projectMgr,
		},
	}
	suite.Suite.SetupSuite()
}

func (suite *WebhookJobTestSuite) TestListWebhookJobs() {
	projectID := int64(1)
	policyID := int64(1)
	suite.Security.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	suite.projectMgr.On("Get", mock.Anything, projectID).Return(&projectModel.Project{ProjectID: projectID}, nil)
	suite.webhookCtl.On("GetPolicy", mock.Anything, policyID).Return(&policyModel.Policy{ID: policyID, Name: "test-policy"}, nil).Once()
	suite.webhookCtl.On("CountExecutions", mock.Anything, policyID, mock.Anything).Return(int64(2), nil)
	t1 := &task.Execution{ID: 1, VendorType: "WEBHOOK", VendorID: policyID, Status: "Success"}
	t2 := &task.Execution{ID: 2, VendorType: "SLACK", VendorID: policyID, Status: "Stopped"}
	t3 := &task.Execution{ID: 2, VendorType: "TEAMS", VendorID: policyID, Status: "Stopped"}
	suite.webhookCtl.On("ListExecutions", mock.Anything, policyID, mock.Anything).Return([]*task.Execution{t1, t2, t3}, nil)

	{
		// query has no policy id should got 422
		url := fmt.Sprintf("/projects/%d/webhook/jobs", projectID)
		var body []*models.WebhookJob
		resp, err := suite.GetJSON(url, &body)
		suite.NoError(err)
		suite.Equal(422, resp.StatusCode)
	}

	{
		// unmatched project id should got 404
		url := fmt.Sprintf("/projects/%d/webhook/jobs?policy_id=%d", projectID, policyID)
		var body []*models.WebhookJob
		resp, err := suite.GetJSON(url, &body)
		suite.NoError(err)
		suite.Equal(404, resp.StatusCode)
	}

	{
		// normal requests should got 200
		suite.webhookCtl.On("GetPolicy", mock.Anything, policyID, mock.Anything).Return(&policyModel.Policy{ID: policyID, Name: "test-policy", ProjectID: projectID}, nil)
		url := fmt.Sprintf("/projects/%d/webhook/jobs?policy_id=%d", projectID, policyID)
		var body []*models.WebhookJob
		resp, err := suite.GetJSON(url, &body)
		suite.NoError(err)
		suite.Equal(200, resp.StatusCode)
		suite.Len(body, 3)
		// verify backward compatible
		suite.Equal(body[0].ID, int64(1))
		suite.Equal(body[0].NotifyType, "http")
		suite.Equal(body[0].Status, "Success")
		suite.Equal(body[1].ID, int64(2))
		suite.Equal(body[1].NotifyType, "slack")
		suite.Equal(body[1].Status, "Stopped")
		suite.Equal(body[2].ID, int64(3))
		suite.Equal(body[2].NotifyType, "teams")
		suite.Equal(body[2].Status, "Stopped")
	}

}

func TestWebhookJobTestSuite(t *testing.T) {
	suite.Run(t, &WebhookJobTestSuite{})
}
