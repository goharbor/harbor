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

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/opt_auth"

	"github.com/astaxie/beego"
	"github.com/mqu/openldap"
)

type LdapAuth struct{}

const META_CHARS = "&|!=~*<>()"

func (l *LdapAuth) Validate(auth models.AuthModel) (*models.User, error) {

	ldapUrl := os.Getenv("LDAP_URL")
	if ldapUrl == "" {
		return nil, errors.New("Can not get any available LDAP_URL.")
	}
	beego.Debug("ldapUrl:", ldapUrl)

	p := auth.Principal
	for _, c := range META_CHARS {
		if strings.ContainsRune(p, c) {
			log.Printf("The principal contains meta char: %q", c)
			return nil, nil
		}
	}

	ldap, err := openldap.Initialize(ldapUrl)
	if err != nil {
		return nil, err
	}
	defer ldap.Close()

	ldap.SetOption(openldap.LDAP_OPT_PROTOCOL_VERSION, openldap.LDAP_VERSION3)

	ldapBaseDn := os.Getenv("LDAP_BASE_DN")
	if ldapBaseDn == "" {
		return nil, errors.New("Can not get any available LDAP_BASE_DN.")
	}

	baseDn := fmt.Sprintf(ldapBaseDn, auth.Principal)
	beego.Debug("baseDn:", baseDn)

	err = ldap.Bind(baseDn, auth.Password)
	if err != nil {
		return nil, err
	}

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
		u.UserId = currentUser.UserId
	} else {
		u.Password = "12345678AbC"
		u.Comment = "registered from LDAP."
		userId, err := dao.Register(u)
		if err != nil {
			return nil, err
		}
		u.UserId = int(userId)
	}
	return &u, nil
}

func init() {
	opt_auth.Register("ldap_auth", &LdapAuth{})
}
