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

package security

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/lib/config"
)

// authModeErrMgr embeds a real config.Manager and forces Load to fail so that
// config.AuthMode returns an error, exercising the middleware's error path.
type authModeErrMgr struct {
	config.Manager
}

func (m *authModeErrMgr) Load(_ context.Context) error {
	return fmt.Errorf("sectest-authmode: forced config load failure")
}

// TestMiddlewareFailsClosedOnAuthModeError asserts that when the configured auth
// mode cannot be loaded, the security middleware rejects the request (5xx) and
// does not invoke the downstream handler. Before the fix the middleware logged a
// warning and continued to next.ServeHTTP (fail-open).
func TestMiddlewareFailsClosedOnAuthModeError(t *testing.T) {
	real, err := config.GetManager(config.DefaultCfgManager)
	assert.NoError(t, err)

	const name = "sectest-authmode-errload"
	config.Register(name, &authModeErrMgr{Manager: real})

	origMgr := config.DefaultCfgManager
	origGens := generators
	config.DefaultCfgManager = name
	generators = []generator{} // isolate the middleware's own auth-mode handling
	defer func() {
		config.DefaultCfgManager = origMgr
		generators = origGens
	}()

	called := false
	handler := http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		called = true
	})
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1:8080/api/v2.0/projects", nil)

	Middleware()(handler).ServeHTTP(rec, req)

	assert.False(t, called, "handler must not run when the auth mode lookup fails")
	assert.GreaterOrEqual(t, rec.Code, http.StatusInternalServerError, "request must be rejected with a 5xx, got %d", rec.Code)
}
