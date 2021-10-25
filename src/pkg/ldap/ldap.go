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
	"github.com/goharbor/harbor/src/lib/config/models"
	"github.com/goharbor/harbor/src/pkg/ldap/model"
	"net"
	"net/url"
	"strings"
	"time"

	goldap "github.com/go-ldap/ldap/v3"
	"github.com/goharbor/harbor/src/lib/log"
)

// ErrNotFound ...
var ErrNotFound = errors.New("entity not found")

// ErrEmptyPassword ...
var ErrEmptyPassword = errors.New("empty password")

// ErrInvalidCredential ...
var ErrInvalidCredential = errors.New("invalid credential")

// ErrLDAPServerTimeout ...
var ErrLDAPServerTimeout = errors.New("ldap server network timeout")

// ErrLDAPPingFail ...
var ErrLDAPPingFail = errors.New("fail to ping LDAP server")

// ErrDNSyntax ...
var ErrDNSyntax = errors.New("invalid DN syntax")

// ErrInvalidFilter ...
var ErrInvalidFilter = errors.New("invalid filter syntax")

// ErrEmptyBaseDN ...
var ErrEmptyBaseDN = errors.New("empty base dn")

// ErrEmptySearchDN ...
var ErrEmptySearchDN = errors.New("empty search dn")

// Session - define a LDAP session
type Session struct {
	basicCfg models.LdapConf
	groupCfg models.GroupConf
	ldapConn *goldap.Conn
}

// NewSession create session with configs
func NewSession(basicCfg models.LdapConf, groupCfg models.GroupConf) *Session {
	return &Session{
		basicCfg: basicCfg,
		groupCfg: groupCfg,
	}
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
		_, port, err := net.SplitHostPort(hostport)
		if err != nil {
			return "", fmt.Errorf("illegal ldap url, error: %v", err)
		}
		if port == "636" {
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

// TestConfig - test ldap session connection, out of the scope of normal session create/close
func TestConfig(ldapConfig models.LdapConf) (bool, error) {
	ts := NewSession(ldapConfig, models.GroupConf{})
	if err := ts.Open(); err != nil {
		if goldap.IsErrorWithCode(err, goldap.ErrorNetwork) {
			return false, ErrLDAPServerTimeout
		}
		return false, ErrLDAPPingFail
	}
	defer ts.Close()

	if ts.basicCfg.SearchDn == "" {
		return false, ErrEmptySearchDN
	}
	if err := ts.Bind(ts.basicCfg.SearchDn, ts.basicCfg.SearchPassword); err != nil {
		if goldap.IsErrorWithCode(err, goldap.LDAPResultInvalidCredentials) {
			return false, ErrInvalidCredential
		}
	}
	return true, nil
}

// SearchUser - search LDAP user by name
func (s *Session) SearchUser(username string) ([]model.User, error) {
	var ldapUsers []model.User
	ldapFilter, err := createUserSearchFilter(s.basicCfg.Filter, s.basicCfg.UID, username)
	if err != nil {
		return nil, err
	}

	result, err := s.SearchLdap(ldapFilter)
	if err != nil {
		return nil, err
	}

	for _, ldapEntry := range result.Entries {
		var u model.User
		groupDNList := make([]string, 0)
		groupAttr := strings.ToLower(s.groupCfg.MembershipAttribute)
		for _, attr := range ldapEntry.Attributes {
			if attr == nil || len(attr.Values) == 0 {
				continue
			}
			// OpenLdap sometimes contain leading space in username
			val := strings.TrimSpace(attr.Values[0])
			log.Debugf("Current ldap entry attr name: %s\n", attr.Name)
			switch strings.ToLower(attr.Name) {
			case strings.ToLower(s.basicCfg.UID):
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
func (s *Session) Bind(dn string, password string) error {
	return s.ldapConn.Bind(dn, password)
}

// Open - open Session, should invoke Close for each Open call
func (s *Session) Open() error {
	ldapURL, err := formatURL(s.basicCfg.URL)
	if err != nil {
		return err
	}
	splitLdapURL := strings.Split(ldapURL, "://")

	protocol, hostport := splitLdapURL[0], splitLdapURL[1]
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return err
	}
	connectionTimeout := s.basicCfg.ConnectionTimeout
	goldap.DefaultTimeout = time.Duration(connectionTimeout) * time.Second

	switch protocol {
	case "ldap":
		ldap, err := goldap.Dial("tcp", hostport)
		if err != nil {
			return err
		}
		s.ldapConn = ldap
	case "ldaps":
		log.Debug("Start to dial ldaps")
		ldap, err := goldap.DialTLS("tcp", hostport, &tls.Config{ServerName: host, InsecureSkipVerify: !s.basicCfg.VerifyCert})
		if err != nil {
			return err
		}
		s.ldapConn = ldap
	}

	return nil

}

// SearchLdap to search ldap with the provide filter
func (s *Session) SearchLdap(filter string) (*goldap.SearchResult, error) {
	attributes := []string{"uid", "cn", "mail", "email"}
	lowerUID := strings.ToLower(s.basicCfg.UID)

	if lowerUID != "uid" && lowerUID != "cn" && lowerUID != "mail" && lowerUID != "email" {
		attributes = append(attributes, s.basicCfg.UID)
	}

	// Add the Group membership attribute
	groupAttr := strings.TrimSpace(s.groupCfg.MembershipAttribute)
	log.Debugf("Membership attribute: %s\n", groupAttr)
	attributes = append(attributes, groupAttr)

	return s.SearchLdapAttribute(s.basicCfg.BaseDn, filter, attributes)
}

// SearchLdapAttribute - to search ldap with the provide filter, with specified attributes
func (s *Session) SearchLdapAttribute(baseDN, filter string, attributes []string) (*goldap.SearchResult, error) {

	if err := s.Bind(s.basicCfg.SearchDn, s.basicCfg.SearchPassword); err != nil {
		return nil, fmt.Errorf("can not bind search dn, error: %v", err)
	}
	filter = normalizeFilter(filter)
	if len(filter) == 0 {
		return nil, ErrInvalidFilter
	}
	if _, err := goldap.CompileFilter(filter); err != nil {
		log.Errorf("Wrong filter format, filter:%v", filter)
		return nil, ErrInvalidFilter
	}
	log.Debugf("Search ldap with filter:%v", filter)
	searchRequest := goldap.NewSearchRequest(
		baseDN,
		s.basicCfg.Scope,
		goldap.NeverDerefAliases,
		0,     // Unlimited results
		0,     // Search Timeout
		false, // Types only
		filter,
		attributes,
		nil,
	)

	result, err := s.ldapConn.Search(searchRequest)
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
func (s *Session) Close() {
	if s.ldapConn != nil {
		s.ldapConn.Close()
	}
}

// SearchGroupByName ...
func (s *Session) SearchGroupByName(groupName string) ([]model.Group, error) {
	return s.searchGroup(s.groupCfg.BaseDN,
		s.groupCfg.Filter,
		groupName,
		s.groupCfg.NameAttribute)
}

// SearchGroupByDN ...
func (s *Session) SearchGroupByDN(groupDN string) ([]model.Group, error) {
	if _, err := goldap.ParseDN(groupDN); err != nil {
		return nil, ErrDNSyntax
	}
	groupList, err := s.searchGroup(groupDN, s.groupCfg.Filter, "", s.groupCfg.NameAttribute)
	if serverError, ok := err.(*goldap.Error); ok {
		log.Debugf("resultCode:%v", serverError.ResultCode)
	}
	if err != nil && goldap.IsErrorWithCode(err, goldap.LDAPResultNoSuchObject) {
		return nil, ErrNotFound
	}
	return groupList, err
}

func (s *Session) groupBaseDN() string {
	if len(s.groupCfg.BaseDN) == 0 {
		return s.basicCfg.BaseDn
	}
	return s.groupCfg.BaseDN
}

// searchGroup -- Given a group DN and filter, search group
func (s *Session) searchGroup(groupDN, filter, gName, groupNameAttribute string) ([]model.Group, error) {
	ldapGroups := make([]model.Group, 0)
	log.Debugf("Groupname: %v, groupDN: %v", gName, groupDN)

	// Check current group DN is under the LDAP group base DN
	isChild, err := UnderBaseDN(s.groupBaseDN(), groupDN)
	if err != nil {
		return ldapGroups, err
	}
	if !isChild {
		return ldapGroups, nil
	}

	// Search the groupDN with LDAP group filter condition
	ldapFilter, err := createGroupSearchFilter(filter, gName, groupNameAttribute)
	if err != nil {
		log.Errorf("wrong filter format: filter:%v, gName:%v, groupNameAttribute:%v", filter, gName, groupNameAttribute)
		return ldapGroups, err
	}

	// There maybe many groups under the LDAP group base DN
	// If return all groups in LDAP group base DN, it might get "Size Limit Exceeded" error
	// Take the groupDN as the baseDN in the search request to avoid return too many records
	result, err := s.SearchLdapAttribute(groupDN, ldapFilter, []string{groupNameAttribute})
	if err != nil {
		return ldapGroups, err
	}
	if len(result.Entries) == 0 {
		return ldapGroups, nil
	}
	groupName := ""
	if len(result.Entries[0].Attributes) > 0 &&
		result.Entries[0].Attributes[0] != nil &&
		len(result.Entries[0].Attributes[0].Values) > 0 {
		groupName = result.Entries[0].Attributes[0].Values[0]
	} else {
		groupName = groupDN
	}
	group := model.Group{
		Dn:   result.Entries[0].DN,
		Name: groupName,
	}
	ldapGroups = append(ldapGroups, group)

	return ldapGroups, nil
}

// UnderBaseDN - check if the childDN is under the baseDN, if the baseDN equals current DN, return true
func UnderBaseDN(baseDN, childDN string) (bool, error) {
	base, err := goldap.ParseDN(strings.ToLower(baseDN))
	if err != nil {
		return false, err
	}
	child, err := goldap.ParseDN(strings.ToLower(childDN))
	if err != nil {
		return false, err
	}
	return base.AncestorOf(child) || base.Equal(child), nil
}

// createGroupSearchFilter - Create group search filter with base filter and group name filter condition
func createGroupSearchFilter(baseFilter, groupName, groupNameAttr string) (string, error) {
	base, err := NewFilterBuilder(baseFilter)
	if err != nil {
		log.Errorf("failed to create group search filter:%v", baseFilter)
		return "", err
	}
	groupName = goldap.EscapeFilter(groupName)
	gFilterStr := ""
	// when groupName is empty, search all groups in current base DN
	if len(groupName) == 0 {
		groupName = "*"
	}
	if len(groupNameAttr) == 0 {
		groupNameAttr = "cn"
	}
	gFilter, err := NewFilterBuilder("(" + goldap.EscapeFilter(groupNameAttr) + "=" + groupName + ")")
	if err != nil {
		log.Errorf("invalid ldap filter:%v", gFilterStr)
		return "", err
	}
	fb := base.And(gFilter)
	return fb.String()
}
