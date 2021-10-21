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

package ldap

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	ldapCtl "github.com/goharbor/harbor/src/controller/ldap"
	ugCtl "github.com/goharbor/harbor/src/controller/usergroup"
	"github.com/goharbor/harbor/src/core/auth"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/ldap"
	"github.com/goharbor/harbor/src/pkg/ldap/model"
	"github.com/goharbor/harbor/src/pkg/user"
	ugModel "github.com/goharbor/harbor/src/pkg/usergroup/model"
)

// Auth implements AuthenticateHelper interface to authenticate against LDAP
type Auth struct {
	auth.DefaultAuthenticateHelper
	userMgr user.Manager
}

// Authenticate checks user's credential against LDAP based on basedn template and LDAP URL,
// if the check is successful a dummy record will be inserted into DB, such that this user can
// be associated to other entities in the system.
func (l *Auth) Authenticate(ctx context.Context, m models.AuthModel) (*models.User, error) {
	p := m.Principal
	if len(strings.TrimSpace(p)) == 0 {
		log.Debugf("LDAP authentication failed for empty user id.")
		return nil, auth.NewErrAuth("Empty user id")
	}
	ldapSession, err := ldapCtl.Ctl.Session(ctx)
	if err != nil {
		return nil, fmt.Errorf("can not load system ldap config: %v", err)
	}

	if err = ldapSession.Open(); err != nil {
		log.Warningf("ldap connection fail: %v", err)
		return nil, err
	}
	defer ldapSession.Close()

	ldapUsers, err := ldapSession.SearchUser(p)
	if err != nil {
		log.Warningf("ldap search fail: %v", err)
		return nil, err
	}
	if len(ldapUsers) == 0 {
		log.Warningf("Not found an entry.")
		return nil, auth.NewErrAuth("Not found an entry")
	} else if len(ldapUsers) != 1 {
		log.Warningf("Found more than one entry.")
		return nil, auth.NewErrAuth("Multiple entries found")
	}
	log.Debugf("Found ldap user: %v", ldapUsers[0].Username)

	dn := ldapUsers[0].DN
	if err = ldapSession.Bind(dn, m.Password); err != nil {
		log.Warningf("Failed to bind user, username: %s, dn: %s, error: %v", p, dn, err)
		return nil, auth.NewErrAuth(err.Error())
	}

	u := models.User{}
	u.Username = ldapUsers[0].Username
	u.Realname = ldapUsers[0].Realname
	u.Email = strings.TrimSpace(ldapUsers[0].Email)

	l.syncUserInfoFromDB(ctx, &u)
	l.attachLDAPGroup(ctx, ldapUsers, &u, ldapSession)

	return &u, nil
}

func (l *Auth) attachLDAPGroup(ctx context.Context, ldapUsers []model.User, u *models.User, sess *ldap.Session) {
	// Retrieve ldap related info in login to avoid too many traffic with LDAP server.
	// Get group admin dn
	groupCfg, err := config.LDAPGroupConf(ctx)
	if err != nil {
		log.Warningf("Failed to fetch ldap group configuration:%v", err)
		// most likely user doesn't configure user group info, it should not block user login
	}
	groupAdminDN := utils.TrimLower(groupCfg.AdminDN)
	// Attach user group
	for _, groupDN := range ldapUsers[0].GroupDNList {

		groupDN = utils.TrimLower(groupDN)
		// Attach LDAP group admin
		if len(groupAdminDN) > 0 && groupAdminDN == groupDN {
			u.AdminRoleInAuth = true
		}

	}
	// skip to attach group when ldap_group_search_filter is empty
	if len(groupCfg.Filter) == 0 {
		return
	}
	userGroups := make([]ugModel.UserGroup, 0)
	for _, dn := range ldapUsers[0].GroupDNList {
		lGroups, err := sess.SearchGroupByDN(dn)
		if err != nil {
			log.Warningf("Can not get the ldap group name with DN %v, error %v", dn, err)
			continue
		}
		if len(lGroups) == 0 {
			log.Warningf("Can not get the ldap group name with DN %v", dn)
			continue
		}
		userGroups = append(userGroups, ugModel.UserGroup{GroupName: lGroups[0].Name, LdapGroupDN: dn, GroupType: common.LDAPGroupType})
	}
	u.GroupIDs, err = ugCtl.Ctl.Populate(ctx, userGroups)
	if err != nil {
		log.Warningf("Failed to fetch ldap group configuration:%v", err)
	}
}

func (l *Auth) syncUserInfoFromDB(ctx context.Context, u *models.User) {
	// Retrieve SysAdminFlag from DB so that it transfer to session
	dbUser, err := l.userMgr.GetByName(ctx, u.Username)
	if errors.IsNotFoundErr(err) {
		return
	} else if err != nil {
		log.Errorf("failed to sync user info from DB error %v", err)
		return
	}
	u.SysAdminFlag = dbUser.SysAdminFlag
}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func (l *Auth) OnBoardUser(ctx context.Context, u *models.User) error {
	if u.Email == "" {
		if strings.Contains(u.Username, "@") {
			u.Email = u.Username
		}
	}
	u.Password = "12345678AbC" // Password is not kept in local db
	u.Comment = "from LDAP."   // Source is from LDAP

	return l.userMgr.Onboard(ctx, u)
}

// SearchUser -- Search user in ldap
func (l *Auth) SearchUser(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	s, err := ldapCtl.Ctl.Session(ctx)
	if err != nil {
		return nil, err
	}
	if err = s.Open(); err != nil {
		return nil, fmt.Errorf("failed to load system ldap config, %v", err)
	}
	defer s.Close()
	lUsers, err := s.SearchUser(username)
	if err != nil {
		return nil, fmt.Errorf("failed to search user in ldap")
	}

	if len(lUsers) > 1 {
		log.Warningf("There are more than one user found, return the first user")
	}
	if len(lUsers) > 0 {

		user.Username = strings.TrimSpace(lUsers[0].Username)
		user.Realname = strings.TrimSpace(lUsers[0].Realname)
		user.Email = strings.TrimSpace(lUsers[0].Email)

		log.Debugf("Found ldap user %v", user)
	} else {
		return nil, errors.NotFoundError(nil).WithMessage("no user found: %v", username)
	}

	return &user, nil
}

// SearchGroup -- Search group in ldap authenticator, groupKey is LDAP group DN.
func (l *Auth) SearchGroup(ctx context.Context, groupKey string) (*ugModel.UserGroup, error) {
	if _, err := goldap.ParseDN(groupKey); err != nil {
		return nil, auth.ErrInvalidLDAPGroupDN
	}
	s, err := ldapCtl.Ctl.Session(ctx)

	if err != nil {
		return nil, fmt.Errorf("can not load system ldap config: %v", err)
	}

	if err = s.Open(); err != nil {
		log.Warningf("ldap connection fail: %v", err)
		return nil, err
	}
	defer s.Close()
	userGroupList, err := s.SearchGroupByDN(groupKey)

	if err != nil {
		log.Warningf("ldap search group fail: %v", err)
		return nil, err
	}

	if len(userGroupList) == 0 {
		return nil, errors.NotFoundError(nil).WithMessage("failed to searh ldap group with groupDN:%v", groupKey)
	}
	userGroup := ugModel.UserGroup{
		GroupName:   userGroupList[0].Name,
		LdapGroupDN: userGroupList[0].Dn,
	}
	return &userGroup, nil
}

// OnBoardGroup -- Create Group in harbor DB, if altGroupName is not empty, take the altGroupName as groupName in harbor DB.
func (l *Auth) OnBoardGroup(ctx context.Context, u *ugModel.UserGroup, altGroupName string) error {
	if _, err := goldap.ParseDN(u.LdapGroupDN); err != nil {
		return auth.ErrInvalidLDAPGroupDN
	}
	if len(altGroupName) > 0 {
		u.GroupName = altGroupName
	}
	u.GroupType = common.LDAPGroupType
	// Check duplicate LDAP DN in usergroup, if usergroup exist, return error
	userGroupList, err := ugCtl.Ctl.List(ctx, q.New(q.KeyWords{"LdapGroupDN": u.LdapGroupDN}))
	if err != nil {
		return err
	}
	if len(userGroupList) > 0 {
		return auth.ErrDuplicateLDAPGroup
	}
	return ugCtl.Ctl.Ensure(ctx, u)
}

// PostAuthenticate -- If user exist in harbor DB, sync email address, if not exist, call OnBoardUser
func (l *Auth) PostAuthenticate(ctx context.Context, u *models.User) error {
	query := q.New(q.KeyWords{"Username": u.Username})
	n, err := l.userMgr.Count(ctx, query)
	if err != nil {
		return err
	}

	if n > 0 {
		dbUser, err := l.userMgr.GetByName(ctx, u.Username)
		if errors.IsNotFoundErr(err) {
			log.Debugf("User not found in DB:%v", u.Username)
			return nil
		} else if err != nil {
			return err
		}
		u.UserID = dbUser.UserID
		if dbUser.Email != u.Email {
			Re := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
			if !Re.MatchString(u.Email) {
				log.Debugf("Not a valid email address: %v, skip to sync", u.Email)
			} else {
				if err = l.userMgr.UpdateProfile(ctx, u, "Email"); err != nil {
					u.Email = dbUser.Email
					log.Errorf("failed to sync user email: %v", err)
				}
			}
		}

		return nil
	}

	err = auth.OnBoardUser(ctx, u)
	if err != nil {
		return err
	}
	if u.UserID <= 0 {
		return fmt.Errorf("cannot OnBoardUser %v", u)
	}
	return nil
}

func init() {
	auth.Register(common.LDAPAuth, &Auth{
		userMgr: user.New(),
	})
}
