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

package notification

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mockjobservice "github.com/goharbor/harbor/src/testing/jobservice"
)

// useTestHTTPClient swaps the package notification clients for a test client for the duration of
// the test, so tests can drive the jobs against a loopback httptest server that the production
// public-network guard would otherwise refuse to dial.
func useTestHTTPClient(t *testing.T, client *http.Client) {
	t.Helper()
	original := httpHelper.clients
	httpHelper.clients = map[string]*http.Client{
		secure:   client,
		insecure: client,
	}
	t.Cleanup(func() {
		httpHelper.clients = original
	})
}

// TestWebhookJobDropsResponseBody is the webhook read-back regression: on a non-2xx
// response the webhook job must not fold the target's response body into the error that is
// persisted to the (project-readable) task log. Red on main (body reflected), green after the fix.
func TestWebhookJobDropsResponseBody(t *testing.T) {
	const secret = "SECTEST-WEBHOOK-INTERNAL-RESPONSE-BODY-AKIAFAKE0000"

	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(secret))
	}))
	defer ts.Close()
	useTestHTTPClient(t, ts.Client())

	rep := &WebhookJob{}
	err := rep.Run(ctx, map[string]any{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          ts.URL,
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "abnormal response code: 403")
	assert.NotContains(t, err.Error(), secret,
		"internal response body must not be reflected into the webhook task log")
}

// TestSlackJobDropsResponseBody is the same read-back regression for the Slack notification job.
func TestSlackJobDropsResponseBody(t *testing.T) {
	const secret = "SECTEST-SLACK-INTERNAL-RESPONSE-BODY-AKIAFAKE0000"

	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(secret))
	}))
	defer ts.Close()
	useTestHTTPClient(t, ts.Client())

	rep := &SlackJob{}
	err := rep.Run(ctx, map[string]any{
		"skip_cert_verify": true,
		"payload":          `{"key": "value"}`,
		"address":          ts.URL,
	})

	require.Error(t, err)
	assert.NotContains(t, err.Error(), secret,
		"internal response body must not be reflected into the slack task log")
}

// TestNotificationClientsBlockRedirect asserts the notification clients refuse to follow
// redirects, so a redirector cannot pivot a webhook to a second (internal) origin. It references
// only the stable CheckRedirect field, so on a reverted fix it fails on assertion (nil) rather
// than a compile error.
func TestNotificationClientsBlockRedirect(t *testing.T) {
	for name, client := range httpHelper.clients {
		require.NotNil(t, client.CheckRedirect, "client %q must set CheckRedirect", name)
		err := client.CheckRedirect(httptest.NewRequest(http.MethodGet, "http://example.com", nil), nil)
		assert.ErrorIs(t, err, http.ErrUseLastResponse, "client %q must not follow redirects", name)
	}
}

// TestNotificationClientsKeepProxy asserts the notification transports still consult the
// jobservice HTTP_PROXY/HTTPS_PROXY/NO_PROXY settings instead of always dialing directly.
func TestNotificationClientsKeepProxy(t *testing.T) {
	for name, client := range httpHelper.clients {
		transport, ok := client.Transport.(*http.Transport)
		require.True(t, ok, "client %q must use *http.Transport", name)
		assert.NotNil(t, transport.Proxy, "client %q must honour the proxy environment", name)
	}
}

// TestNotificationClientsGuardPrivateTargets asserts the production notification clients carry
// the public-network dial guard: a loopback target is refused before any request reaches it.
func TestNotificationClientsGuardPrivateTargets(t *testing.T) {
	t.Setenv("HARBOR_ALLOW_PRIVATE_NETWORK_ACCESS", "false")

	var hits atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	require.NotEmpty(t, httpHelper.clients)
	for name, client := range httpHelper.clients {
		req, err := http.NewRequest(http.MethodPost, ts.URL, strings.NewReader(`{"key": "value"}`))
		require.NoError(t, err)
		resp, err := client.Do(req)
		if resp != nil {
			_ = resp.Body.Close()
		}
		assert.ErrorContains(t, err, "private network address", "client %q must refuse a loopback target", name)
	}
	assert.Zero(t, hits.Load(), "no request may reach a private target")
}

// TestWebhookJobReusesConnectionOnErrorResponse asserts a non-2xx response body is drained
// (without being reflected) so the keep-alive connection goes back to the pool.
func TestWebhookJobReusesConnectionOnErrorResponse(t *testing.T) {
	ctx := &mockjobservice.MockJobContext{}
	logger := &mockjobservice.MockJobLogger{}
	ctx.On("GetLogger").Return(logger)

	var newConns atomic.Int32
	ts := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal error details"))
	}))
	ts.Config.ConnState = func(_ net.Conn, state http.ConnState) {
		if state == http.StateNew {
			newConns.Add(1)
		}
	}
	ts.Start()
	defer ts.Close()
	useTestHTTPClient(t, ts.Client())

	for range 2 {
		err := (&WebhookJob{}).Run(ctx, map[string]any{
			"payload": `{"key": "value"}`,
			"address": ts.URL,
		})
		require.ErrorContains(t, err, "abnormal response code: 500")
		assert.NotContains(t, err.Error(), "internal error details")
	}
	assert.Equal(t, int32(1), newConns.Load(), "the connection must be reused after a non-2xx response")
}
