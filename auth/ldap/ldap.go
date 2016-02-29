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
	"log"
	"os"
	"strings"

	"github.com/vmware/harbor/auth"
	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego"
	"github.com/mqu/openldap"
)

// Auth implements Authenticator interface to authenticate against LDAP
type Auth struct{}

const metaChars = "&|!=~*<>()"

// Authenticate checks user's credential agains LDAP based on basedn template and LDAP URL,
// if the check is successful a dummy record will be insert into DB, such that this user can
// be associated to other entities in the system.
func (l *Auth) Authenticate(m models.AuthModel) (*models.User, error) {

	ldapURL := os.Getenv("LDAP_URL")
	if ldapURL == "" {
		return nil, errors.New("Can not get any available LDAP_URL.")
	}
	beego.Debug("ldapURL:", ldapURL)

	p := m.Principal
	for _, c := range metaChars {
		if strings.ContainsRune(p, c) {
			return nil, fmt.Errorf("the principal contains meta char: %q", c)
		}
	}

	ldap, err := openldap.Initialize(ldapURL)
	if err != nil {
		return nil, err
	}

	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)

	ldapBaseDn := os.Getenv("LDAP_BASE_DN")
	if ldapBaseDn == "" {
		return nil, errors.New("Can not get any available LDAP_BASE_DN.")
	}

	baseDn := fmt.Sprintf(ldapBaseDn, m.Principal)
	beego.Debug("baseDn:", baseDn)

	err = ldap.Bind(baseDn, m.Password)
	if err != nil {
		return nil, err
	}
	defer ldap.Close()

	scope := openldap.LDAP_SCOPE_SUBTREE // LDAP_SCOPE_BASE, LDAP_SCOPE_ONELEVEL, LDAP_SCOPE_SUBTREE
	filter := "objectClass=*"
	attributes := []string{"cn", "mail", "uid"}

	result, err := ldap.SearchAll(baseDn, scope, filter, attributes)
	if err != nil {
		return nil, err
	}
	if len(result.Entries()) != 1 {
		log.Printf("Found more than one entry.")
		return nil, nil
	}
	en := result.Entries()[0]
	u := models.User{}
	for _, attr := range en.Attributes() {
		val := attr.Values()[0]
		switch attr.Name() {
		case "uid":
			u.Username = val
		case "mail":
			u.Email = val
		case "cn":
			u.Realname = val
		}
	}

	beego.Debug("username:", u.Username, ",email:", u.Email, ",realname:", u.Realname)

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
		u.Password = "12345678AbC"
		u.Comment = "registered from LDAP."
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
