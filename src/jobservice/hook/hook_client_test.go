// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package hook

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/jobservice/job"
)

// HookClientTestSuite tests functions of hook client
type HookClientTestSuite struct {
	suite.Suite

	mockServer *httptest.Server
	client     Client
}

// SetupSuite prepares test suite
func (suite *HookClientTestSuite) SetupSuite() {
	suite.client = NewClient(context.Background())
	suite.mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		change := &job.StatusChange{}
		err = json.Unmarshal(bytes, change)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if change.JobID == "job_ID_failed" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, _ = fmt.Fprintln(w, "ok")
	}))
}

// TearDownSuite clears test suite
func (suite *HookClientTestSuite) TearDownSuite() {
	suite.mockServer.Close()
}

// TestHookClientTestSuite is entry of go test
func TestHookClientTestSuite(t *testing.T) {
	suite.Run(t, new(HookClientTestSuite))
}

// TestHookClient ...
func (suite *HookClientTestSuite) TestHookClient() {
	changeData := &job.StatusChange{
		JobID:  "fake_job_ID",
		Status: "running",
	}
	evt := &Event{
		URL:       suite.mockServer.URL,
		Data:      changeData,
		Message:   fmt.Sprintf("Status of job %s changed to: %s", changeData.JobID, changeData.Status),
		Timestamp: time.Now().Unix(),
	}
	err := suite.client.SendEvent(evt)
	assert.Nil(suite.T(), err, "send event: nil error expected but got %s", err)
}

// TestReportStatusFailed ...
func (suite *HookClientTestSuite) TestReportStatusFailed() {
	changeData := &job.StatusChange{
		JobID:  "job_ID_failed",
		Status: "running",
	}
	evt := &Event{
		URL:       suite.mockServer.URL,
		Data:      changeData,
		Message:   fmt.Sprintf("Status of job %s changed to: %s", changeData.JobID, changeData.Status),
		Timestamp: time.Now().Unix(),
	}

	err := suite.client.SendEvent(evt)
	assert.NotNil(suite.T(), err, "send event: expected non nil error but got nil")
}

func TestCoreNotificationCallbackCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/service/notifications/") {
			require.Equal(t, "jobservice-test", secret.FromRequest(r))
		} else {
			require.Empty(t, r.Header.Get("Authorization"))
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("CORE_URL", server.URL)
	t.Setenv("JOBSERVICE_SECRET", "jobservice-test")
	client := NewClient(context.Background())
	for _, path := range []string{"/service/notifications/tasks/12", "/service/notifications/jobs/adminjob/1", "/other-hook"} {
		event := &Event{URL: server.URL + path, Data: &job.StatusChange{JobID: "job-1"}, Message: "test", Timestamp: time.Now().Unix()}
		require.NoError(t, client.SendEvent(event))
	}
}

func TestCoreNotificationCallbackMissingCredential(t *testing.T) {
	t.Setenv("CORE_URL", "http://core.example")
	t.Setenv("JOBSERVICE_SECRET", "")
	event := &Event{URL: "http://core.example/service/notifications/tasks/12", Data: &job.StatusChange{JobID: "job-1"}, Message: "test", Timestamp: time.Now().Unix()}
	require.EqualError(t, NewClient(context.Background()).SendEvent(event), "jobservice callback credential is not configured")
}

func TestCoreNotificationRedirectBoundary(t *testing.T) {
	for _, sameHost := range []bool{false, true} {
		t.Run(fmt.Sprintf("same-host-%t", sameHost), func(t *testing.T) {
			targetReached := false
			target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				targetReached = true
				w.WriteHeader(http.StatusOK)
			}))
			defer target.Close()
			core := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path == "/other-hook" {
					targetReached = true
					w.WriteHeader(http.StatusOK)
					return
				}
				require.Equal(t, "jobservice-test", secret.FromRequest(r))
				location := target.URL + "/service/notifications/tasks/12"
				if sameHost {
					location = "/other-hook"
				}
				http.Redirect(w, r, location, http.StatusTemporaryRedirect)
			}))
			defer core.Close()
			t.Setenv("CORE_URL", core.URL)
			t.Setenv("JOBSERVICE_SECRET", "jobservice-test")
			event := &Event{URL: core.URL + "/service/notifications/tasks/12", Data: &job.StatusChange{JobID: "job-1"}, Message: "test", Timestamp: time.Now().Unix()}
			require.Error(t, NewClient(context.Background()).SendEvent(event))
			require.False(t, targetReached)
		})
	}
}

func TestIsCoreNotificationURL(t *testing.T) {
	t.Setenv("CORE_URL", "http://core.example:8080")
	for _, tc := range []struct {
		url   string
		match bool
	}{
		{"http://core.example:8080/service/notifications/tasks/1", true},
		{"http://other.example:8080/service/notifications/tasks/1", false},
		{"http://core.example:8081/service/notifications/tasks/1", false},
		{"https://core.example:8080/service/notifications/tasks/1", false},
		{"http://core.example:8080/other-hook", false},
		{"http://core.example:8080/service/notifications-other/tasks/1", false},
		{"http://core.example:8080/service/notifications/../token", false},
		{"http://core.example:8080/service/notifications/%2e%2e/token", false},
		{"http://user@core.example:8080/service/notifications/tasks/1", false},
		{"HTTP://core.example:8080/service/notifications/tasks/1", true},
		{"Http://core.example:8080/service/notifications/tasks/1", true},
	} {
		t.Run(tc.url, func(t *testing.T) {
			target, err := url.Parse(tc.url)
			require.NoError(t, err)
			require.Equal(t, tc.match, isCoreNotificationURL(target))
		})
	}

	t.Run("uppercase CORE_URL scheme", func(t *testing.T) {
		t.Setenv("CORE_URL", "HTTP://core.example:8080")
		target, err := url.Parse("http://core.example:8080/service/notifications/tasks/1")
		require.NoError(t, err)
		require.True(t, isCoreNotificationURL(target))
	})
}

func TestExternalCallbackWithoutCredential(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Empty(t, r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	t.Setenv("CORE_URL", "http://core.example:8080")
	t.Setenv("JOBSERVICE_SECRET", "jobservice-test")
	event := &Event{URL: server.URL + "/service/notifications/tasks/12", Data: &job.StatusChange{JobID: "job-1"}, Message: "test", Timestamp: time.Now().Unix()}
	require.NoError(t, NewClient(context.Background()).SendEvent(event))
}
