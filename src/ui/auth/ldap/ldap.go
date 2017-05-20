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

// Auth implements Authenticator interface to authenticate against LDAP
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

	ldapConfs, err := ldapUtils.GetSystemLdapConf()

	if err != nil {
		return nil, fmt.Errorf("can't load system configuration: %v", err)
	}

	ldapConfs, err = ldapUtils.ValidateLdapConf(ldapConfs)

	if err != nil {
		return nil, fmt.Errorf("invalid ldap request: %v", err)
	}

	ldapConfs.LdapFilter = ldapUtils.MakeFilter(p, ldapConfs.LdapFilter, ldapConfs.LdapUID)

	ldapUsers, err := ldapUtils.SearchUser(ldapConfs)

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
	if err := ldapUtils.Bind(ldapConfs, dn, m.Password); err != nil {
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
	} else {
		userID, err := ldapUtils.ImportUser(ldapUsers[0])
		if err != nil {
			log.Errorf("Can't import user %s, error: %v", ldapUsers[0].Username, err)
			return nil, fmt.Errorf("can't import user %s, error: %v", ldapUsers[0].Username, err)
		}
		u.UserID = int(userID)
	}

	return &u, nil

}

func init() {
	auth.Register("ldap_auth", &Auth{})
}
