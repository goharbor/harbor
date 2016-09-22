/*
   Copyright (c) 2016 VMware, Inc. All Rights Reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package ldap

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/vmware/harbor/utils/log"

	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/mqu/openldap"
)

// Auth implements Authenticator interface to authenticate against LDAP
type Auth struct{}

const metaChars = "&|!=~*<>()"

// Authenticate checks user's credential against LDAP based on basedn template and LDAP URL,
// if the check is successful a dummy record will be inserted into DB, such that this user can
// be associated to other entities in the system.
func (l *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	p := m.Principal
	for _, c := range metaChars {
		if strings.ContainsRune(p, c) {
			return nil, fmt.Errorf("the principal contains meta char: %q", c)
		}
	}
	ldapURL := os.Getenv("LDAP_URL")
	if ldapURL == "" {
		return nil, errors.New("Can not get any available LDAP_URL.")
	}
	log.Debug("ldapURL:", ldapURL)
	ldap, err := openldap.Initialize(ldapURL)
	if err != nil {
		return nil, err
	}
	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)

	ldapBaseDn := os.Getenv("LDAP_BASE_DN")
	if ldapBaseDn == "" {
		return nil, errors.New("Can not get any available LDAP_BASE_DN.")
	}
	log.Debug("baseDn:", ldapBaseDn)

	ldapSearchDn := os.Getenv("LDAP_SEARCH_DN")
	if ldapSearchDn != "" {
		log.Debug("Search DN: ", ldapSearchDn)
		ldapSearchPwd := os.Getenv("LDAP_SEARCH_PWD")
		err = ldap.Bind(ldapSearchDn, ldapSearchPwd)
		if err != nil {
			log.Debug("Bind search dn error", err)
			return nil, err
		}
	}

	attrName := os.Getenv("LDAP_UID")
	filter := os.Getenv("LDAP_FILTER")
	if filter != "" {
		filter = "(&" + filter + "(" + attrName + "=" + m.Principal + "))"
	} else {
		filter = "(" + attrName + "=" + m.Principal + ")"
	}
	log.Debug("one or more filter", filter)

	ldapScope := os.Getenv("LDAP_SCOPE")
	var scope int
	if ldapScope == "1" {
		scope = openldap.LDAP_SCOPE_BASE
	} else if ldapScope == "2" {
		scope = openldap.LDAP_SCOPE_ONELEVEL
	} else {
		scope = openldap.LDAP_SCOPE_SUBTREE
	}
	attributes := []string{"uid", "cn", "mail", "email"}
	result, err := ldap.SearchAll(ldapBaseDn, scope, filter, attributes)
	if err != nil {
		return nil, err
	}
	if len(result.Entries()) == 0 {
		log.Warningf("Not found an entry.")
		return nil, nil
	} else if len(result.Entries()) != 1 {
		log.Warningf("Found more than one entry.")
		return nil, nil
	}
	en := result.Entries()[0]
	bindDN := en.Dn()
	log.Debug("found entry:", en)
	err = ldap.Bind(bindDN, m.Password)
	if err != nil {
		log.Debug("Bind user error", err)
		return nil, err
	}
	defer ldap.Close()

	u := models.User{}
	for _, attr := range en.Attributes() {
		val := attr.Values()[0]
		switch attr.Name() {
		case "uid":
			u.Realname = val
		case "cn":
			u.Realname = val
		case "mail":
			u.Email = val
		case "email":
			u.Email = val
		}
	}
	u.Username = m.Principal
	log.Debug("username:", u.Username, ",email:", u.Email)
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
		u.Realname = m.Principal
		u.Password = "12345678AbC"
		u.Comment = "registered from LDAP."
		if u.Email == "" {
			u.Email = u.Username + "@placeholder.com"
		}
		userID, err := dao.Register(u)
		if err != nil {
			return nil, err
		}
		u.UserID = int(userID)
	}
	return &u, nil
}

func init() {
	auth.Register("ldap_auth", &Auth{})
}
