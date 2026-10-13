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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/rbac"
	rbacproject "github.com/goharbor/harbor/src/common/rbac/project"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/q"
	labelmodel "github.com/goharbor/harbor/src/pkg/label/model"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/label"
	testsecurity "github.com/goharbor/harbor/src/testing/common/security"
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
