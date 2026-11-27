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
	"strings"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	rbacproject "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/q"
	labelmodel "github.com/goharbor/harbor/src/pkg/label/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/label"
	testsecurity "github.com/goharbor/harbor/src/testing/common/security"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	testlabel "github.com/goharbor/harbor/src/testing/pkg/label"
)

func TestListLabelsAccess(t *testing.T) {
	for _, tc := range []struct {
		name, scope                               string
		withContext, authenticated, projectAccess bool
		code                                      int
	}{
		{"missing-context", common.LabelScopeGlobal, false, false, false, http.StatusUnauthorized},
		{"anonymous-global", common.LabelScopeGlobal, true, false, false, http.StatusUnauthorized},
		{"signed-in-global", common.LabelScopeGlobal, true, true, false, http.StatusOK},
		{"anonymous-public-project", common.LabelScopeProject, true, false, true, http.StatusOK},
		{"anonymous-private-project", common.LabelScopeProject, true, false, false, http.StatusUnauthorized},
		{"signed-in-project-without-access", common.LabelScopeProject, true, true, false, http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			if tc.withContext {
				secCtx := &testsecurity.Context{}
				secCtx.On("IsAuthenticated").Return(tc.authenticated)
				secCtx.On("Can", mock.Anything, rbac.ActionList, rbacproject.NewNamespace(1).Resource(rbac.ResourceLabel)).Return(tc.projectAccess)
				ctx = security.NewContext(ctx, secCtx)
			}
			manager := &testlabel.Manager{}
			if tc.code == http.StatusOK {
				query := mock.MatchedBy(func(query *q.Query) bool {
					return query.Keywords["Scope"] == tc.scope && query.Keywords["Level"] == common.LabelLevelUser
				})
				manager.On("Count", mock.Anything, query).Return(int64(1), nil).Once()
				manager.On("List", mock.Anything, query).Return([]*labelmodel.Label{{ID: 1, Name: "example", Description: "Example label", Scope: tc.scope}}, nil).Once()
			}
			params := operation.NewListLabelsParams()
			params.Scope = &tc.scope
			projectID := int64(1)
			params.ProjectID = &projectID
			params.HTTPRequest = httptest.NewRequest(http.MethodGet, "/api/v2.0/labels", nil)
			response := httptest.NewRecorder()
			(&labelAPI{labelMgr: manager}).ListLabels(ctx, params).WriteResponse(response, runtime.JSONProducer())
			require.Equal(t, tc.code, response.Code)
			if tc.code == http.StatusOK {
				require.Contains(t, response.Body.String(), `"name":"example"`)
				manager.AssertExpectations(t)
			} else {
				manager.AssertNotCalled(t, "Count", mock.Anything, mock.Anything)
				manager.AssertNotCalled(t, "List", mock.Anything, mock.Anything)
			}
		})
	}
}

func TestCreateLabelValidation(t *testing.T) {
	for _, tc := range []struct {
		name  string
		label *models.Label
		code  int
	}{
		{"empty-name", &models.Label{Name: "", Scope: common.LabelScopeGlobal}, http.StatusBadRequest},
		{"name-too-long", &models.Label{Name: strings.Repeat("a", 129), Scope: common.LabelScopeGlobal}, http.StatusBadRequest},
		{"invalid-scope", &models.Label{Name: "l", Scope: "x"}, http.StatusBadRequest},
		{"project-scope-without-project", &models.Label{Name: "l", Scope: common.LabelScopeProject}, http.StatusBadRequest},
		{"missing-project", &models.Label{Name: "l", Scope: common.LabelScopeProject, ProjectID: 99}, http.StatusBadRequest},
		{"project-label", &models.Label{Name: "l", Scope: common.LabelScopeProject, ProjectID: 1}, http.StatusCreated},
		{"global-label", &models.Label{Name: "l", Scope: common.LabelScopeGlobal}, http.StatusCreated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			secCtx := &testsecurity.Context{}
			secCtx.On("IsAuthenticated").Return(true)
			secCtx.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
			ctx := security.NewContext(context.Background(), secCtx)

			projectCtl := &projecttesting.Controller{}
			projectCtl.On("Exists", mock.Anything, int64(1)).Return(true, nil)
			projectCtl.On("Exists", mock.Anything, int64(99)).Return(false, nil)
			manager := &testlabel.Manager{}
			manager.On("Create", mock.Anything, mock.Anything).Return(int64(1), nil)

			response := httptest.NewRecorder()
			(&labelAPI{labelMgr: manager, projectCtl: projectCtl}).CreateLabel(ctx, operation.CreateLabelParams{
				HTTPRequest: httptest.NewRequest(http.MethodPost, "/api/v2.0/labels", nil),
				Label:       tc.label,
			}).WriteResponse(response, runtime.JSONProducer())
			require.Equal(t, tc.code, response.Code, response.Body.String())
			if tc.code != http.StatusCreated {
				manager.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			}
		})
	}
}
