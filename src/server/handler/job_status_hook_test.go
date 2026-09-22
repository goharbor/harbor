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
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common/security"
	securitytesting "github.com/goharbor/harbor/src/testing/common/security"
)

func TestJobStatusHandler_NoSecurityContext(t *testing.T) {
	handler := NewJobStatusHandler()
	body := strings.NewReader(`{"job_id":"test-job-1","status":"Success"}`)
	req := httptest.NewRequest(http.MethodPost, "/service/notifications/tasks/1", body)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized")
}

func TestJobStatusHandler_UnauthenticatedUser(t *testing.T) {
	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(false)

	handler := NewJobStatusHandler()
	body := strings.NewReader(`{"job_id":"test-job-1","status":"Success"}`)
	req := httptest.NewRequest(http.MethodPost, "/service/notifications/tasks/1", body)
	req = req.WithContext(security.NewContext(req.Context(), secCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized")
}

func TestJobStatusHandler_NonSolutionUser(t *testing.T) {
	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(true)
	secCtx.On("IsSolutionUser").Return(false)

	handler := NewJobStatusHandler()
	body := strings.NewReader(`{"job_id":"test-job-1","status":"Success"}`)
	req := httptest.NewRequest(http.MethodPost, "/service/notifications/tasks/1", body)
	req = req.WithContext(security.NewContext(req.Context(), secCtx))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)
	assert.Contains(t, rr.Body.String(), "unauthorized")
}

func TestJobStatusHandler_SolutionUser(t *testing.T) {
	secCtx := &securitytesting.Context{}
	secCtx.On("IsAuthenticated").Return(true)
	secCtx.On("IsSolutionUser").Return(true)

	handler := NewJobStatusHandler()
	body := strings.NewReader(`{"job_id":"test-job-1","status":"Success"}`)
	req := httptest.NewRequest(http.MethodPost, "/service/notifications/tasks/1", body)
	req = req.WithContext(security.NewContext(req.Context(), secCtx))
	rr := httptest.NewRecorder()

	// Passes auth, panics in Handle() due to nil DAO in test.
	assert.Panics(t, func() {
		handler.ServeHTTP(rr, req)
	})
	assert.NotEqual(t, http.StatusUnauthorized, rr.Code)
}
