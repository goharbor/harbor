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
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/lib/errors"
	importmodel "github.com/goharbor/harbor/src/pkg/modelimport/model"
	promodels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/task"
	modelimporttesting "github.com/goharbor/harbor/src/testing/controller/modelimport"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
)

type ModelImportTestSuite struct {
	suite.Suite
	ctl  *modelimporttesting.Controller
	proj *projecttesting.Controller
	api  *modelImportAPI
}

func (suite *ModelImportTestSuite) SetupTest() {
	suite.ctl = &modelimporttesting.Controller{}
	suite.proj = &projecttesting.Controller{}
	suite.api = &modelImportAPI{
		modelImportCtl: suite.ctl,
		projectCtl:     suite.proj,
	}
	mock.OnAnything(suite.proj, "GetByName").Return(&promodels.Project{ProjectID: 1, Name: "proj"}, nil)
	mock.OnAnything(suite.ctl, "GetPolicyByName").Return(&importmodel.Policy{ID: 5, ProjectID: 1, Name: "pol"}, nil)
}

// A caller-supplied execution ID that belongs to a different policy (or project)
// must not resolve, otherwise executions/tasks/logs leak across policies.
func (suite *ModelImportTestSuite) TestAuthorizeExecutionRejectsForeignExecution() {
	mock.OnAnything(suite.ctl, "GetExecution").Return(&task.Execution{ID: 42, VendorID: 99}, nil)

	_, _, err := suite.api.authorizeExecution(context.Background(), "proj", "pol", 42)
	suite.Require().Error(err)
	suite.True(errors.IsNotFoundErr(err), "expected a not-found error, got %v", err)
}

func (suite *ModelImportTestSuite) TestAuthorizeExecutionAllowsOwnedExecution() {
	mock.OnAnything(suite.ctl, "GetExecution").Return(&task.Execution{ID: 42, VendorID: 5}, nil)

	policy, exec, err := suite.api.authorizeExecution(context.Background(), "proj", "pol", 42)
	suite.Require().NoError(err)
	suite.Equal(int64(5), policy.ID)
	suite.Equal(int64(42), exec.ID)
}

func TestModelImportTestSuite(t *testing.T) {
	suite.Run(t, new(ModelImportTestSuite))
}
