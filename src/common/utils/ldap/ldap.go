// Copyright Project Harbor Authors
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
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"

	goldap "gopkg.in/ldap.v2"
)

// ErrNotFound ...
var ErrNotFound = errors.New("entity not found")

// ErrDNSyntax ...
var ErrDNSyntax = errors.New("invalid DN syntax")

// ErrInvalidFilter ...
var ErrInvalidFilter = errors.New("invalid filter syntax")

// Session - define a LDAP session
type Session struct {
	ldapConfig      models.LdapConf
	ldapGroupConfig models.LdapGroupConf
	ldapConn        *goldap.Conn
}

// LoadSystemLdapConfig - load LDAP configure
func LoadSystemLdapConfig() (*Session, error) {

	ldapConf, err := config.LDAPConf()

	if err != nil {
		return nil, err
	}

	ldapGroupConfig, err := config.LDAPGroupConf()

	if err != nil {
		return nil, err
	}

	return CreateWithAllConfig(*ldapConf, *ldapGroupConfig)
}

// CreateWithConfig -
func CreateWithConfig(ldapConf models.LdapConf) (*Session, error) {
	return CreateWithAllConfig(ldapConf, models.LdapGroupConf{})
}

// CreateWithAllConfig - create a Session with internal config
func CreateWithAllConfig(ldapConf models.LdapConf, ldapGroupConfig models.LdapGroupConf) (*Session, error) {
	var session Session

	if ldapConf.LdapURL == "" {
		return nil, fmt.Errorf("can not get any available LDAP_URL")
	}

	ldapURL, err := formatURL(ldapConf.LdapURL)
	if err != nil {
		return nil, err
	}

	ldapConf.LdapURL = ldapURL
	session.ldapConfig = ldapConf
	session.ldapGroupConfig = ldapGroupConfig
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
		port, err := strconv.Atoi(splitHostPort[1])
		if err != nil {
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

// ConnectionTest - test ldap session connection with system default setting
func (session *Session) ConnectionTest() error {
	session, err := LoadSystemLdapConfig()
	if err != nil {
		return fmt.Errorf("Failed to load system ldap config")
	}

	return ConnectionTestWithAllConfig(session.ldapConfig, session.ldapGroupConfig)
}

// ConnectionTestWithConfig -
func ConnectionTestWithConfig(ldapConfig models.LdapConf) error {
	return ConnectionTestWithAllConfig(ldapConfig, models.LdapGroupConf{})
}

// ConnectionTestWithAllConfig - test ldap session connection, out of the scope of normal session create/close
func ConnectionTestWithAllConfig(ldapConfig models.LdapConf, ldapGroupConfig models.LdapGroupConf) error {

	// If no password present, use the system default password
	if ldapConfig.LdapSearchPassword == "" {

		session, err := LoadSystemLdapConfig()

		if err != nil {
			return fmt.Errorf("Failed to load system ldap config")
		}

		ldapConfig.LdapSearchPassword = session.ldapConfig.LdapSearchPassword
	}

	testSession, err := CreateWithAllConfig(ldapConfig, ldapGroupConfig)

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

// SearchUser - search LDAP user by name
func (session *Session) SearchUser(username string) ([]models.LdapUser, error) {
	var ldapUsers []models.LdapUser
	ldapFilter, err := createUserSearchFilter(session.ldapConfig.LdapFilter, session.ldapConfig.LdapUID, username)
	if err != nil {
		return nil, err
	}

	result, err := session.SearchLdap(ldapFilter)
	if err != nil {
		return nil, err
	}

	for _, ldapEntry := range result.Entries {
		var u models.LdapUser
		groupDNList := []string{}
		groupAttr := strings.ToLower(session.ldapGroupConfig.LdapGroupMembershipAttribute)
		for _, attr := range ldapEntry.Attributes {
			// OpenLdap sometimes contain leading space in useranme
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
			case groupAttr:
				for _, dnItem := range attr.Values {
					groupDNList = append(groupDNList, strings.TrimSpace(dnItem))
					log.Debugf("Found memberof %v", dnItem)
				}
			}
			u.GroupDNList = groupDNList
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

// Open - open Session, should invoke Close for each Open call
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
	attributes := []string{"uid", "cn", "mail", "email"}
	lowerUID := strings.ToLower(session.ldapConfig.LdapUID)

	if lowerUID != "uid" && lowerUID != "cn" && lowerUID != "mail" && lowerUID != "email" {
		attributes = append(attributes, session.ldapConfig.LdapUID)
	}

	// Add the Group membership attribute
	groupAttr := strings.TrimSpace(session.ldapGroupConfig.LdapGroupMembershipAttribute)
	log.Debugf("Membership attribute: %s\n", groupAttr)
	attributes = append(attributes, groupAttr)

	return session.SearchLdapAttribute(session.ldapConfig.LdapBaseDn, filter, attributes)
}

// SearchLdapAttribute - to search ldap with the provide filter, with specified attributes
func (session *Session) SearchLdapAttribute(baseDN, filter string, attributes []string) (*goldap.SearchResult, error) {

	if err := session.Bind(session.ldapConfig.LdapSearchDn, session.ldapConfig.LdapSearchPassword); err != nil {
		return nil, fmt.Errorf("Can not bind search dn, error: %v", err)
	}
	filter = strings.TrimSpace(filter)
	if !(strings.HasPrefix(filter, "(") || strings.HasSuffix(filter, ")")) {
		filter = "(" + filter + ")"
	}
	if _, err := goldap.CompileFilter(filter); err != nil {
		log.Errorf("Wrong filter format, filter:%v", filter)
		return nil, ErrInvalidFilter
	}
	log.Debugf("Search ldap with filter:%v", filter)
	searchRequest := goldap.NewSearchRequest(
		baseDN,
		session.ldapConfig.LdapScope,
		goldap.NeverDerefAliases,
		0,     // Unlimited results
		0,     // Search Timeout
		false, // Types only
		filter,
		attributes,
		nil,
	)

	result, err := session.ldapConn.SearchWithPaging(searchRequest, common.DefaultLDAPPageSize)
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

// createUserSearchFilter - create filter to search user with specified username
func createUserSearchFilter(origFilter, ldapUID, username string) (string, error) {
	oFilter, err := NewFilterBuilder(origFilter)
	if err != nil {
		return "", err
	}
	var filterTag string
	filterTag = goldap.EscapeFilter(username)
	if len(filterTag) == 0 {
		filterTag = "*"
	}
	uFilterStr := fmt.Sprintf("(%v=%v)", ldapUID, filterTag)
	uFilter, err := NewFilterBuilder(uFilterStr)
	if err != nil {
		return "", err
	}
	filter := oFilter.And(uFilter)
	return filter.String()
}

// Close - close current session
func (session *Session) Close() {
	if session.ldapConn != nil {
		session.ldapConn.Close()
	}
}

// SearchGroupByName ...
func (session *Session) SearchGroupByName(groupName string) ([]models.LdapGroup, error) {
	return session.searchGroup(session.ldapGroupConfig.LdapGroupBaseDN,
		session.ldapGroupConfig.LdapGroupFilter,
		groupName,
		session.ldapGroupConfig.LdapGroupNameAttribute)
}

// SearchGroupByDN ...
func (session *Session) SearchGroupByDN(groupDN string) ([]models.LdapGroup, error) {
	if _, err := goldap.ParseDN(groupDN); err != nil {
		return nil, ErrDNSyntax
	}
	groupList, err := session.searchGroup(groupDN, session.ldapGroupConfig.LdapGroupFilter, "", session.ldapGroupConfig.LdapGroupNameAttribute)
	if serverError, ok := err.(*goldap.Error); ok {
		log.Debugf("resultCode:%v", serverError.ResultCode)
	}
	if err != nil && goldap.IsErrorWithCode(err, goldap.LDAPResultNoSuchObject) {
		return nil, ErrNotFound
	}
	return groupList, err
}

func (session *Session) groupBaseDN() string {
	if len(session.ldapGroupConfig.LdapGroupBaseDN) == 0 {
		return session.ldapConfig.LdapBaseDn
	}
	return session.ldapGroupConfig.LdapGroupBaseDN
}

func (session *Session) searchGroup(groupDN, filter, groupName, groupNameAttribute string) ([]models.LdapGroup, error) {
	ldapGroups := make([]models.LdapGroup, 0)
	log.Debugf("Groupname: %v, groupDN: %v", groupName, groupDN)
	ldapFilter, err := createGroupSearchFilter(filter, groupName, groupNameAttribute)
	if err != nil {
		log.Errorf("wrong filter format: filter:%v, groupName:%v, groupNameAttribute:%v", filter, groupName, groupNameAttribute)
		return nil, err
	}
	attributes := []string{groupNameAttribute}
	result, err := session.SearchLdapAttribute(session.groupBaseDN(), ldapFilter, attributes)
	if err != nil {
		return nil, err
	}
	for _, ldapEntry := range result.Entries {
		var group models.LdapGroup
		if groupDN != ldapEntry.DN {
			continue
		}
		group.GroupDN = ldapEntry.DN
		for _, attr := range ldapEntry.Attributes {
			// OpenLdap sometimes contain leading space in username
			val := strings.TrimSpace(attr.Values[0])
			log.Debugf("Current ldap entry attr name: %s\n", attr.Name)
			switch strings.ToLower(attr.Name) {
			case strings.ToLower(groupNameAttribute):
				group.GroupName = val
			}
		}
		ldapGroups = append(ldapGroups, group)
	}
	return ldapGroups, nil
}

func createGroupSearchFilter(oldFilterStr, groupName, groupNameAttribute string) (string, error) {
	origFilter, err := NewFilterBuilder(oldFilterStr)
	if err != nil {
		log.Errorf("failed to create group search filter:%v", oldFilterStr)
		return "", err
	}
	groupName = goldap.EscapeFilter(groupName)
	gFilterStr := ""
	if len(groupName) > 0 {
		gFilterStr = fmt.Sprintf("(%v=%v)", goldap.EscapeFilter(groupNameAttribute), groupName)
	}
	gFilter, err := NewFilterBuilder(gFilterStr)
	if err != nil {
		log.Errorf("invalid ldap filter:%v", gFilterStr)
		return "", err
	}
	fb := origFilter.And(gFilter)
	return fb.String()
}

// normalizeFilter - add '(' and ')' in ldap filter if it doesn't exist
func normalizeFilter(filter string) string {
	norFilter := strings.TrimSpace(filter)
	if len(norFilter) == 0 {
		return norFilter
	}
	if strings.HasPrefix(norFilter, "(") && strings.HasSuffix(norFilter, ")") {
		return norFilter
	}
	return "(" + norFilter + ")"
}
