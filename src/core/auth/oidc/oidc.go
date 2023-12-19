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

package oidc

import (
	"context"
	"fmt"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/pkg/usergroup"
	"github.com/goharbor/harbor/src/pkg/usergroup/model"
)

// Auth of OIDC mode only implements the funcs for onboarding group
type Auth struct {
	auth.DefaultAuthenticateHelper
}

// SearchGroup is skipped in OIDC mode, so it makes sure any group will be onboarded.
func (a *Auth) SearchGroup(_ context.Context, groupKey string) (*model.UserGroup, error) {
	return &model.UserGroup{
		GroupName: groupKey,
		GroupType: common.OIDCGroupType,
	}, nil
}

// OnBoardGroup create user group entity in Harbor DB, altGroupName is not used.
func (a *Auth) OnBoardGroup(ctx context.Context, u *model.UserGroup, _ string) error {
	// if group name provided, on board the user group
	if len(u.GroupName) == 0 || u.GroupType != common.OIDCGroupType {
		return fmt.Errorf("invalid input group for OIDC mode: %v", *u)
	}
	return usergroup.Mgr.Onboard(ctx, u)
}

func init() {
	auth.Register(common.OIDCAuth, &Auth{})
}
