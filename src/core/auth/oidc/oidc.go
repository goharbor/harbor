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
func (a *Auth) SearchGroup(ctx context.Context, groupKey string) (*model.UserGroup, error) {
	return &model.UserGroup{
		GroupName: groupKey,
		GroupType: common.OIDCGroupType,
	}, nil
}

// OnBoardGroup create user group entity in Harbor DB, altGroupName is not used.
func (a *Auth) OnBoardGroup(ctx context.Context, u *model.UserGroup, altGroupName string) error {
	// if group name provided, on board the user group
	if len(u.GroupName) == 0 || u.GroupType != common.OIDCGroupType {
		return fmt.Errorf("invalid input group for OIDC mode: %v", *u)
	}
	return usergroup.Mgr.Onboard(ctx, u)
}

func init() {
	auth.Register(common.OIDCAuth, &Auth{})
}
