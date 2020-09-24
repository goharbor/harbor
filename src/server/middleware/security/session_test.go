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
	"fmt"

	"github.com/astaxie/beego"
	beegosession "github.com/astaxie/beego/session"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

func TestSession(t *testing.T) {
	var err error
	// initialize beego session manager
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

	user := &models.User{
		Username:     "session_test",
		Email:        "session_test@example.com",
		SysAdminFlag: true,
	}
	err = dao.OnBoardUser(user)
	require.Nil(t, err)
	defer func(id int) {
		sql := fmt.Sprintf("DELETE FROM harbor_user WHERE user_id=%d", id)
		dao.ExecuteBatchSQL([]string{sql})
	}(user.UserID)

	req, err := http.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
	require.Nil(t, err)
	store, err := beego.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
	require.Nil(t, err)
	err = store.Set(UserIDSessionKey, 999)
	require.Nil(t, err)

	session := &session{}
	ctx := session.Generate(req)
	assert.Nil(t, ctx)
	err = store.Set(UserIDSessionKey, user.UserID)
	require.Nil(t, err)
	ctx = session.Generate(req)
	assert.NotNil(t, ctx)

}
