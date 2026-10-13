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

package handler

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/rbac"
	rbac_project "github.com/goharbor/harbor/src/common/rbac/project"
	projectmodels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	retentiontesting "github.com/goharbor/harbor/src/testing/controller/retention"
	"github.com/goharbor/harbor/src/testing/mock"
	metadatatesting "github.com/goharbor/harbor/src/testing/pkg/project/metadata"
	htesting "github.com/goharbor/harbor/src/testing/server/v2.0/handler"
)

const (
	retentionProjectID int64 = 1
	retentionPolicyID  int64 = 2
)

type RetentionTestSuite struct {
	htesting.Suite

	projectCtl   *projecttesting.Controller
	proMetaMgr   *metadatatesting.Manager
	retentionCtl *retentiontesting.Controller
}

func (suite *RetentionTestSuite) SetupSuite() {
	suite.projectCtl = &projecttesting.Controller{}
	suite.proMetaMgr = &metadatatesting.Manager{}
	suite.retentionCtl = &retentiontesting.Controller{}
	suite.Config = &restapi.Config{
		RetentionAPI: &retentionAPI{
			projectCtl:   suite.projectCtl,
			proMetaMgr:   suite.proMetaMgr,
			retentionCtl: suite.retentionCtl,
		},
	}
	suite.Suite.SetupSuite()
}

func (suite *RetentionTestSuite) expectPermission(
	action rbac.Action,
	resource rbac.Resource,
	allowed bool,
) {
	suite.Security.On(
		"Can",
		mock.Anything,
		action,
		rbac_project.NewNamespace(retentionProjectID).Resource(resource),
	).Return(allowed).Once()
}

func (suite *RetentionTestSuite) expectAuthenticated(times int) {
	suite.Security.On("IsAuthenticated").Return(true).Times(times)
}

func (suite *RetentionTestSuite) expectPolicyMetadata() {
	suite.proMetaMgr.On(
		"Get",
		mock.Anything,
		retentionProjectID,
		"retention_id",
	).Return(map[string]string{
		"retention_id": strconv.FormatInt(retentionPolicyID, 10),
	}, nil).Once()
}

func retentionModel(cron string) *models.RetentionPolicy {
	return &models.RetentionPolicy{
		Algorithm: "or",
		Rules:     []*models.RetentionRule{},
		Scope: &models.RetentionPolicyScope{
			Level: policy.ScopeLevelProject,
			Ref:   retentionProjectID,
		},
		Trigger: &models.RetentionRuleTrigger{
			Kind: policy.TriggerKindSchedule,
			Settings: map[string]any{
				policy.TriggerSettingsCron: cron,
			},
		},
	}
}

func retentionMetadata(cron string) *policy.Metadata {
	return &policy.Metadata{
		ID: retentionPolicyID,
		Scope: &policy.Scope{
			Level:     policy.ScopeLevelProject,
			Reference: retentionProjectID,
		},
		Trigger: &policy.Trigger{
			Kind: policy.TriggerKindSchedule,
			Settings: map[string]any{
				policy.TriggerSettingsCron: cron,
			},
		},
	}
}

func (suite *RetentionTestSuite) TestTriggerRetentionExecutionAuthorization() {
	endpoint := "/retentions/2/executions"
	policy := retentionMetadata("")

	suite.expectAuthenticated(2)
	suite.expectPermission(rbac.ActionUpdate, rbac.ResourceTagRetention, true)
	suite.expectPermission(rbac.ActionDelete, rbac.ResourceArtifact, false)
	suite.retentionCtl.On("GetRetention", mock.Anything, retentionPolicyID).
		Return(policy, nil).Once()

	res, err := suite.PostJSON(endpoint, map[string]bool{"dry_run": false})
	suite.Require().NoError(err)
	suite.Equal(http.StatusForbidden, res.StatusCode)
	suite.Require().NoError(res.Body.Close())
	suite.retentionCtl.AssertNotCalled(
		suite.T(),
		"TriggerRetentionExec",
		mock.Anything,
		retentionPolicyID,
		mock.Anything,
		false,
	)

	suite.expectAuthenticated(1)
	suite.expectPermission(rbac.ActionUpdate, rbac.ResourceTagRetention, true)
	suite.retentionCtl.On("GetRetention", mock.Anything, retentionPolicyID).
		Return(policy, nil).Once()
	suite.retentionCtl.On(
		"TriggerRetentionExec",
		mock.Anything,
		retentionPolicyID,
		mock.Anything,
		true,
	).Return(int64(10), nil).Once()

	res, err = suite.PostJSON(endpoint, map[string]bool{"dry_run": true})
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, res.StatusCode)
	suite.Require().NoError(res.Body.Close())

	suite.expectAuthenticated(1)
	suite.expectPermission(rbac.ActionUpdate, rbac.ResourceTagRetention, true)
	suite.expectPermission(rbac.ActionDelete, rbac.ResourceArtifact, true)
	suite.retentionCtl.On("GetRetention", mock.Anything, retentionPolicyID).
		Return(policy, nil).Once()
	suite.retentionCtl.On(
		"TriggerRetentionExec",
		mock.Anything,
		retentionPolicyID,
		mock.Anything,
		false,
	).Return(int64(11), nil).Once()

	res, err = suite.PostJSON(endpoint, map[string]bool{"dry_run": false})
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, res.StatusCode)
	suite.Require().NoError(res.Body.Close())
}

func (suite *RetentionTestSuite) TestCreateScheduledRetentionAuthorization() {
	endpoint := "/retentions"

	suite.expectAuthenticated(2)
	suite.expectPermission(rbac.ActionCreate, rbac.ResourceTagRetention, true)
	suite.expectPermission(rbac.ActionDelete, rbac.ResourceArtifact, false)
	res, err := suite.PostJSON(endpoint, retentionModel("0 0 0 * * *"))
	suite.Require().NoError(err)
	suite.Equal(http.StatusForbidden, res.StatusCode)
	suite.Require().NoError(res.Body.Close())
	suite.retentionCtl.AssertNotCalled(
		suite.T(),
		"CreateRetention",
		mock.Anything,
		mock.Anything,
	)

	suite.expectAuthenticated(1)
	suite.expectPermission(rbac.ActionCreate, rbac.ResourceTagRetention, true)
	suite.expectPermission(rbac.ActionDelete, rbac.ResourceArtifact, true)
	suite.expectCreateRetention()
	res, err = suite.PostJSON(endpoint, retentionModel("0 0 0 * * *"))
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, res.StatusCode)
	suite.Require().NoError(res.Body.Close())

	suite.expectAuthenticated(1)
	suite.expectPermission(rbac.ActionCreate, rbac.ResourceTagRetention, true)
	suite.expectCreateRetention()
	res, err = suite.PostJSON(endpoint, retentionModel(""))
	suite.Require().NoError(err)
	suite.Equal(http.StatusCreated, res.StatusCode)
	suite.Require().NoError(res.Body.Close())
}

func (suite *RetentionTestSuite) expectCreateRetention() {
	suite.projectCtl.On("Get", mock.Anything, retentionProjectID).
		Return(&projectmodels.Project{ProjectID: retentionProjectID}, nil).Once()
	suite.proMetaMgr.On(
		"Get",
		mock.Anything,
		retentionProjectID,
		"retention_id",
	).Return(map[string]string{}, nil).Once()
	suite.retentionCtl.On("CreateRetention", mock.Anything, mock.Anything).
		Return(retentionPolicyID, nil).Once()
	suite.proMetaMgr.On(
		"Add",
		mock.Anything,
		retentionProjectID,
		map[string]string{
			"retention_id": strconv.FormatInt(retentionPolicyID, 10),
		},
	).Return(nil).Once()
}

func (suite *RetentionTestSuite) TestUpdateScheduledRetentionAuthorization() {
	endpoint := "/retentions/2"

	suite.expectAuthenticated(2)
	suite.expectPermission(rbac.ActionUpdate, rbac.ResourceTagRetention, true)
	suite.expectPolicyMetadata()
	suite.retentionCtl.On("GetRetention", mock.Anything, retentionPolicyID).
		Return(retentionMetadata("0 0 0 * * *"), nil).Once()
	suite.expectPermission(rbac.ActionDelete, rbac.ResourceArtifact, false)
	res, err := suite.PutJSON(endpoint, retentionModel(""))
	suite.Require().NoError(err)
	suite.Equal(http.StatusForbidden, res.StatusCode)
	suite.Require().NoError(res.Body.Close())
	suite.retentionCtl.AssertNotCalled(
		suite.T(),
		"UpdateRetention",
		mock.Anything,
		mock.Anything,
	)

	suite.expectAuthenticated(2)
	suite.expectPermission(rbac.ActionUpdate, rbac.ResourceTagRetention, true)
	suite.expectPolicyMetadata()
	suite.retentionCtl.On("GetRetention", mock.Anything, retentionPolicyID).
		Return(retentionMetadata(""), nil).Once()
	suite.expectPermission(rbac.ActionDelete, rbac.ResourceArtifact, false)
	res, err = suite.PutJSON(endpoint, retentionModel("0 0 0 * * *"))
	suite.Require().NoError(err)
	suite.Equal(http.StatusForbidden, res.StatusCode)
	suite.Require().NoError(res.Body.Close())

	suite.expectAuthenticated(1)
	suite.expectPermission(rbac.ActionUpdate, rbac.ResourceTagRetention, true)
	suite.expectPolicyMetadata()
	suite.retentionCtl.On("GetRetention", mock.Anything, retentionPolicyID).
		Return(retentionMetadata("0 0 0 * * *"), nil).Once()
	suite.expectPermission(rbac.ActionDelete, rbac.ResourceArtifact, true)
	suite.retentionCtl.On("UpdateRetention", mock.Anything, mock.Anything).
		Return(nil).Once()
	res, err = suite.PutJSON(endpoint, retentionModel("0 0 0 * * *"))
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Require().NoError(res.Body.Close())

	suite.expectAuthenticated(1)
	suite.expectPermission(rbac.ActionUpdate, rbac.ResourceTagRetention, true)
	suite.expectPolicyMetadata()
	suite.retentionCtl.On("GetRetention", mock.Anything, retentionPolicyID).
		Return(retentionMetadata(""), nil).Once()
	suite.retentionCtl.On("UpdateRetention", mock.Anything, mock.Anything).
		Return(nil).Once()
	res, err = suite.PutJSON(endpoint, retentionModel(""))
	suite.Require().NoError(err)
	suite.Equal(http.StatusOK, res.StatusCode)
	suite.Require().NoError(res.Body.Close())
}

func TestRetentionTestSuite(t *testing.T) {
	suite.Run(t, &RetentionTestSuite{})
}
