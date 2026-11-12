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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/retention/policy"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/retention"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	retentiontesting "github.com/goharbor/harbor/src/testing/controller/retention"
	metadatatesting "github.com/goharbor/harbor/src/testing/pkg/project/metadata"
)

func TestDeleteMissingRetention(t *testing.T) {
	cases := map[string]struct {
		authenticated bool
		code          int
	}{
		"authenticated": {authenticated: true, code: http.StatusOK},
		"anonymous":     {authenticated: false, code: http.StatusUnauthorized},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctl := &retentiontesting.Controller{}
			ctl.On("GetRetention", mock.Anything, int64(99)).
				Return(nil, errors.NotFoundError(nil).WithMessage("no such Retention policy with id 99"))
			secCtx := &securitytesting.Context{}
			secCtx.On("IsAuthenticated").Return(c.authenticated)
			ctx := security.NewContext(context.Background(), secCtx)

			rr := httptest.NewRecorder()
			api := &retentionAPI{retentionCtl: ctl}
			api.DeleteRetention(ctx, operation.DeleteRetentionParams{ID: 99}).WriteResponse(rr, runtime.JSONProducer())
			assert.Equal(t, c.code, rr.Code)
			ctl.AssertNotCalled(t, "DeleteRetention", mock.Anything, mock.Anything)
		})
	}
}

func TestDeleteRetentionRemovedConcurrently(t *testing.T) {
	notFound := errors.NotFoundError(nil).WithMessage("not found")
	p := &policy.Metadata{ID: 99, Scope: &policy.Scope{Level: "project", Reference: 1}}
	cases := map[string]struct {
		recheck    *policy.Metadata
		recheckErr error
		code       int
	}{
		"policy gone":         {recheckErr: notFound, code: http.StatusOK},
		"policy still exists": {recheck: p, code: http.StatusNotFound},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			ctl := &retentiontesting.Controller{}
			ctl.On("GetRetention", mock.Anything, int64(99)).Return(p, nil).Once()
			ctl.On("DeleteRetention", mock.Anything, int64(99)).Return(notFound)
			ctl.On("GetRetention", mock.Anything, int64(99)).Return(c.recheck, c.recheckErr).Once()
			metaMgr := &metadatatesting.Manager{}
			secCtx := &securitytesting.Context{}
			secCtx.On("IsAuthenticated").Return(true)
			secCtx.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
			ctx := security.NewContext(context.Background(), secCtx)

			rr := httptest.NewRecorder()
			api := &retentionAPI{retentionCtl: ctl, proMetaMgr: metaMgr}
			api.DeleteRetention(ctx, operation.DeleteRetentionParams{ID: 99}).WriteResponse(rr, runtime.JSONProducer())
			assert.Equal(t, c.code, rr.Code)
			metaMgr.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestGetMissingRetention(t *testing.T) {
	ctl := &retentiontesting.Controller{}
	ctl.On("GetRetention", mock.Anything, int64(99)).
		Return(nil, errors.NotFoundError(nil).WithMessage("no such Retention policy with id 99"))
	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(true)
	ctx := security.NewContext(context.Background(), secCtx)

	rr := httptest.NewRecorder()
	api := &retentionAPI{retentionCtl: ctl}
	api.GetRetention(ctx, operation.GetRetentionParams{ID: 99}).WriteResponse(rr, runtime.JSONProducer())
	assert.Equal(t, http.StatusNotFound, rr.Code)
}
