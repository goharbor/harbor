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

	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/retention"
	retentiontesting "github.com/goharbor/harbor/src/testing/controller/retention"
)

func TestRetentionPolicyRequiredFields(t *testing.T) {
	scope := &models.RetentionPolicyScope{Level: "project", Ref: 1}
	trigger := &models.RetentionRuleTrigger{Kind: "Schedule", Settings: map[string]any{"cron": ""}}
	cases := map[string]*models.RetentionPolicy{
		"missing scope":   {Algorithm: "or", Trigger: trigger},
		"missing trigger": {Algorithm: "or", Scope: scope},
		"empty policy":    {},
	}
	for name, policy := range cases {
		t.Run(name, func(t *testing.T) {
			ctl := &retentiontesting.Controller{}
			api := &retentionAPI{retentionCtl: ctl}

			rr := httptest.NewRecorder()
			api.CreateRetention(context.Background(), operation.CreateRetentionParams{
				HTTPRequest: httptest.NewRequest(http.MethodPost, "/api/v2.0/retentions", nil),
				Policy:      policy,
			}).WriteResponse(rr, runtime.JSONProducer())
			assert.Equal(t, http.StatusBadRequest, rr.Code)

			rr = httptest.NewRecorder()
			api.UpdateRetention(context.Background(), operation.UpdateRetentionParams{
				HTTPRequest: httptest.NewRequest(http.MethodPut, "/api/v2.0/retentions/1", nil),
				ID:          1,
				Policy:      policy,
			}).WriteResponse(rr, runtime.JSONProducer())
			assert.Equal(t, http.StatusBadRequest, rr.Code)
			ctl.AssertNotCalled(t, "CreateRetention")
		})
	}
}
