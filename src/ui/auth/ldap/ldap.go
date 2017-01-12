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
	"crypto/tls"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/auth"
	"github.com/vmware/harbor/src/ui/config"

	goldap "gopkg.in/ldap.v2"
)

// Auth implements Authenticator interface to authenticate against LDAP
type Auth struct{}

const metaChars = "&|!=~*<>()"

// Connect checks the LDAP configuration directives, and connects to the LDAP URL
// Returns an LDAP connection
func Connect(settings *models.LDAP) (*goldap.Conn, error) {
	ldapURL := settings.URL
	if ldapURL == "" {
		return nil, errors.New("can not get any available LDAP_URL")
	}
	log.Debug("ldapURL:", ldapURL)

	// This routine keeps compability with the old format used on harbor.cfg
	splitLdapURL := strings.Split(ldapURL, "://")
	protocol, hostport := splitLdapURL[0], splitLdapURL[1]

	var host, port string

	// This tries to detect the used port, if not defined
	if strings.Contains(hostport, ":") {
		splitHostPort := strings.Split(hostport, ":")
		host, port = splitHostPort[0], splitHostPort[1]
	} else {
		host = hostport
		switch protocol {
		case "ldap":
			port = "389"
		case "ldaps":
			port = "636"
		}
	}

	// Sets a Dial Timeout for LDAP
	goldap.DefaultTimeout = time.Duration(settings.Timeout) * time.Second

	var ldap *goldap.Conn
	var err error
	switch protocol {
	case "ldap":
		ldap, err = goldap.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	case "ldaps":
		ldap, err = goldap.DialTLS("tcp", fmt.Sprintf("%s:%s", host, port), &tls.Config{InsecureSkipVerify: true})
	}

	if err != nil {
		return nil, err
	}

	return ldap, nil

}

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

	settings, err := config.LDAP()
	if err != nil {
		return nil, err
	}

	ldap, err := Connect(settings)
	if err != nil {
		return nil, err
	}

	ldapBaseDn := settings.BaseDN
	if ldapBaseDn == "" {
		return nil, errors.New("can not get any available LDAP_BASE_DN")
	}
	log.Debug("baseDn:", ldapBaseDn)

	ldapSearchDn := settings.SearchDN
	if ldapSearchDn != "" {
		log.Debug("Search DN: ", ldapSearchDn)
		ldapSearchPwd := settings.SearchPwd
		err = ldap.Bind(ldapSearchDn, ldapSearchPwd)
		if err != nil {
			log.Debug("Bind search dn error", err)
			return nil, err
		}
	}

	attrName := settings.UID
	filter := settings.Filter
	if filter != "" {
		filter = "(&" + filter + "(" + attrName + "=" + m.Principal + "))"
	} else {
		filter = "(" + attrName + "=" + m.Principal + ")"
	}
	log.Debug("one or more filter", filter)

	ldapScope := settings.Scope
	var scope int
	if ldapScope == 1 {
		scope = goldap.ScopeBaseObject
	} else if ldapScope == 2 {
		scope = goldap.ScopeSingleLevel
	} else {
		scope = goldap.ScopeWholeSubtree
	}
	attributes := []string{"uid", "cn", "mail", "email"}

	searchRequest := goldap.NewSearchRequest(
		ldapBaseDn,
		scope,
		goldap.NeverDerefAliases,
		0,     // Unlimited results. TODO: Limit this (as we expect only one result)?
		0,     // Search Timeout. TODO: Limit this (check what is the unit of timeout) and make configurable
		false, // Types Only
		filter,
		attributes,
		nil,
	)

	result, err := ldap.Search(searchRequest)
	if err != nil {
		return nil, err
	}

	if len(result.Entries) == 0 {
		log.Warningf("Not found an entry.")
		return nil, nil
	} else if len(result.Entries) != 1 {
		log.Warningf("Found more than one entry.")
		return nil, nil
	}

	entry := result.Entries[0]
	bindDN := entry.DN
	log.Debug("found entry:", bindDN)
	err = ldap.Bind(bindDN, m.Password)
	if err != nil {
		log.Debug("Bind user error", err)
		return nil, err
	}
	defer ldap.Close()

	u := models.User{}

	for _, attr := range entry.Attributes {
		val := attr.Values[0]
		switch attr.Name {
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
