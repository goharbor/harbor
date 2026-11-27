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

package authproxy

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/user"
)

type postAuthUserManager struct {
	user.Manager
	existing  *models.User
	onboarded bool
}

func (m *postAuthUserManager) GetByName(_ context.Context, _ string) (*models.User, error) {
	if m.existing == nil {
		return nil, errors.NotFoundError(nil)
	}
	return m.existing, nil
}

func (m *postAuthUserManager) Onboard(_ context.Context, u *models.User) error {
	m.onboarded = true
	u.UserID = 42
	return nil
}

func TestPostAuthenticateSetsIdentityAndPreservesProxyPrivileges(t *testing.T) {
	for _, tc := range []struct {
		name     string
		existing *models.User
	}{
		{name: "first login"},
		{name: "returning login", existing: &models.User{UserID: 42, SysAdminFlag: true}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			manager := &postAuthUserManager{existing: tc.existing}
			authenticator := &Auth{userMgr: manager}
			u := &models.User{Username: "proxy-user", AdminRoleInAuth: true, GroupIDs: []int{17, 23}}
			if err := authenticator.PostAuthenticate(context.Background(), u); err != nil {
				t.Fatal(err)
			}
			if u.UserID != 42 || u.SessionAuthMode != common.HTTPAuth {
				t.Fatalf("unexpected user ID or auth mode: %d, %q", u.UserID, u.SessionAuthMode)
			}
			if !u.AdminRoleInAuth || u.SysAdminFlag || len(u.GroupIDs) != 2 || u.GroupIDs[0] != 17 || u.GroupIDs[1] != 23 {
				t.Fatalf("proxy privileges changed: %+v", u)
			}
			if manager.onboarded != (tc.existing == nil) {
				t.Fatalf("unexpected onboarding: %v", manager.onboarded)
			}
		})
	}
}

func TestNewInitializesUserManager(t *testing.T) {
	if New().userMgr == nil {
		t.Fatal("auth proxy authenticator has no user manager")
	}
}
