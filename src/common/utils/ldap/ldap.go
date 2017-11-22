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
func LoadSystemLdapConfig() (*Session, error) {
	var session Session

	authMode, err := config.AuthMode()
	if err != nil {
		log.Errorf("can't load auth mode from system, error: %v", err)
		return nil, err
	}

	if authMode != "ldap_auth" {
		return nil, fmt.Errorf("system auth_mode isn't ldap_auth, please check configuration")
	}

	ldap, err := config.LDAP()

	if err != nil {
		return nil, err
	}
	if ldap.URL == "" {
		return nil, fmt.Errorf("can not get any available LDAP_URL")
	}

	ldapURL, err := formatURL(ldap.URL)
	if err != nil {
		return nil, err
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
		return nil, fmt.Errorf("invalid ldap search scope")
	}

	return &session, nil
}

// CreateWithUIConfig - create a Session with config from UI
func CreateWithUIConfig(ldapConfs models.LdapConf) (*Session, error) {

	switch ldapConfs.LdapScope {
	case 1:
		ldapConfs.LdapScope = goldap.ScopeBaseObject
	case 2:
		ldapConfs.LdapScope = goldap.ScopeSingleLevel
	case 3:
		ldapConfs.LdapScope = goldap.ScopeWholeSubtree
	default:
		return nil, fmt.Errorf("invalid ldap search scope")
	}

	return createWithInternalConfig(ldapConfs)
}

// createWithInternalConfig - create a Session with internal config
func createWithInternalConfig(ldapConfs models.LdapConf) (*Session, error) {

	var session Session

	if ldapConfs.LdapURL == "" {
		return nil, fmt.Errorf("can not get any available LDAP_URL")
	}

	ldapURL, err := formatURL(ldapConfs.LdapURL)
	if err != nil {
		return nil, err
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
	return &session, nil

}

func formatURL(ldapURL string) (string, error) {

	var protocol, hostport string

	_, err := url.Parse(ldapURL)
	if err != nil {
		return "", fmt.Errorf("parse Ldap Host ERR: %s", err)
	}

	if strings.Contains(ldapURL, "://") {
		splitLdapURL := strings.Split(ldapURL, "://")
		protocol, hostport = splitLdapURL[0], splitLdapURL[1]
		if !((protocol == "ldap") || (protocol == "ldaps")) {
			return "", fmt.Errorf("unknown ldap protocol")
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
	session, err := LoadSystemLdapConfig()
	if err != nil {
		return fmt.Errorf("Failed to load system ldap config")
	}

	return ConnectionTestWithConfig(session.ldapConfig)
}

//ConnectionTestWithConfig - test ldap session connection, out of the scope of normal session create/close
func ConnectionTestWithConfig(ldapConfig models.LdapConf) error {

	//If no password present, use the system default password
	if ldapConfig.LdapSearchPassword == "" {

		session, err := LoadSystemLdapConfig()

		if err != nil {
			return fmt.Errorf("Failed to load system ldap config")
		}

		ldapConfig.LdapSearchPassword = session.ldapConfig.LdapSearchPassword
	}

	testSession, err := createWithInternalConfig(ldapConfig)

	if err != nil {
		return err
	}
	err = testSession.Open()

	if err != nil {
		return err
	}

	defer testSession.Close()

	if testSession.ldapConfig.LdapSearchDn != "" {
		err = testSession.Bind(testSession.ldapConfig.LdapSearchDn, testSession.ldapConfig.LdapSearchPassword)
		if err != nil {
			return err
		}
	}

	return nil
}

//SearchUser - search LDAP user by name
func (session *Session) SearchUser(username string) ([]models.LdapUser, error) {
	var ldapUsers []models.LdapUser
	ldapFilter := session.createUserFilter(username)
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
				u.Realname = val
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

// Bind with specified DN and password, used in authentication
func (session *Session) Bind(dn string, password string) error {
	return session.ldapConn.Bind(dn, password)
}

//Open - open Session
func (session *Session) Open() error {

	splitLdapURL := strings.Split(session.ldapConfig.LdapURL, "://")
	protocol, hostport := splitLdapURL[0], splitLdapURL[1]
	host := strings.Split(hostport, ":")[0]

	connectionTimeout := session.ldapConfig.LdapConnectionTimeout
	goldap.DefaultTimeout = time.Duration(connectionTimeout) * time.Second

	switch protocol {
	case "ldap":
		ldap, err := goldap.Dial("tcp", hostport)
		if err != nil {
			return err
		}
		session.ldapConn = ldap
	case "ldaps":
		log.Debug("Start to dial ldaps")
		ldap, err := goldap.DialTLS("tcp", hostport, &tls.Config{ServerName: host, InsecureSkipVerify: !session.ldapConfig.LdapVerifyCert})
		if err != nil {
			return err
		}
		session.ldapConn = ldap
	}

	return nil

}

// SearchLdap to search ldap with the provide filter
func (session *Session) SearchLdap(filter string) (*goldap.SearchResult, error) {

	if err := session.Bind(session.ldapConfig.LdapSearchDn, session.ldapConfig.LdapSearchPassword); err != nil {
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

//CreateUserFilter - create filter to search user with specified username
func (session *Session) createUserFilter(username string) string {
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
