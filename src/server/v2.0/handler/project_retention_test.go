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
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	usermodel "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	configmetadata "github.com/goharbor/harbor/src/lib/config/metadata"
	regmodel "github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/project"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	registrytesting "github.com/goharbor/harbor/src/testing/controller/registry"
	retentiontesting "github.com/goharbor/harbor/src/testing/controller/retention"
	usertesting "github.com/goharbor/harbor/src/testing/controller/user"
	configtesting "github.com/goharbor/harbor/src/testing/lib/config"
	metadatatesting "github.com/goharbor/harbor/src/testing/pkg/project/metadata"
)

func TestCreateProjectRetention(t *testing.T) {
	t.Setenv("PERMITTED_REGISTRY_TYPES_FOR_PROXY_CACHE", regmodel.RegistryTypeDockerRegistry)
	cfg := configtesting.NewManager(t)
	cfg.On("Load", mock.Anything).Return(nil)
	cfg.On("Get", mock.Anything, common.ProjectCreationRestriction).Return(&configmetadata.ConfigureValue{
		Name: common.ProjectCreationRestriction, Value: common.ProCrtRestrEveryone,
	})
	cfg.On("Get", mock.Anything, common.QuotaPerProjectEnable).Return(&configmetadata.ConfigureValue{
		Name: common.QuotaPerProjectEnable, Value: "false",
	})
	originalConfig := config.DefaultCfgManager
	config.Register("project-retention-test", cfg)
	config.DefaultCfgManager = "project-retention-test"
	t.Cleanup(func() { config.DefaultCfgManager = originalConfig })

	tests := []struct {
		name        string
		registryID  *int64
		days        *int64
		retentionID *string
		wantDays    int
		wantStatus  int
	}{
		{name: "default", registryID: swag.Int64(1), wantDays: 7, wantStatus: http.StatusCreated},
		{name: "custom", registryID: swag.Int64(1), days: swag.Int64(30), wantDays: 30, wantStatus: http.StatusCreated},
		{name: "zero", registryID: swag.Int64(1), days: swag.Int64(0), wantDays: 0, wantStatus: http.StatusCreated},
		{name: "maximum", registryID: swag.Int64(1), days: swag.Int64(18250), wantDays: 18250, wantStatus: http.StatusCreated},
		{name: "empty retention ID", registryID: swag.Int64(1), days: swag.Int64(30), retentionID: swag.String(""), wantDays: 30, wantStatus: http.StatusCreated},
		{name: "retention ID with days", registryID: swag.Int64(1), days: swag.Int64(30), retentionID: swag.String("73"), wantStatus: http.StatusBadRequest},
		{name: "retention ID with zero days", registryID: swag.Int64(1), days: swag.Int64(0), retentionID: swag.String("73"), wantStatus: http.StatusBadRequest},
		{name: "ordinary project", wantStatus: http.StatusCreated},
		{name: "ordinary project with retention", days: swag.Int64(30), wantStatus: http.StatusBadRequest},
		{name: "ordinary project with zero retention", days: swag.Int64(0), wantStatus: http.StatusBadRequest},
		{name: "negative", registryID: swag.Int64(1), days: swag.Int64(-1), wantStatus: http.StatusBadRequest},
		{name: "above maximum", registryID: swag.Int64(1), days: swag.Int64(18251), wantStatus: http.StatusBadRequest},
		{name: "duration overflow", registryID: swag.Int64(1), days: swag.Int64(106752), wantStatus: http.StatusBadRequest},
		{name: "integer overflow", registryID: swag.Int64(1), days: swag.Int64(math.MaxInt64), wantStatus: http.StatusBadRequest},
		{name: "invalid registry", registryID: swag.Int64(0), days: swag.Int64(30), wantStatus: http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			projectCtl := projecttesting.NewController(t)
			retentionCtl := retentiontesting.NewController(t)
			metadataMgr := metadatatesting.NewManager(t)
			userCtl := usertesting.NewController(t)
			registryCtl := registrytesting.NewController(t)
			originalRegistry := registry.Ctl
			registry.Ctl = registryCtl
			t.Cleanup(func() { registry.Ctl = originalRegistry })

			sec := securitytesting.NewContext(t)
			sec.On("IsAuthenticated").Return(true)
			if tt.registryID != nil {
				sec.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
			}
			ctx := security.NewContext(context.Background(), sec)

			if tt.wantStatus == http.StatusCreated {
				sec.On("IsSolutionUser").Return(false)
				sec.On("GetUsername").Return("admin")
				userCtl.On("GetByName", mock.Anything, "admin").Return(&usermodel.User{UserID: 1}, nil).Once()
				projectCtl.On("Create", mock.Anything, mock.Anything).Return(int64(42), nil).Run(func(args mock.Arguments) {
					p := args.Get(1).(*project.Project)
					require.Equal(t, "proxy-project", p.Name)
					require.Equal(t, lib.Int64Value(tt.registryID), p.RegistryID)
				}).Once()
				if tt.registryID != nil {
					registryCtl.On("Get", mock.Anything, int64(1)).Return(&regmodel.Registry{Type: regmodel.RegistryTypeDockerRegistry}, nil).Once()
					retentionCtl.On("CreateRetention", mock.Anything, mock.Anything).Return(int64(73), nil).Run(func(args mock.Arguments) {
						p := args.Get(1).(*policy.Metadata)
						require.Equal(t, int64(42), p.Scope.Reference)
						require.Equal(t, "project", p.Scope.Level)
						require.Equal(t, "Schedule", p.Trigger.Kind)
						require.Equal(t, "0 0 0 * * *", p.Trigger.Settings["cron"])
						require.Len(t, p.Rules, 1)
						rule := p.Rules[0]
						require.Equal(t, "retain", rule.Action)
						require.Equal(t, "nDaysSinceLastPull", rule.Template)
						require.Equal(t, tt.wantDays, rule.Parameters["nDaysSinceLastPull"])
						require.Equal(t, "**", rule.TagSelectors[0].Pattern)
						require.Equal(t, "**", rule.ScopeSelectors["repository"][0].Pattern)
					}).Once()
					metadataMgr.On("Add", mock.Anything, int64(42), map[string]string{"retention_id": "73"}).Return(nil).Once()
				}
			}

			api := &projectAPI{projectCtl: projectCtl, retentionCtl: retentionCtl, metadataMgr: metadataMgr, userCtl: userCtl}
			res := api.CreateProject(ctx, operation.CreateProjectParams{
				HTTPRequest: httptest.NewRequest(http.MethodPost, "/api/v2.0/projects", nil),
				Project: &models.ProjectReq{
					ProjectName: "proxy-project", RegistryID: tt.registryID, RetentionDays: tt.days,
					Metadata: &models.ProjectMetadata{RetentionID: tt.retentionID},
				},
			})
			rw := httptest.NewRecorder()
			res.WriteResponse(rw, runtime.JSONProducer())
			require.Equal(t, tt.wantStatus, rw.Code, rw.Body.String())
			if tt.wantStatus == http.StatusCreated {
				require.Equal(t, "/api/v2.0/projects/42", rw.Header().Get("Location"))
			} else {
				projectCtl.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
			}
			if tt.registryID == nil || tt.wantStatus != http.StatusCreated {
				retentionCtl.AssertNotCalled(t, "CreateRetention", mock.Anything, mock.Anything)
				metadataMgr.AssertNotCalled(t, "Add", mock.Anything, mock.Anything, mock.Anything)
			}
		})
	}
}

func TestProjectRetentionRequestValidation(t *testing.T) {
	for _, tt := range []struct {
		body string
		days *int64
	}{
		{body: `{}`},
		{body: `{"retention_days": null}`},
		{body: `{"retention_days": 7}`, days: swag.Int64(7)},
		{body: `{"retention_days": 30}`, days: swag.Int64(30)},
		{body: `{"retention_days": 0}`, days: swag.Int64(0)},
		{body: `{"retention_days": 18250}`, days: swag.Int64(18250)},
	} {
		t.Run(tt.body, func(t *testing.T) {
			var req models.ProjectReq
			require.NoError(t, json.Unmarshal([]byte(tt.body), &req))
			require.NoError(t, req.Validate(strfmt.Default))
			require.Equal(t, tt.days, req.RetentionDays)
		})
	}

	for _, body := range []string{
		`{"retention_days": -1}`,
		`{"retention_days": 18251}`,
		`{"retention_days": 106752}`,
		`{"retention_days": 1.5}`,
		`{"retention_days": "30"}`,
		`{"retention_days": 9223372036854775808}`,
	} {
		t.Run(body, func(t *testing.T) {
			var req models.ProjectReq
			err := json.Unmarshal([]byte(body), &req)
			if err == nil {
				err = req.Validate(strfmt.Default)
			}
			require.Error(t, err)
		})
	}
}

func TestUpdateProjectRetention(t *testing.T) {
	for _, projectType := range []struct {
		name       string
		registryID int64
	}{
		{name: "ordinary project"},
		{name: "proxy cache project", registryID: 1},
	} {
		t.Run(projectType.name, func(t *testing.T) {
			for _, tt := range []struct {
				name       string
				body       string
				wantStatus int
			}{
				{name: "omitted", body: `{"metadata":{"public":"true"}}`, wantStatus: http.StatusOK},
				{name: "null", body: `{"metadata":{"public":"true"},"retention_days":null}`, wantStatus: http.StatusOK},
				{name: "zero", body: `{"metadata":{"public":"true"},"retention_days":0}`, wantStatus: http.StatusBadRequest},
				{name: "custom", body: `{"metadata":{"public":"true"},"retention_days":30}`, wantStatus: http.StatusBadRequest},
				{name: "with registry", body: `{"metadata":{"public":"true"},"retention_days":30,"registry_id":1}`, wantStatus: http.StatusBadRequest},
			} {
				t.Run(tt.name, func(t *testing.T) {
					projectCtl := projecttesting.NewController(t)
					retentionCtl := retentiontesting.NewController(t)
					p := &project.Project{
						ProjectID: 42, Name: "test-project", RegistryID: projectType.registryID,
						Metadata: map[string]string{"public": "false"},
					}
					projectCtl.On("Get", mock.Anything, int64(42)).Return(p, nil).Once()
					projectCtl.On("Update", mock.Anything, p).Return(nil).Maybe()

					sec := securitytesting.NewContext(t)
					sec.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
					ctx := security.NewContext(context.Background(), sec)

					var req models.ProjectReq
					require.NoError(t, json.Unmarshal([]byte(tt.body), &req))
					api := &projectAPI{projectCtl: projectCtl, retentionCtl: retentionCtl}
					res := api.UpdateProject(ctx, operation.UpdateProjectParams{
						ProjectNameOrID: "42", Project: &req,
					})
					rw := httptest.NewRecorder()
					res.WriteResponse(rw, runtime.JSONProducer())
					require.Equal(t, tt.wantStatus, rw.Code, rw.Body.String())
					if tt.wantStatus == http.StatusOK {
						projectCtl.AssertNumberOfCalls(t, "Update", 1)
						require.Equal(t, "true", p.Metadata["public"])
					} else {
						require.Contains(t, rw.Body.String(), "retention_days is only supported when creating a proxy cache project")
						projectCtl.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
						require.Equal(t, "false", p.Metadata["public"])
					}
				})
			}
		})
	}
}
