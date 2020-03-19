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
	"github.com/astaxie/beego"
	beegosession "github.com/astaxie/beego/session"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/internal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSession(t *testing.T) {
	carrySession := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		carrySession = internal.GetCarrySession(r.Context())
	})
	// no session
	req, err := http.NewRequest("POST", "http://127.0.0.1:8080/api/users", nil)
	require.Nil(t, err)
	Middleware()(handler).ServeHTTP(nil, req)
	assert.False(t, carrySession)

	// contains session
	beego.BConfig.WebConfig.Session.SessionName = config.SessionCookieName
	conf := &beegosession.ManagerConfig{
		CookieName:      beego.BConfig.WebConfig.Session.SessionName,
		Gclifetime:      beego.BConfig.WebConfig.Session.SessionGCMaxLifetime,
		ProviderConfig:  filepath.ToSlash(beego.BConfig.WebConfig.Session.SessionProviderConfig),
		Secure:          beego.BConfig.Listen.EnableHTTPS,
		EnableSetCookie: beego.BConfig.WebConfig.Session.SessionAutoSetCookie,
		Domain:          beego.BConfig.WebConfig.Session.SessionDomain,
		CookieLifeTime:  beego.BConfig.WebConfig.Session.SessionCookieLifeTime,
	}
	beego.GlobalSessions, err = beegosession.NewManager("memory", conf)
	require.Nil(t, err)
	_, err = beego.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	require.Nil(t, err)
	Middleware()(handler).ServeHTTP(nil, req)
	assert.True(t, carrySession)
}
