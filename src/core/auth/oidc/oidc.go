package oidc

import (
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao/group"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/auth"
)

// Auth of OIDC mode only implements the funcs for onboarding group
type Auth struct {
	auth.DefaultAuthenticateHelper
}

// SearchGroup is skipped in OIDC mode, so it makes sure any group will be onboarded.
func (a *Auth) SearchGroup(groupKey string) (*models.UserGroup, error) {
	return &models.UserGroup{
		GroupName: groupKey,
		GroupType: common.OIDCGroupType,
	}, nil
}

// OnBoardGroup create user group entity in Harbor DB, altGroupName is not used.
func (a *Auth) OnBoardGroup(u *models.UserGroup, altGroupName string) error {
	// if group name provided, on board the user group
	if len(u.GroupName) == 0 || u.GroupType != common.OIDCGroupType {
		return fmt.Errorf("invalid input group for OIDC mode: %v", *u)
	}
	return group.OnBoardUserGroup(u)
}

func init() {
	auth.Register(common.OIDCAuth, &Auth{})
}
