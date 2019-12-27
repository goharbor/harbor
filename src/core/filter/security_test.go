// Copyright 2018 Project Harbor Authors
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

package filter

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/common/utils/oidc"
	"github.com/stretchr/testify/require"

	"github.com/astaxie/beego"
	beegoctx "github.com/astaxie/beego/context"
	"github.com/astaxie/beego/session"
	config2 "github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	commonsecret "github.com/goharbor/harbor/src/common/secret"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/common/utils/test"
	_ "github.com/goharbor/harbor/src/core/auth/authproxy"
	_ "github.com/goharbor/harbor/src/core/auth/db"
	_ "github.com/goharbor/harbor/src/core/auth/ldap"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/promgr"
	driver_local "github.com/goharbor/harbor/src/core/promgr/pmsdriver/local"
	"github.com/stretchr/testify/assert"

	"github.com/goharbor/harbor/src/common"
	fiter_test "github.com/goharbor/harbor/src/core/filter/test"
)

func TestMain(m *testing.M) {
	// initialize beego session manager
	conf := &session.ManagerConfig{
		CookieName:      beego.BConfig.WebConfig.Session.SessionName,
		Gclifetime:      beego.BConfig.WebConfig.Session.SessionGCMaxLifetime,
		ProviderConfig:  filepath.ToSlash(beego.BConfig.WebConfig.Session.SessionProviderConfig),
		Secure:          beego.BConfig.Listen.EnableHTTPS,
		EnableSetCookie: beego.BConfig.WebConfig.Session.SessionAutoSetCookie,
		Domain:          beego.BConfig.WebConfig.Session.SessionDomain,
		CookieLifeTime:  beego.BConfig.WebConfig.Session.SessionCookieLifeTime,
	}

	var err error
	beego.GlobalSessions, err = session.NewManager("memory", conf)
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}
	config.Init()
	test.InitDatabaseFromEnv()

	config.Upload(test.GetUnitTestConfig())
	Init()

	os.Exit(m.Run())
}

func TestSecurityFilter(t *testing.T) {
	// nil request
	ctx, err := newContext(nil)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}
	SecurityFilter(ctx)
	assert.Nil(t, securityContext(ctx))
	assert.Nil(t, projectManager(ctx))

	// the pattern of request needs security check
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}

	ctx, err = newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}
	SecurityFilter(ctx)
	assert.NotNil(t, securityContext(ctx))
	assert.NotNil(t, projectManager(ctx))
}

func TestConfigCtxModifier(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	conf := map[string]interface{}{
		common.AUTHMode:         common.OIDCAuth,
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://accounts.google.com",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile, offline_access",
		common.OIDCGroupsClaim:  "groups",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "secret",
		common.ExtEndpoint:      "https://harbor.test",
	}
	config.InitWithSettings(conf)
	ctx, err := newContext(req)
	m := &configCtxModifier{}
	f := m.Modify(ctx)
	assert.False(t, f)
	assert.Equal(t, common.OIDCAuth, req.Context().Value(AuthModeKey).(string))
}

func TestSecretReqCtxModifier(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	commonsecret.AddToRequest(req, "secret")
	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}

	modifier := &secretReqCtxModifier{}
	modified := modifier.Modify(ctx)
	assert.True(t, modified)
	assert.IsType(t, &secret.SecurityContext{},
		securityContext(ctx))
	assert.NotNil(t, projectManager(ctx))
}

func TestOIDCCliReqCtxModifier(t *testing.T) {
	conf := map[string]interface{}{
		common.AUTHMode:         common.OIDCAuth,
		common.OIDCName:         "test",
		common.OIDCEndpoint:     "https://accounts.google.com",
		common.OIDCVerifyCert:   "true",
		common.OIDCScope:        "openid, profile, offline_access",
		common.OIDCCLientID:     "client",
		common.OIDCClientSecret: "secret",
		common.ExtEndpoint:      "https://harbor.test",
	}

	kp := &config2.PresetKeyProvider{Key: "naa4JtarA1Zsc3uY"}
	config.InitWithSettings(conf, kp)

	modifier := &oidcCliReqCtxModifier{}
	req1, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	ctx1, err := newContext(req1)
	require.Nil(t, err)
	assert.False(t, modifier.Modify(ctx1))
	req2, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/service/token", nil)
	require.Nil(t, err)
	addToReqContext(req2, AuthModeKey, common.OIDCAuth)
	ctx2, err := newContext(req2)
	require.Nil(t, err)
	assert.False(t, modifier.Modify(ctx2))
	username := "oidcModiferTester"
	password := "oidcSecret"
	u := &models.User{
		Username: username,
		Email:    "testtest@test.org",
		Password: "12345678",
	}
	id, err := dao.Register(*u)
	require.Nil(t, err)
	oidc.SetHardcodeVerifierForTest(password)
	req3, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/service/token", nil)
	require.Nil(t, err)
	req3.SetBasicAuth(username, password)
	addToReqContext(req3, AuthModeKey, common.OIDCAuth)
	ctx3, err := newContext(req3)
	assert.True(t, modifier.Modify(ctx3))
	o := dao.GetOrmer()
	_, err = o.Delete(&models.User{UserID: int(id)})
	assert.Nil(t, err)
}

func TestIdTokenReqCtxModifier(t *testing.T) {
	bc := context.Background()
	it := &idTokenReqCtxModifier{}
	r1, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/chartrepo/", nil)
	require.Nil(t, err)
	req1 := r1.WithContext(context.WithValue(bc, AuthModeKey, common.DBAuth))
	ctx1, err := newContext(req1)
	require.Nil(t, err)
	assert.False(t, it.Modify(ctx1))

	req2 := r1.WithContext(context.WithValue(bc, AuthModeKey, common.OIDCAuth))
	ctx2, err := newContext(req2)
	require.Nil(t, err)
	assert.False(t, it.Modify(ctx2))

	r2, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	req3 := r2.WithContext(context.WithValue(bc, AuthModeKey, common.OIDCAuth))
	ctx3, err := newContext(req3)
	require.Nil(t, err)
	assert.False(t, it.Modify(ctx3))
}

func TestRobotReqCtxModifier(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	req.SetBasicAuth("robot$test1", "Harbor12345")
	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}

	modifier := &robotAuthReqCtxModifier{}
	modified := modifier.Modify(ctx)
	assert.False(t, modified)
}

func TestAuthProxyReqCtxModifier(t *testing.T) {

	server, err := fiter_test.NewAuthProxyTestServer()
	assert.Nil(t, err)
	defer server.Close()

	c := map[string]interface{}{
		common.HTTPAuthProxySkipSearch:          "true",
		common.HTTPAuthProxyVerifyCert:          "false",
		common.HTTPAuthProxyEndpoint:            "https://auth.proxy/suffix",
		common.HTTPAuthProxyTokenReviewEndpoint: server.URL,
		common.AUTHMode:                         common.HTTPAuth,
	}

	config.Upload(c)
	v, e := config.HTTPAuthProxySetting()
	assert.Nil(t, e)
	assert.Equal(t, *v, models.HTTPAuthProxy{
		Endpoint:            "https://auth.proxy/suffix",
		SkipSearch:          true,
		VerifyCert:          false,
		TokenReviewEndpoint: server.URL,
	})

	// No onboard
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/service/token", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	req.SetBasicAuth("tokenreview$administrator@vsphere.local", "reviEwt0k3n")
	addToReqContext(req, AuthModeKey, common.HTTPAuth)
	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	modifier := &authProxyReqCtxModifier{}
	modified := modifier.Modify(ctx)
	assert.True(t, modified)

}

func TestBasicAuthReqCtxModifier(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	req.SetBasicAuth("admin", "Harbor12345")

	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}

	modifier := &basicAuthReqCtxModifier{}
	modified := modifier.Modify(ctx)
	assert.True(t, modified)

	sc := securityContext(ctx)
	assert.IsType(t, &local.SecurityContext{}, sc)
	s := sc.(security.Context)
	assert.Equal(t, "admin", s.GetUsername())
	assert.NotNil(t, projectManager(ctx))
}

func TestSessionReqCtxModifier(t *testing.T) {
	user := models.User{
		Username:     "admin",
		UserID:       1,
		Email:        "admin@example.com",
		SysAdminFlag: true,
	}
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	store, err := beego.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	if err != nil {
		t.Fatalf("failed to create session store: %v", err)
	}
	if err = store.Set("user", user); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}

	addSessionIDToCookie(req, store.SessionID())
	addToReqContext(req, AuthModeKey, common.DBAuth)
	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to create context: %v", err)
	}

	modifier := &sessionReqCtxModifier{}
	modified := modifier.Modify(ctx)

	assert.True(t, modified)
	sc := securityContext(ctx)
	assert.IsType(t, &local.SecurityContext{}, sc)
	s := sc.(security.Context)
	assert.Equal(t, "admin", s.GetUsername())
	assert.True(t, s.IsSysAdmin())
	assert.NotNil(t, projectManager(ctx))

}

func TestSessionReqCtxModifierFailed(t *testing.T) {
	user := "admin"
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	store, err := beego.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	if err != nil {
		t.Fatalf("failed to create session store: %v", err)
	}
	if err = store.Set("user", user); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}

	req, err = http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	addSessionIDToCookie(req, store.SessionID())
	addToReqContext(req, AuthModeKey, common.DBAuth)
	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}
	modifier := &sessionReqCtxModifier{}
	modified := modifier.Modify(ctx)

	assert.False(t, modified)

}

// TODO add test case
func TestTokenReqCtxModifier(t *testing.T) {

}

func TestUnauthorizedReqCtxModifier(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}

	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}

	modifier := &unauthorizedReqCtxModifier{}
	modified := modifier.Modify(ctx)
	assert.True(t, modified)

	sc := securityContext(ctx)
	assert.NotNil(t, sc)
	s := sc.(security.Context)
	assert.False(t, s.IsAuthenticated())
	assert.NotNil(t, projectManager(ctx))
}

func newContext(req *http.Request) (*beegoctx.Context, error) {
	var err error
	ctx := beegoctx.NewContext()
	ctx.Reset(httptest.NewRecorder(), req)
	if req != nil {
		ctx.Input.CruSession, err = beego.GlobalSessions.SessionStart(ctx.ResponseWriter, req)
	}
	return ctx, err
}

func addSessionIDToCookie(req *http.Request, sessionID string) {
	cookie := &http.Cookie{
		Name:     beego.BConfig.WebConfig.Session.SessionName,
		Value:    url.QueryEscape(sessionID),
		Path:     "/",
		HttpOnly: true,
		Secure:   beego.BConfig.Listen.EnableHTTPS,
		Domain:   beego.BConfig.WebConfig.Session.SessionDomain,
	}
	cookie.MaxAge = beego.BConfig.WebConfig.Session.SessionCookieLifeTime
	cookie.Expires = time.Now().Add(
		time.Duration(
			beego.BConfig.WebConfig.Session.SessionCookieLifeTime) * time.Second)
	req.AddCookie(cookie)
}

func securityContext(ctx *beegoctx.Context) interface{} {
	c, err := GetSecurityContext(ctx.Request)
	if err != nil {
		log.Printf("failed to get security context: %v \n", err)
		return nil
	}
	return c
}

func projectManager(ctx *beegoctx.Context) interface{} {
	if ctx.Request == nil {
		return nil
	}
	return ctx.Request.Context().Value(PmKey)
}

func TestGetSecurityContext(t *testing.T) {
	// nil request
	ctx, err := GetSecurityContext(nil)
	assert.NotNil(t, err)

	// the request contains no security context
	req, err := http.NewRequest("", "", nil)
	assert.Nil(t, err)
	ctx, err = GetSecurityContext(req)
	assert.NotNil(t, err)

	// the request contains a variable which is not the correct type
	req, err = http.NewRequest("", "", nil)
	assert.Nil(t, err)
	req = req.WithContext(context.WithValue(req.Context(),
		SecurCtxKey, "test"))
	ctx, err = GetSecurityContext(req)
	assert.NotNil(t, err)

	// the request contains a correct variable
	req, err = http.NewRequest("", "", nil)
	assert.Nil(t, err)
	req = req.WithContext(context.WithValue(req.Context(),
		SecurCtxKey, local.NewSecurityContext(nil, nil)))
	ctx, err = GetSecurityContext(req)
	assert.Nil(t, err)
	_, ok := ctx.(security.Context)
	assert.True(t, ok)
}

func TestGetProjectManager(t *testing.T) {
	// nil request
	pm, err := GetProjectManager(nil)
	assert.NotNil(t, err)

	// the request contains no project manager
	req, err := http.NewRequest("", "", nil)
	assert.Nil(t, err)
	pm, err = GetProjectManager(req)
	assert.NotNil(t, err)

	// the request contains a variable which is not the correct type
	req, err = http.NewRequest("", "", nil)
	assert.Nil(t, err)
	req = req.WithContext(context.WithValue(req.Context(),
		PmKey, "test"))
	pm, err = GetProjectManager(req)
	assert.NotNil(t, err)

	// the request contains a correct variable
	req, err = http.NewRequest("", "", nil)
	assert.Nil(t, err)
	req = req.WithContext(context.WithValue(req.Context(),
		PmKey, promgr.NewDefaultProjectManager(driver_local.NewDriver(), true)))
	pm, err = GetProjectManager(req)
	assert.Nil(t, err)
	_, ok := pm.(promgr.ProjectManager)
	assert.True(t, ok)
}
