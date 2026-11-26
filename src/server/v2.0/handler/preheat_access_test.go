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
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/rbac"
	rbacproject "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	securityrobot "github.com/goharbor/harbor/src/common/security/robot"
	preheatcontroller "github.com/goharbor/harbor/src/controller/p2p/preheat"
	projectcontroller "github.com/goharbor/harbor/src/controller/project"
	robotcontroller "github.com/goharbor/harbor/src/controller/robot"
	"github.com/goharbor/harbor/src/pkg/p2p/preheat/models/policy"
	"github.com/goharbor/harbor/src/pkg/permission/types"
	projectmodel "github.com/goharbor/harbor/src/pkg/project/models"
	robotmodel "github.com/goharbor/harbor/src/pkg/robot/model"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/preheat"
	testpreheat "github.com/goharbor/harbor/src/testing/controller/p2p/preheat"
	testproject "github.com/goharbor/harbor/src/testing/controller/project"
)

type manualPreheatPolicyController struct {
	preheatcontroller.Controller
}

func (*manualPreheatPolicyController) GetPolicyByName(_ context.Context, projectID int64, name string) (*policy.Schema, error) {
	return &policy.Schema{ID: 2, ProjectID: projectID, Name: name}, nil
}

func TestManualPreheatRequiresUpdatePermission(t *testing.T) {
	for _, tc := range []struct {
		name        string
		allowUpdate bool
		code        int
	}{
		{"read-only-robot", false, http.StatusForbidden},
		{"robot-with-update-permission", true, http.StatusCreated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			projectCtl := &testproject.Controller{}
			project := &projectmodel.Project{ProjectID: 1, Name: "example"}
			projectCtl.On("GetByName", mock.Anything, "example").Return(project, nil)
			projectCtl.On("Get", mock.Anything, int64(1), mock.Anything).Return(project, nil)
			previous := baseProjectCtl
			previousController := projectcontroller.Ctl
			baseProjectCtl = projectCtl
			projectcontroller.Ctl = projectCtl
			defer func() {
				baseProjectCtl = previous
				projectcontroller.Ctl = previousController
			}()
			access := []*types.Policy{{Resource: rbac.ResourcePreatPolicy, Action: rbac.ActionRead}}
			if tc.allowUpdate {
				access = append(access, &types.Policy{Resource: rbac.ResourcePreatPolicy, Action: rbac.ActionUpdate})
			}
			secCtx := securityrobot.NewSecurityContext(&robotcontroller.Robot{
				Robot:       robotmodel.Robot{Name: "robot$example", ProjectID: 1},
				Level:       robotcontroller.LEVELPROJECT,
				Permissions: []*robotcontroller.Permission{{Scope: "/project/1", Kind: "project", Namespace: "example", Access: access}},
			})
			ctx := security.NewContext(context.Background(), secCtx)
			resource := rbacproject.NewNamespace(1).Resource(rbac.ResourcePreatPolicy)
			require.True(t, secCtx.Can(ctx, rbac.ActionRead, resource))
			require.Equal(t, tc.allowUpdate, secCtx.Can(ctx, rbac.ActionUpdate, resource))
			enforcer := &testpreheat.FakeEnforcer{}
			if tc.allowUpdate {
				enforcer.On("EnforcePolicy", mock.Anything, int64(2)).Return(int64(3), nil).Once()
			}
			api := &preheatAPI{projectCtl: projectCtl, preheatCtl: &manualPreheatPolicyController{}, enforcer: enforcer}
			response := httptest.NewRecorder()
			api.ManualPreheat(ctx, operation.ManualPreheatParams{ProjectName: "example", PreheatPolicyName: "policy"}).WriteResponse(response, runtime.JSONProducer())
			require.Equal(t, tc.code, response.Code)
			if tc.allowUpdate {
				require.Equal(t, "/projects/example/preheat/policies/policy/executions/3", response.Header().Get("Location"))
				enforcer.AssertExpectations(t)
			} else {
				enforcer.AssertNotCalled(t, "EnforcePolicy", mock.Anything, mock.Anything)
			}
		})
	}
}
