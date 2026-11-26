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

	beegocontext "github.com/beego/beego/v2/server/web/context"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/security"
	securitysecret "github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/server/router"
	testsecurity "github.com/goharbor/harbor/src/testing/common/security"
)

type recordingHookHandler struct {
	called bool
	taskID int64
	change *job.StatusChange
	err    error
}

func (h *recordingHookHandler) Handle(_ context.Context, change *job.StatusChange) error {
	h.called, h.change = true, change
	return h.err
}

func (h *recordingHookHandler) HandleForTask(ctx context.Context, id int64, change *job.StatusChange) error {
	h.taskID = id
	return h.Handle(ctx, change)
}

func TestJobStatusHandlerRequiresServiceIdentity(t *testing.T) {
	store := secret.NewStore(map[string]string{"jobservice-test": secret.JobserviceUser, "core-test": secret.CoreUser})
	for _, identity := range []string{"missing", "ordinary-user", "invalid-secret", "jobservice", "core"} {
		t.Run(identity, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/service/notifications/jobs/adminjob/1", strings.NewReader(`{"job_id":"job-1","metadata":{}}`))
			ctx := req.Context()
			allowed := false
			switch identity {
			case "ordinary-user":
				secCtx := &testsecurity.Context{}
				secCtx.On("IsSolutionUser").Return(false)
				ctx = security.NewContext(ctx, secCtx)
			case "invalid-secret":
				ctx = security.NewContext(ctx, securitysecret.NewSecurityContext("invalid-test", store))
			case "jobservice", "core":
				allowed = true
				ctx = security.NewContext(ctx, securitysecret.NewSecurityContext(identity+"-test", store))
			}
			hook := &recordingHookHandler{}
			response := httptest.NewRecorder()
			(&jobStatusHandler{handler: hook}).ServeHTTP(response, req.WithContext(ctx))
			if allowed {
				require.Equal(t, http.StatusOK, response.Code)
				require.True(t, hook.called)
			} else {
				require.Equal(t, http.StatusUnauthorized, response.Code)
				require.False(t, hook.called)
			}
		})
	}
}

func TestTaskJobStatusHandlerValidation(t *testing.T) {
	for _, tc := range []struct {
		name, id, body string
		code           int
	}{
		{"valid", "12", `{"job_id":"job-1","metadata":{}}`, http.StatusOK},
		{"periodic", "12", `{"job_id":"job-1@1605104520","metadata":{"upstream_job_id":"job-1"},"check_in":"checkin"}`, http.StatusOK},
		{"invalid-id", "abc", `{"job_id":"job-1","metadata":{}}`, http.StatusBadRequest},
		{"zero-id", "0", `{"job_id":"job-1","metadata":{}}`, http.StatusBadRequest},
		{"negative-id", "-1", `{"job_id":"job-1","metadata":{}}`, http.StatusBadRequest},
		{"missing-id", "", `{"job_id":"job-1","metadata":{}}`, http.StatusBadRequest},
		{"invalid-json", "12", `{`, http.StatusBadRequest},
		{"missing-job", "12", `{"metadata":{}}`, http.StatusBadRequest},
		{"missing-metadata", "12", `{"job_id":"job-1"}`, http.StatusBadRequest},
		{"null-metadata", "12", `{"job_id":"job-1","metadata":null}`, http.StatusBadRequest},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/service/notifications/tasks/"+tc.id, strings.NewReader(tc.body))
			input := beegocontext.NewInput()
			input.SetParam(":id", tc.id)
			ctx := context.WithValue(req.Context(), router.ContextKeyInput{}, input)
			store := secret.NewStore(map[string]string{"test": secret.JobserviceUser})
			ctx = security.NewContext(ctx, securitysecret.NewSecurityContext("test", store))
			hook := &recordingHookHandler{}
			response := httptest.NewRecorder()
			(&jobStatusHandler{handler: hook, bindTaskID: true}).ServeHTTP(response, req.WithContext(ctx))
			require.Equal(t, tc.code, response.Code)
			require.Equal(t, tc.code == http.StatusOK, hook.called)
			if hook.called {
				require.Equal(t, int64(12), hook.taskID)
				require.True(t, strings.HasPrefix(hook.change.JobID, "job-1"), "job ID must not be replaced with the numeric task ID")
			}
		})
	}
}

func TestJobStatusHandlerPreservesRetryResponse(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/service/notifications/jobs/replication/1", strings.NewReader(`{"job_id":"job-1","metadata":{}}`))
	store := secret.NewStore(map[string]string{"test": secret.JobserviceUser})
	ctx := security.NewContext(req.Context(), securitysecret.NewSecurityContext("test", store))
	hook := &recordingHookHandler{err: errors.NotFoundError(nil)}
	response := httptest.NewRecorder()
	(&jobStatusHandler{handler: hook}).ServeHTTP(response, req.WithContext(ctx))
	require.Equal(t, http.StatusNotFound, response.Code)
	require.Equal(t, int64(0), hook.taskID, "legacy routes retain job-based lookup")
}
