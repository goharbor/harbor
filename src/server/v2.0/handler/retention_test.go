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
	stderrors "errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/retention"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	retentiontesting "github.com/goharbor/harbor/src/testing/controller/retention"
	metadatatesting "github.com/goharbor/harbor/src/testing/pkg/project/metadata"
)

func TestRetentionHandlersPreservePolicyLookupErrors(t *testing.T) {
	const policyID = int64(42)

	handlers := []struct {
		name   string
		method string
		path   string
		invoke func(*retentionAPI, context.Context, *http.Request) middleware.Responder
	}{
		{
			name:   "get policy",
			method: http.MethodGet,
			path:   "/api/v2.0/retentions/42",
			invoke: func(api *retentionAPI, ctx context.Context, req *http.Request) middleware.Responder {
				return api.GetRetention(ctx, operation.GetRetentionParams{HTTPRequest: req, ID: policyID})
			},
		},
		{
			name:   "delete policy",
			method: http.MethodDelete,
			path:   "/api/v2.0/retentions/42",
			invoke: func(api *retentionAPI, ctx context.Context, req *http.Request) middleware.Responder {
				return api.DeleteRetention(ctx, operation.DeleteRetentionParams{HTTPRequest: req, ID: policyID})
			},
		},
		{
			name:   "trigger execution",
			method: http.MethodPost,
			path:   "/api/v2.0/retentions/42/executions",
			invoke: func(api *retentionAPI, ctx context.Context, req *http.Request) middleware.Responder {
				return api.TriggerRetentionExecution(ctx, operation.TriggerRetentionExecutionParams{
					HTTPRequest: req,
					ID:          policyID,
				})
			},
		},
		{
			name:   "operate execution",
			method: http.MethodPatch,
			path:   "/api/v2.0/retentions/42/executions/7",
			invoke: func(api *retentionAPI, ctx context.Context, req *http.Request) middleware.Responder {
				return api.OperateRetentionExecution(ctx, operation.OperateRetentionExecutionParams{
					HTTPRequest: req,
					ID:          policyID,
					Eid:         7,
					Body:        operation.OperateRetentionExecutionBody{Action: "stop"},
				})
			},
		},
		{
			name:   "list executions",
			method: http.MethodGet,
			path:   "/api/v2.0/retentions/42/executions",
			invoke: func(api *retentionAPI, ctx context.Context, req *http.Request) middleware.Responder {
				return api.ListRetentionExecutions(ctx, operation.ListRetentionExecutionsParams{
					HTTPRequest: req,
					ID:          policyID,
				})
			},
		},
		{
			name:   "list tasks",
			method: http.MethodGet,
			path:   "/api/v2.0/retentions/42/executions/7/tasks",
			invoke: func(api *retentionAPI, ctx context.Context, req *http.Request) middleware.Responder {
				return api.ListRetentionTasks(ctx, operation.ListRetentionTasksParams{
					HTTPRequest: req,
					ID:          policyID,
					Eid:         7,
				})
			},
		},
		{
			name:   "get task log",
			method: http.MethodGet,
			path:   "/api/v2.0/retentions/42/executions/7/tasks/8",
			invoke: func(api *retentionAPI, ctx context.Context, req *http.Request) middleware.Responder {
				return api.GetRetentionTaskLog(ctx, operation.GetRetentionTaskLogParams{
					HTTPRequest: req,
					ID:          policyID,
					Eid:         7,
					Tid:         8,
				})
			},
		},
	}

	errorCases := []struct {
		name       string
		err        error
		statusCode int
		bodyText   string
	}{
		{
			name:       "not found",
			err:        errors.New("retention policy 42 not found").WithCode(errors.NotFoundCode),
			statusCode: http.StatusNotFound,
			bodyText:   errors.NotFoundCode,
		},
		{
			name:       "unexpected lookup error",
			err:        stderrors.New("database unavailable"),
			statusCode: http.StatusInternalServerError,
			bodyText:   "internal server error",
		},
	}

	for _, handler := range handlers {
		for _, errorCase := range errorCases {
			t.Run(handler.name+"/"+errorCase.name, func(t *testing.T) {
				retentionCtl := retentiontesting.NewController(t)
				retentionCtl.On("GetRetention", mock.Anything, policyID).Return(nil, errorCase.err).Once()
				api := &retentionAPI{retentionCtl: retentionCtl}

				responder := handler.invoke(api, context.Background(), httptest.NewRequest(handler.method, handler.path, nil))
				recorder := httptest.NewRecorder()
				responder.WriteResponse(recorder, runtime.JSONProducer())

				require.Equal(t, errorCase.statusCode, recorder.Code, recorder.Body.String())
				require.Contains(t, recorder.Body.String(), errorCase.bodyText)
			})
		}
	}
}

func TestUpdateRetentionReturnsNotFoundForMissingPolicy(t *testing.T) {
	const policyID = int64(42)
	notFoundErr := errors.New("retention policy 42 not found").WithCode(errors.NotFoundCode)

	retentionCtl := retentiontesting.NewController(t)
	retentionCtl.On("UpdateRetention", mock.Anything, mock.Anything).Return(notFoundErr).Once()
	metadataMgr := metadatatesting.NewManager(t)
	metadataMgr.On("Get", mock.Anything, int64(1), "retention_id").Return(map[string]string{"retention_id": "42"}, nil).Once()
	securityCtx := securitytesting.NewContext(t)
	securityCtx.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)

	api := &retentionAPI{retentionCtl: retentionCtl, proMetaMgr: metadataMgr}
	ctx := security.NewContext(context.Background(), securityCtx)
	responder := api.UpdateRetention(ctx, operation.UpdateRetentionParams{
		HTTPRequest: httptest.NewRequest(http.MethodPut, "/api/v2.0/retentions/42", nil),
		ID:          policyID,
		Policy: &models.RetentionPolicy{
			Scope: &models.RetentionPolicyScope{Level: "project", Ref: 1},
		},
	})
	recorder := httptest.NewRecorder()
	responder.WriteResponse(recorder, runtime.JSONProducer())

	require.Equal(t, http.StatusNotFound, recorder.Code, recorder.Body.String())
	require.Contains(t, recorder.Body.String(), errors.NotFoundCode)
}
