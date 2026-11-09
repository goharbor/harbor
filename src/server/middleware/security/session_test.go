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
	"bytes"
	"encoding/gob"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/beego/beego/v2/server/web"
	beegosession "github.com/beego/beego/v2/server/web/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/lib"
	pkguser "github.com/goharbor/harbor/src/pkg/user"
	usertesting "github.com/goharbor/harbor/src/testing/pkg/user"
)

func TestSessionAuthModeSurvivesSerialization(t *testing.T) {
	// Beego's persistent session providers encode their values with gob.
	original := models.User{UserID: 1, SessionAuthMode: common.HTTPAuth}
	var data bytes.Buffer
	require.NoError(t, gob.NewEncoder(&data).Encode(original))
	var decoded models.User
	require.NoError(t, gob.NewDecoder(&data).Decode(&decoded))
	assert.Equal(t, original.SessionAuthMode, decoded.SessionAuthMode)
}

func TestSession(t *testing.T) {
	conf := &beegosession.ManagerConfig{
		CookieName:      web.BConfig.WebConfig.Session.SessionName,
		Gclifetime:      web.BConfig.WebConfig.Session.SessionGCMaxLifetime,
		ProviderConfig:  filepath.ToSlash(web.BConfig.WebConfig.Session.SessionProviderConfig),
		Secure:          web.BConfig.Listen.EnableHTTPS,
		EnableSetCookie: web.BConfig.WebConfig.Session.SessionAutoSetCookie,
		Domain:          web.BConfig.WebConfig.Session.SessionDomain,
		CookieLifeTime:  web.BConfig.WebConfig.Session.SessionCookieLifeTime,
	}
	previousSessions := web.GlobalSessions
	var err error
	web.GlobalSessions, err = beegosession.NewManager("memory", conf)
	require.NoError(t, err)
	t.Cleanup(func() { web.GlobalSessions = previousSessions })

	tests := []struct {
		name        string
		cached      any
		current     *models.User
		lookupError error
		expectUser  bool
		expectAdmin bool
	}{
		{name: "missing session user"},
		{name: "invalid session user", cached: "invalid"},
		{
			name:       "admin flag refreshed",
			cached:     models.User{UserID: 1, Username: "old", SysAdminFlag: true},
			current:    &models.User{UserID: 1, Username: "updated", Password: "hash", Salt: "salt"},
			expectUser: true,
		},
		{
			name:        "admin flag granted",
			cached:      models.User{UserID: 1},
			current:     &models.User{UserID: 1, SysAdminFlag: true},
			expectUser:  true,
			expectAdmin: true,
		},
		{
			name:       "external roles preserved",
			cached:     models.User{UserID: 1, AdminRoleInAuth: true, GroupIDs: []int{3, 5}},
			current:    &models.User{UserID: 1, Username: "ldap"},
			expectUser: true, expectAdmin: true,
		},
		{name: "user deleted", cached: models.User{UserID: 1}, current: &models.User{UserID: 1, Deleted: true}},
		{name: "user missing", cached: models.User{UserID: 1}, lookupError: errors.New("not found")},
		{name: "lookup failed", cached: models.User{UserID: 1}, lookupError: errors.New("database unavailable")},
		{name: "empty lookup result", cached: models.User{UserID: 1}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			previousUserMgr := pkguser.Mgr
			mgr := usertesting.NewManager(t)
			pkguser.Mgr = mgr
			t.Cleanup(func() { pkguser.Mgr = previousUserMgr })
			if _, ok := tc.cached.(models.User); ok {
				mgr.On("Get", mock.Anything, 1).Return(tc.current, tc.lookupError).Once()
			}

			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
			store, err := web.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
			require.NoError(t, err)
			if tc.cached != nil {
				require.NoError(t, store.Set(req.Context(), "user", tc.cached))
			}
			ctx := (&session{}).Generate(req)
			if !tc.expectUser {
				assert.Nil(t, ctx)
				return
			}
			require.NotNil(t, ctx)
			assert.Equal(t, tc.expectAdmin, ctx.IsSysAdmin())
			localCtx, ok := ctx.(*local.SecurityContext)
			require.True(t, ok)
			assert.Equal(t, tc.current.Username, localCtx.GetUsername())
			assert.Empty(t, localCtx.User().Password)
			assert.Empty(t, localCtx.User().Salt)
			cached := tc.cached.(models.User)
			assert.Equal(t, cached.GroupIDs, localCtx.User().GroupIDs)
		})
	}
}

// TestSessionAuthOrigin checks that a database administrator flag cannot
// grant privileges to an externally authenticated user with the same account.
func TestSessionAuthOrigin(t *testing.T) {
	conf := &beegosession.ManagerConfig{
		CookieName:      web.BConfig.WebConfig.Session.SessionName,
		Gclifetime:      web.BConfig.WebConfig.Session.SessionGCMaxLifetime,
		ProviderConfig:  filepath.ToSlash(web.BConfig.WebConfig.Session.SessionProviderConfig),
		EnableSetCookie: web.BConfig.WebConfig.Session.SessionAutoSetCookie,
	}
	previousSessions := web.GlobalSessions
	var err error
	web.GlobalSessions, err = beegosession.NewManager("memory", conf)
	require.NoError(t, err)
	t.Cleanup(func() { web.GlobalSessions = previousSessions })

	tests := []struct {
		name        string
		authMode    string
		cached      models.User
		current     *models.User
		expectUser  bool
		expectAdmin bool
	}{
		{
			name:       "proxy user cannot inherit database admin",
			cached:     models.User{UserID: 42, Username: "admin", SessionAuthMode: common.HTTPAuth, GroupIDs: []int{3, 5}},
			current:    &models.User{UserID: 42, Username: "admin", SysAdminFlag: true},
			expectUser: true,
		},
		{
			name:       "proxy admin retains provider privilege",
			cached:     models.User{UserID: 42, Username: "proxy-admin", SessionAuthMode: common.HTTPAuth, AdminRoleInAuth: true},
			current:    &models.User{UserID: 42, Username: "proxy-admin"},
			expectUser: true, expectAdmin: true,
		},
		{
			name:       "local administrator retains database privilege",
			cached:     models.User{UserID: 1, Username: "admin", SessionAuthMode: common.DBAuth},
			current:    &models.User{UserID: 1, Username: "admin", SysAdminFlag: true},
			expectUser: true, expectAdmin: true,
		},
		{
			name:       "local administrator database privilege revoked",
			cached:     models.User{UserID: 1, Username: "admin", SessionAuthMode: common.DBAuth, SysAdminFlag: true},
			current:    &models.User{UserID: 1, Username: "admin"},
			expectUser: true,
		},
		{
			name:       "legacy DB-auth session without origin remains valid",
			authMode:   common.DBAuth,
			cached:     models.User{UserID: 1, Username: "admin"},
			current:    &models.User{UserID: 1, Username: "admin", SysAdminFlag: true},
			expectUser: true, expectAdmin: true,
		},
		{
			name:   "legacy session without origin rejected",
			cached: models.User{UserID: 42, Username: "admin"},
		},
		{
			name:   "legacy proxy session without ID rejected",
			cached: models.User{Username: "proxy-user"},
		},
		{
			name:   "proxy session with origin but without ID rejected",
			cached: models.User{Username: "proxy-user", SessionAuthMode: common.HTTPAuth},
		},
		{
			name:   "local session without ID rejected",
			cached: models.User{Username: "admin", SessionAuthMode: common.DBAuth},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			previousUserMgr := pkguser.Mgr
			mgr := usertesting.NewManager(t)
			pkguser.Mgr = mgr
			t.Cleanup(func() { pkguser.Mgr = previousUserMgr })
			if tc.current != nil {
				mgr.On("Get", mock.Anything, tc.cached.UserID).Return(tc.current, nil).Once()
			}
			req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/api/projects/", nil)
			mode := tc.authMode
			if mode == "" {
				mode = common.HTTPAuth
			}
			req = req.WithContext(lib.WithAuthMode(req.Context(), mode))
			store, err := web.GlobalSessions.SessionStart(httptest.NewRecorder(), req)
			require.NoError(t, err)
			require.NoError(t, store.Set(req.Context(), "user", tc.cached))
			ctx := (&session{}).Generate(req)
			if !tc.expectUser {
				assert.Nil(t, ctx)
				return
			}
			require.NotNil(t, ctx)
			assert.Equal(t, tc.expectAdmin, ctx.IsSysAdmin())
			localCtx, ok := ctx.(*local.SecurityContext)
			require.True(t, ok)
			assert.Equal(t, tc.cached.GroupIDs, localCtx.User().GroupIDs)
		})
	}
}
