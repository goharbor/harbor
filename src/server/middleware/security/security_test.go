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
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/beego/beego/v2/server/web"
	beegosession "github.com/beego/beego/v2/server/web/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/utils/test"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	"github.com/goharbor/harbor/src/lib/orm"
)

func TestMain(m *testing.M) {
	test.InitDatabaseFromEnv()
	os.Exit(m.Run())
}

func TestSecurity(t *testing.T) {
	orig := generators
	defer func() { generators = orig }()

	var ctx security.Context
	var exist bool
	generators = []generator{&unauthorized{}}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, exist = security.FromContext(r.Context())
	})
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/api/users", nil)
	require.Nil(t, err)
	Middleware()(handler).ServeHTTP(nil, req)
	require.True(t, exist)
	assert.NotNil(t, ctx)
}

func initMemorySessionManager(t *testing.T) {
	t.Helper()
	conf := &beegosession.ManagerConfig{
		CookieName:      web.BConfig.WebConfig.Session.SessionName,
		Gclifetime:      web.BConfig.WebConfig.Session.SessionGCMaxLifetime,
		ProviderConfig:  filepath.ToSlash(web.BConfig.WebConfig.Session.SessionProviderConfig),
		Secure:          web.BConfig.Listen.EnableHTTPS,
		EnableSetCookie: web.BConfig.WebConfig.Session.SessionAutoSetCookie,
		Domain:          web.BConfig.WebConfig.Session.SessionDomain,
		CookieLifeTime:  web.BConfig.WebConfig.Session.SessionCookieLifeTime,
	}
	var err error
	web.GlobalSessions, err = beegosession.NewManager("memory", conf)
	require.Nil(t, err)
}

// TestSecuritySkipsSessionWhenAuthorizationPresent covers goharbor/harbor#7704:
// invalid Basic auth must not fall back to the portal session user.
func TestSecuritySkipsSessionWhenAuthorizationPresent(t *testing.T) {
	initMemorySessionManager(t)

	orig := generators
	defer func() { generators = orig }()
	generators = []generator{&basicAuth{}, &session{}}

	user := models.User{
		Username:     "admin",
		UserID:       1,
		Email:        "admin@example.com",
		SysAdminFlag: true,
	}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req = req.WithContext(orm.Context())
	store, err := web.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	require.Nil(t, err)
	require.Nil(t, store.Set(req.Context(), "user", user))

	// Fake credentials with an active portal session — Authorization is present.
	req.SetBasicAuth("fake-user", "wrong-password")

	var ctx security.Context
	var exist bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, exist = security.FromContext(r.Context())
	})
	Middleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	assert.False(t, exist, "security context must not fall back to session when Authorization fails")
	assert.Nil(t, ctx)
}

func TestSecurityUsesSessionWhenNoAuthorization(t *testing.T) {
	initMemorySessionManager(t)

	orig := generators
	defer func() { generators = orig }()
	generators = []generator{&basicAuth{}, &session{}}

	user := models.User{
		Username:     "admin",
		UserID:       1,
		Email:        "admin@example.com",
		SysAdminFlag: true,
	}
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	store, err := web.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	require.Nil(t, err)
	require.Nil(t, store.Set(req.Context(), "user", user))

	var ctx security.Context
	var exist bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, exist = security.FromContext(r.Context())
	})
	Middleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	require.True(t, exist)
	require.NotNil(t, ctx)
	assert.Equal(t, "admin", ctx.GetUsername())
}
