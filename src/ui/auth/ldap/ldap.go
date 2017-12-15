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

package ldap

import (
	"fmt"
	"strings"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	ldapUtils "github.com/vmware/harbor/src/common/utils/ldap"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
)

// Auth implements AuthenticateHelper interface to authenticate against LDAP
type Auth struct{}

const metaChars = "&|!=~*<>()"

// Authenticate checks user's credential against LDAP based on basedn template and LDAP URL,
// if the check is successful a dummy record will be inserted into DB, such that this user can
// be associated to other entities in the system.
func (l *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	p := m.Principal
	if len(strings.TrimSpace(p)) == 0 {
		log.Debugf("LDAP authentication failed for empty user id.")
		return nil, nil
	}
	for _, c := range metaChars {
		if strings.ContainsRune(p, c) {
			return nil, fmt.Errorf("the principal contains meta char: %q", c)
		}
	}

	ldapSession, err := ldapUtils.LoadSystemLdapConfig()

	if err != nil {
		return nil, fmt.Errorf("can not load system ldap config: %v", err)
	}

	if err = ldapSession.Open(); err != nil {
		log.Warningf("ldap connection fail: %v", err)
		return nil, nil
	}
	defer ldapSession.Close()

	ldapUsers, err := ldapSession.SearchUser(p)

	if err != nil {
		log.Warningf("ldap search fail: %v", err)
		return nil, nil
	}

	if len(ldapUsers) == 0 {
		log.Warningf("Not found an entry.")
		return nil, nil
	} else if len(ldapUsers) != 1 {
		log.Warningf("Found more than one entry.")
		return nil, nil
	}

	u := models.User{}
	u.Username = ldapUsers[0].Username
	u.Email = ldapUsers[0].Email
	u.Realname = ldapUsers[0].Realname

	dn := ldapUsers[0].DN

	log.Debugf("username: %s, dn: %s", u.Username, dn)
	if err = ldapSession.Bind(dn, m.Password); err != nil {
		log.Warningf("Failed to bind user, username: %s, dn: %s, error: %v", u.Username, dn, err)
		return nil, nil
	}
	exist, err := dao.UserExists(u, "username")
	if err != nil {
		return nil, err
	}

	if exist {
		currentUser, err := dao.GetUser(u)
		if err != nil {
			return nil, err
		}
		u.UserID = currentUser.UserID
		u.HasAdminRole = currentUser.HasAdminRole
	} else {
		var user models.User
		user.Username = ldapUsers[0].Username
		user.Email = ldapUsers[0].Email
		user.Realname = ldapUsers[0].Realname

		err = auth.OnBoardUser(&user)
		if err != nil || user.UserID <= 0 {
			log.Errorf("Can't import user %s, error: %v", ldapUsers[0].Username, err)
			return nil, fmt.Errorf("can't import user %s, error: %v", ldapUsers[0].Username, err)
		}
		u.UserID = user.UserID
	}

	return &u, nil

}

// OnBoardUser will check if a user exists in user table, if not insert the user and
// put the id in the pointer of user model, if it does exist, return the user's profile.
func (l *Auth) OnBoardUser(u *models.User) error {
	if u.Email == "" {
		if strings.Contains(u.Username, "@") {
			u.Email = u.Username
		} else {
			u.Email = u.Username + "@placeholder.com"
		}
	}
	u.Password = "12345678AbC" //Password is not kept in local db
	u.Comment = "from LDAP."   //Source is from LDAP

	return dao.OnBoardUser(u)
}

//SearchUser -- Search user in ldap
func (l *Auth) SearchUser(username string) (*models.User, error) {
	var user models.User
	ldapSession, err := ldapUtils.LoadSystemLdapConfig()
	if err = ldapSession.Open(); err != nil {
		return nil, fmt.Errorf("Failed to load system ldap config, %v", err)
	}

	ldapUsers, err := ldapSession.SearchUser(username)
	if err != nil {
		return nil, fmt.Errorf("Failed to search user in ldap")
	}

	if len(ldapUsers) > 1 {
		log.Warningf("There are more than one user found, return the first user")
	}
	if len(ldapUsers) > 0 {

		user.Username = strings.TrimSpace(ldapUsers[0].Username)
		user.Realname = strings.TrimSpace(ldapUsers[0].Realname)

		log.Debugf("Found ldap user %v", user)
	} else {
		return nil, fmt.Errorf("No user found, %v", username)
	}

	return &user, nil
}

func init() {
	auth.Register("ldap_auth", &Auth{})
}
