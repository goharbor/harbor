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
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/system_cve_allowlist"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
	allowlisttesting "github.com/goharbor/harbor/src/testing/pkg/allowlist"
)

func TestPutSystemCVEAllowlistNilItem(t *testing.T) {
	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(true)
	secCtx.On("Can", mock.Anything, mock.Anything, mock.Anything).Return(true)
	ctx := security.NewContext(context.Background(), secCtx)
	mgr := &allowlisttesting.Manager{}

	rr := httptest.NewRecorder()
	systemCVEAllowListAPI{mgr: mgr}.PutSystemCVEAllowlist(ctx, system_cve_allowlist.PutSystemCVEAllowlistParams{
		HTTPRequest: httptest.NewRequest(http.MethodPut, "/api/v2.0/system/CVEAllowlist", nil),
		Allowlist:   &models.CVEAllowlist{Items: []*models.CVEAllowlistItem{nil}},
	}).WriteResponse(rr, runtime.JSONProducer())
	assert.Equal(t, http.StatusBadRequest, rr.Code)
	mgr.AssertNotCalled(t, "SetSys", mock.Anything, mock.Anything)
}
