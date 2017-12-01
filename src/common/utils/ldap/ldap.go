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
	"crypto/tls"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"

	goldap "gopkg.in/ldap.v2"
)

//Session - define a LDAP session
type Session struct {
	ldapConfig models.LdapConf
	ldapConn   *goldap.Conn
}

//LoadSystemLdapConfig - load LDAP configure from adminserver
func (session *Session) LoadSystemLdapConfig() error {
	var err error
	var authMode string

	authMode, err = config.AuthMode()
	if err != nil {
		log.Errorf("can't load auth mode from system, error: %v", err)
		return err
	}

	if authMode != "ldap_auth" {
		return fmt.Errorf("system auth_mode isn't ldap_auth, please check configuration")
	}

	ldap, err := config.LDAP()

	if err != nil {
		return err
	}
	if ldap.URL == "" {
		return fmt.Errorf("can not get any available LDAP_URL")
	}

	ldapURL, err := formatURL(ldap.URL)
	if err != nil {
		return err
	}

	session.ldapConfig.LdapURL = ldapURL
	session.ldapConfig.LdapSearchDn = ldap.SearchDN
	session.ldapConfig.LdapSearchPassword = ldap.SearchPassword
	session.ldapConfig.LdapBaseDn = ldap.BaseDN
	session.ldapConfig.LdapFilter = ldap.Filter
	session.ldapConfig.LdapUID = ldap.UID
	session.ldapConfig.LdapConnectionTimeout = ldap.Timeout
	session.ldapConfig.LdapVerifyCert = ldap.VerifyCert
	log.Debugf("Load system configuration: %v", ldap)

	switch ldap.Scope {
	case 1:
		session.ldapConfig.LdapScope = goldap.ScopeBaseObject
	case 2:
		session.ldapConfig.LdapScope = goldap.ScopeSingleLevel
	case 3:
		session.ldapConfig.LdapScope = goldap.ScopeWholeSubtree
	default:
		log.Errorf("invalid ldap search scope %v", ldap.Scope)
		return fmt.Errorf("invalid ldap search scope")
	}

	return err
}

func formatURL(ldapURL string) (string, error) {

	var protocol, hostport string
	var err error

	_, err = url.Parse(ldapURL)
	if err != nil {
		return "", fmt.Errorf("parse Ldap Host ERR: %s", err)
	}

	if strings.Contains(ldapURL, "://") {
		splitLdapURL := strings.Split(ldapURL, "://")
		protocol, hostport = splitLdapURL[0], splitLdapURL[1]
		if !((protocol == "ldap") || (protocol == "ldaps")) {
			return "", fmt.Errorf("unknown ldap protocl")
		}
	} else {
		hostport = ldapURL
		protocol = "ldap"
	}

	if strings.Contains(hostport, ":") {
		splitHostPort := strings.Split(hostport, ":")
		port, error := strconv.Atoi(splitHostPort[1])
		if error != nil {
			return "", fmt.Errorf("illegal url port")
		}
		if port == 636 {
			protocol = "ldaps"
		}

	} else {
		switch protocol {
		case "ldap":
			hostport = hostport + ":389"
		case "ldaps":
			hostport = hostport + ":636"
		}
	}

	fLdapURL := protocol + "://" + hostport

	return fLdapURL, nil

}

//ConnectionTest - test ldap session connection with system default setting
func (session *Session) ConnectionTest() error {
	err := session.LoadSystemLdapConfig()
	if err != nil {
		return fmt.Errorf("Failed to load system ldap config")
	}

	return session.ConnectionTestWithConfig(session.ldapConfig)
}

//ConnectionTestWithConfig - test ldap session connection, out of the scope of normal session create/close
func (session *Session) ConnectionTestWithConfig(ldapConfig models.LdapConf) error {

	var err error

	//If no password present, use the system default password
	if ldapConfig.LdapSearchPassword == "" {

		err = session.LoadSystemLdapConfig()

		if err != nil {
			return fmt.Errorf("Failed to load system ldap config")
		}

		ldapConfig.LdapSearchPassword = session.ldapConfig.LdapSearchPassword
	}

	err = session.CreateWithInternalConfig(ldapConfig)

	if err != nil {
		return err
	}

	defer session.Close()

	if session.ldapConfig.LdapSearchDn != "" {
		err = session.BindSearchDn()
		if err != nil {
			return err
		}
	}

	return nil
}

//SearchUser - search LDAP user by name
func (session *Session) SearchUser(username string) ([]models.LdapUser, error) {
	var ldapUsers []models.LdapUser
	ldapFilter := session.CreateUserFilter(username)
	result, err := session.SearchLdap(ldapFilter)

	if err != nil {
		return nil, err
	}

	for _, ldapEntry := range result.Entries {
		var u models.LdapUser
		for _, attr := range ldapEntry.Attributes {
			//OpenLdap sometimes contain leading space in useranme
			val := strings.TrimSpace(attr.Values[0])
			log.Debugf("Current ldap entry attr name: %s\n", attr.Name)
			switch strings.ToLower(attr.Name) {
			case strings.ToLower(session.ldapConfig.LdapUID):
				u.Username = val
			case "uid":
				u.Realname = val
			case "cn":
				u.Email = val
			case "mail":
				u.Email = val
			case "email":
				u.Email = val
			}
		}
		u.DN = ldapEntry.DN
		ldapUsers = append(ldapUsers, u)

	}

	return ldapUsers, nil

}

//ImportUser - Import user to harbor database
func (session *Session) ImportUser(user models.LdapUser) (int64, error) {
	var u models.User
	u.Username = user.Username
	u.Email = user.Email
	u.Realname = user.Realname

	log.Debug("username:", u.Username, ",email:", u.Email)

	exist, err := dao.UserExists(u, "username")
	if err != nil {
		log.Errorf("system checking user %s failed, error: %v", user.Username, err)
		return 0, fmt.Errorf("internal_error")
	}

	if exist {
		return 0, fmt.Errorf("duplicate_username")
	}

	exist, err = dao.UserExists(u, "email")
	if err != nil {
		log.Errorf("system checking %s mailbox failed, error :%v", user.Username, err)
		return 0, fmt.Errorf("internal_error")
	}

	if exist {
		return 0, fmt.Errorf("duplicate_mailbox")
	}

	u.Password = "12345678AbC"
	u.Comment = "from LDAP."
	if u.Email == "" {
		u.Email = u.Username + "@placeholder.com"
	}

	UserID, err := dao.Register(u)

	if err != nil {
		log.Errorf("system register user %s failed, error: %v", user.Username, err)
		return 0, fmt.Errorf("register_user_error")
	}

	return UserID, nil
}

// Bind with specified DN and password, used in authentication
func (session *Session) Bind(dn string, password string) error {
	return session.ldapConn.Bind(dn, password)
}

// BindSearchDn - bind current search DN
func (session *Session) BindSearchDn() error {

	err := session.Bind(session.ldapConfig.LdapSearchDn, session.ldapConfig.LdapSearchPassword)
	if err != nil {
		log.Debug("Bind search dn error", err)
	}

	return nil
}

//Create - create Session
func (session *Session) Create() error {

	var err error
	err = session.LoadSystemLdapConfig()
	if err != nil {
		return err
	}

	return session.CreateWithInternalConfig(session.ldapConfig)

}

// CreateWithUIConfig - create a Session with config from UI
func (session *Session) CreateWithUIConfig(ldapConfs models.LdapConf) error {

	switch ldapConfs.LdapScope {
	case 1:
		ldapConfs.LdapScope = goldap.ScopeBaseObject
	case 2:
		ldapConfs.LdapScope = goldap.ScopeSingleLevel
	case 3:
		ldapConfs.LdapScope = goldap.ScopeWholeSubtree
	default:
		return fmt.Errorf("invalid ldap search scope")
	}

	return session.CreateWithInternalConfig(ldapConfs)
}

// CreateWithInternalConfig - create a Session with internal config
func (session *Session) CreateWithInternalConfig(ldapConfs models.LdapConf) error {

	var err error
	var ldap *goldap.Conn

	if ldapConfs.LdapURL == "" {
		return fmt.Errorf("can not get any available LDAP_URL")
	}

	ldapURL, err := formatURL(ldapConfs.LdapURL)
	if err != nil {
		return err
	}

	session.ldapConfig.LdapURL = ldapURL
	session.ldapConfig.LdapSearchDn = ldapConfs.LdapSearchDn
	session.ldapConfig.LdapSearchPassword = ldapConfs.LdapSearchPassword
	session.ldapConfig.LdapBaseDn = ldapConfs.LdapBaseDn
	session.ldapConfig.LdapFilter = ldapConfs.LdapFilter
	session.ldapConfig.LdapUID = ldapConfs.LdapUID
	session.ldapConfig.LdapConnectionTimeout = ldapConfs.LdapConnectionTimeout
	session.ldapConfig.LdapVerifyCert = ldapConfs.LdapVerifyCert
	session.ldapConfig.LdapScope = ldapConfs.LdapScope

	splitLdapURL := strings.Split(session.ldapConfig.LdapURL, "://")
	protocol, hostport := splitLdapURL[0], splitLdapURL[1]
	host := strings.Split(hostport, ":")[0]

	connectionTimeout := session.ldapConfig.LdapConnectionTimeout
	goldap.DefaultTimeout = time.Duration(connectionTimeout) * time.Second

	switch protocol {
	case "ldap":
		ldap, err = goldap.Dial("tcp", hostport)
	case "ldaps":
		log.Debug("Start to dial ldaps")
		ldap, err = goldap.DialTLS("tcp", hostport, &tls.Config{ServerName: host, InsecureSkipVerify: !session.ldapConfig.LdapVerifyCert})
	}

	session.ldapConn = ldap

	return err

}

// SearchLdap to search ldap with the provide filter
func (session *Session) SearchLdap(filter string) (*goldap.SearchResult, error) {

	var err error

	if err := session.BindSearchDn(); err != nil {
		return nil, fmt.Errorf("Can not bind search dn, error: %v", err)
	}

	attributes := []string{"uid", "cn", "mail", "email"}
	lowerUID := strings.ToLower(session.ldapConfig.LdapUID)

	if lowerUID != "uid" && lowerUID != "cn" && lowerUID != "mail" && lowerUID != "email" {
		attributes = append(attributes, session.ldapConfig.LdapUID)
	}
	log.Debugf("Search ldap with filter:%v", filter)
	searchRequest := goldap.NewSearchRequest(
		session.ldapConfig.LdapBaseDn,
		session.ldapConfig.LdapScope,
		goldap.NeverDerefAliases,
		0,     //Unlimited results
		0,     //Search Timeout
		false, //Types only
		filter,
		attributes,
		nil,
	)

	result, err := session.ldapConn.Search(searchRequest)
	if result != nil {
		log.Debugf("Found entries:%v\n", len(result.Entries))
	} else {
		log.Debugf("No entries")
	}

	if err != nil {
		log.Debug("LDAP search error", err)
		return nil, err
	}

	return result, nil

}

//SearchAndImport - Search this user in ldap and import if exist
func (session *Session) SearchAndImport(username string) (int64, error) {
	var err error
	var userID int64

	searchFilter := session.CreateUserFilter(username)
	log.Debugf("Search LDAP with filter %v", searchFilter)
	ldapUsers, err := session.SearchUser(username)
	if err != nil {
		log.Errorf("Can not search ldap, error %v, filter: %s", err, session.ldapConfig.LdapFilter)
		return 0, err
	}

	if len(ldapUsers) > 0 {
		log.Debugf("Importing user %s to local database", ldapUsers[0].Username)
		if userID, err = session.ImportUser(ldapUsers[0]); err != nil {
			log.Errorf("Can not import ldap user to local db, error %v", err)
			return 0, err
		}
	}

	return userID, err
}

//CreateUserFilter - create filter to search user with specified username
func (session *Session) CreateUserFilter(username string) string {
	var filterTag string

	if username == "" {
		filterTag = "*"
	} else {
		filterTag = username
	}

	ldapFilter := session.ldapConfig.LdapFilter
	ldapUID := session.ldapConfig.LdapUID

	if ldapFilter == "" {
		ldapFilter = "(" + ldapUID + "=" + filterTag + ")"
	} else {
		ldapFilter = "(&" + ldapFilter + "(" + ldapUID + "=" + filterTag + "))"
	}

	log.Debug("ldap filter :", ldapFilter)

	return ldapFilter
}

//Close - close current session
func (session *Session) Close() {
	if session.ldapConn != nil {
		session.ldapConn.Close()
	}
}
