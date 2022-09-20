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
	"path/filepath"
	"testing"

	"github.com/beego/beego/v2/server/web"
	beegosession "github.com/beego/beego/v2/server/web/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common/models"
)

func TestSession(t *testing.T) {
	var err error
	// initialize beego session manager
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
	err = store.Set(req.Context(), "user", user)
	require.Nil(t, err)

	session := &session{}
	ctx := session.Generate(req)
	assert.NotNil(t, ctx)
}
