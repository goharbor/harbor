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

package session

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/beego/beego/v2/server/web"
	beegosession "github.com/beego/beego/v2/server/web/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	_ "github.com/goharbor/harbor/src/pkg/config/inmemory"
)

func TestSession(t *testing.T) {
	config.InitWithSettings(map[string]any{})
	carrySession := false
	skipRenewal := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		carrySession = lib.GetCarrySession(r.Context())
		skipRenewal = lib.GetSkipSessionRenewal(r.Context())
	})
	// no session
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/api/users", nil)
	require.Nil(t, err)
	Middleware()(handler).ServeHTTP(nil, req)
	assert.False(t, carrySession)
	assert.False(t, skipRenewal)

	// contains session
	web.BConfig.WebConfig.Session.SessionName = config.SessionCookieName
	conf := &beegosession.ManagerConfig{
		CookieName:      web.BConfig.WebConfig.Session.SessionName,
		Gclifetime:      web.BConfig.WebConfig.Session.SessionGCMaxLifetime,
		ProviderConfig:  filepath.ToSlash(web.BConfig.WebConfig.Session.SessionProviderConfig),
		Secure:          web.BConfig.Listen.EnableHTTPS,
		EnableSetCookie: web.BConfig.WebConfig.Session.SessionAutoSetCookie,
		Domain:          web.BConfig.WebConfig.Session.SessionDomain,
		CookieLifeTime:  web.BConfig.WebConfig.Session.SessionCookieLifeTime,
	}
	web.GlobalSessions, err = beegosession.NewManager("memory", conf)
	require.Nil(t, err)
	_, err = web.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	require.Nil(t, err)
	Middleware()(handler).ServeHTTP(nil, req)
	assert.True(t, carrySession)
	assert.False(t, skipRenewal)

	// With no-session-renewal header.
	req.Header.Set(HeaderNoSessionRenewal, "true")
	Middleware()(handler).ServeHTTP(nil, req)
	assert.True(t, carrySession)
	assert.True(t, skipRenewal)

	// Without no-session-renewal header.
	req.Header.Del(HeaderNoSessionRenewal)
	Middleware()(handler).ServeHTTP(nil, req)
	assert.False(t, skipRenewal)
}

func TestSessionCookieFlags(t *testing.T) {
	// Initialize in-memory config settings
	conf := map[string]any{
		common.ExtEndpoint: "https://harbor.test",
	}
	config.InitWithSettings(conf)

	// Mock http request
	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/users", nil)
	require.Nil(t, err)

	// Call middleware on request
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that Scheme has been updated to https
		assert.Equal(t, "https", r.URL.Scheme)
	})
	Middleware()(handler).ServeHTTP(httptest.NewRecorder(), req)

	// Switch to HTTP external endpoint
	conf = map[string]any{
		common.ExtEndpoint: "http://harbor.test",
	}
	config.InitWithSettings(conf)

	req2, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/users", nil)
	require.Nil(t, err)

	handler2 := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that Scheme remains http (empty or http)
		assert.Equal(t, "http", r.URL.Scheme)
	})
	Middleware()(handler2).ServeHTTP(httptest.NewRecorder(), req2)
}

func TestSessionCookieSecureSameSite(t *testing.T) {
	// 1. Test HTTPS endpoint (Secure=true)
	config.InitWithSettings(map[string]any{
		common.ExtEndpoint: "https://harbor.test",
	})

	// Override/initialize using the production start hook logic
	secure := true
	sameSite := http.SameSiteLaxMode

	conf := &beegosession.ManagerConfig{
		CookieName:      "sid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Secure:          secure,
		CookieLifeTime:  3600,
		CookieSameSite:  sameSite,
	}

	var err error
	web.GlobalSessions, err = beegosession.NewManager("memory", conf)
	require.Nil(t, err)

	req, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/users", nil)
	require.Nil(t, err)

	// Simulate session middleware scheme mutation
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Start session, which emits the cookie
		_, err := web.GlobalSessions.SessionStart(w, r)
		require.Nil(t, err)
	})

	rec := httptest.NewRecorder()
	Middleware()(handler).ServeHTTP(rec, req)

	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	assert.Equal(t, "sid", cookies[0].Name)
	assert.True(t, cookies[0].Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookies[0].SameSite)

	// 2. Test HTTP endpoint (Secure=false)
	config.InitWithSettings(map[string]any{
		common.ExtEndpoint: "http://harbor.test",
	})

	confHTTP := &beegosession.ManagerConfig{
		CookieName:      "sid",
		EnableSetCookie: true,
		Gclifetime:      3600,
		Secure:          false,
		CookieLifeTime:  3600,
		CookieSameSite:  sameSite,
	}

	web.GlobalSessions, err = beegosession.NewManager("memory", confHTTP)
	require.Nil(t, err)

	reqHTTP, err := http.NewRequest("GET", "http://127.0.0.1:8080/api/users", nil)
	require.Nil(t, err)

	recHTTP := httptest.NewRecorder()
	Middleware()(handler).ServeHTTP(recHTTP, reqHTTP)

	cookiesHTTP := recHTTP.Result().Cookies()
	require.Len(t, cookiesHTTP, 1)
	assert.Equal(t, "sid", cookiesHTTP[0].Name)
	assert.False(t, cookiesHTTP[0].Secure)
	assert.Equal(t, http.SameSiteLaxMode, cookiesHTTP[0].SameSite)
}
