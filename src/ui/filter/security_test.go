// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
	"github.com/astaxie/beego/session"
	"github.com/stretchr/testify/assert"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/security"
	"github.com/vmware/harbor/src/common/security/rbac"
	"github.com/vmware/harbor/src/common/security/secret"
	_ "github.com/vmware/harbor/src/ui/auth/db"
	_ "github.com/vmware/harbor/src/ui/auth/ldap"
	"github.com/vmware/harbor/src/ui/config"
)

func TestMain(m *testing.M) {
	// initialize beego session manager
	conf := map[string]interface{}{
		"cookieName":      beego.BConfig.WebConfig.Session.SessionName,
		"gclifetime":      beego.BConfig.WebConfig.Session.SessionGCMaxLifetime,
		"providerConfig":  filepath.ToSlash(beego.BConfig.WebConfig.Session.SessionProviderConfig),
		"secure":          beego.BConfig.Listen.EnableHTTPS,
		"enableSetCookie": beego.BConfig.WebConfig.Session.SessionAutoSetCookie,
		"domain":          beego.BConfig.WebConfig.Session.SessionDomain,
		"cookieLifeTime":  beego.BConfig.WebConfig.Session.SessionCookieLifeTime,
	}
	confBytes, err := json.Marshal(conf)
	if err != nil {
		log.Fatalf("failed to marshal session conf: %v", err)
	}

	beego.GlobalSessions, err = session.NewManager("memory", string(confBytes))
	if err != nil {
		log.Fatalf("failed to create session manager: %v", err)
	}

	if err := config.Init(); err != nil {
		log.Fatalf("failed to initialize configurations: %v", err)
	}
	database, err := config.Database()
	if err != nil {
		log.Fatalf("failed to get database configurations: %v", err)
	}
	if err = dao.InitDatabase(database); err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	os.Exit(m.Run())
}

func TestSecurityFilter(t *testing.T) {
	// nil request
	ctx, err := newContext(nil)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}
	SecurityFilter(ctx)
	assert.Nil(t, securityContext(ctx))
	assert.Nil(t, projectManager(ctx))

	// the pattern of request does not need security check
	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/static/index.html", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}

	ctx, err = newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}
	SecurityFilter(ctx)
	assert.Nil(t, securityContext(ctx))
	assert.Nil(t, projectManager(ctx))

	// the pattern of request needs security check
	req, err = http.NewRequest(http.MethodGet,
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

func TestFillContext(t *testing.T) {
	// secret
	req, err := http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	req.AddCookie(&http.Cookie{
		Name:  "secret",
		Value: "secret",
	})

	ctx, err := newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}

	fillContext(ctx)
	assert.IsType(t, &secret.SecurityContext{},
		securityContext(ctx))
	assert.NotNil(t, projectManager(ctx))

	// session
	req, err = http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	store, err := beego.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	if err != nil {
		t.Fatalf("failed to create session store: %v", err)
	}
	if err = store.Set("username", "admin"); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}
	if err = store.Set("isSysAdmin", true); err != nil {
		t.Fatalf("failed to set session: %v", err)
	}

	req, err = http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	addSessionIDToCookie(req, store.SessionID())

	ctx, err = newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}
	fillContext(ctx)
	sc := securityContext(ctx)
	assert.IsType(t, &rbac.SecurityContext{}, sc)
	s := sc.(security.Context)
	assert.Equal(t, "admin", s.GetUsername())
	assert.True(t, s.IsSysAdmin())
	assert.NotNil(t, projectManager(ctx))

	// basic auth
	req, err = http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}
	req.SetBasicAuth("admin", "Harbor12345")

	ctx, err = newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}
	fillContext(ctx)
	sc = securityContext(ctx)
	assert.IsType(t, &rbac.SecurityContext{}, sc)
	s = sc.(security.Context)
	assert.Equal(t, "admin", s.GetUsername())
	assert.NotNil(t, projectManager(ctx))

	// no credential
	req, err = http.NewRequest(http.MethodGet,
		"http://127.0.0.1/api/projects/", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", req)
	}

	ctx, err = newContext(req)
	if err != nil {
		t.Fatalf("failed to crate context: %v", err)
	}
	fillContext(ctx)
	sc = securityContext(ctx)
	assert.IsType(t, &rbac.SecurityContext{}, sc)
	s = sc.(security.Context)
	assert.False(t, s.IsAuthenticated())
	assert.NotNil(t, projectManager(ctx))
}

func newContext(req *http.Request) (*context.Context, error) {
	var err error
	ctx := context.NewContext()
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

func securityContext(ctx *context.Context) interface{} {
	return ctx.Input.Data()[HarborSecurityContext]
}

func projectManager(ctx *context.Context) interface{} {
	return ctx.Input.Data()[HarborProjectManager]
}
