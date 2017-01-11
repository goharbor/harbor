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

package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"crypto/tls"

	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"

	goldap "gopkg.in/ldap.v2"
)

// LdapAPI handles requesst to /api/ldap/ping /api/ldap/search
type LdapAPI struct {
	api.BaseAPI
}

var ldapConfs models.LdapConf

// Prepare ...
func (l *LdapAPI) Prepare() {

	userID := l.ValidateUser()
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("error occurred in IsAdminRole: %v", err)
		l.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if !isSysAdmin {
		l.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
}

// Ping ...
func (l *LdapAPI) Ping() {
	l.DecodeJSONReqAndValidate(&ldapConfs)

	err := validateLdapReq(ldapConfs)
	if err != nil {
		log.Errorf("Invalid ldap request, error: %v", err)
		l.RenderError(http.StatusBadRequest, fmt.Sprintf("invalid ldap request: %v", err))
		return
	}

	err = connectTest(ldapConfs)
	if err != nil {
		log.Errorf("Ldap connect fail, error: %v", err)
		l.RenderError(http.StatusBadRequest, fmt.Sprintf("ldap connect fail: %v", err))
		return
	}
}

func validateLdapReq(ldapConfs models.LdapConf) error {
	ldapURL := ldapConfs.LdapURL
	if ldapURL == "" {
		return fmt.Errorf("can not get any available LDAP_URL")
	}
	log.Debug("ldapURL:", ldapURL)

	ldapConnectionTimeout := ldapConfs.LdapConnectionTimeout
	log.Debug("ldapConnectionTimeout:", ldapConnectionTimeout)

	return nil

}

func connectTest(ldapConfs models.LdapConf) error {

	var ldap *goldap.Conn
	var protocol, hostport string
	var host, port string
	var err error

	ldapURL := ldapConfs.LdapURL

	// This routine keeps compability with the old format used on harbor.cfg

	if strings.Contains(ldapURL, "://") {
		splitLdapURL := strings.Split(ldapURL, "://")
		protocol, hostport = splitLdapURL[0], splitLdapURL[1]
		if !((protocol == "ldap") || (protocol == "ldaps")) {
			return fmt.Errorf("unknown ldap protocl")
		}
	} else {
		hostport = ldapURL
		protocol = "ldap"
	}

	// This tries to detect the used port, if not defined
	if strings.Contains(hostport, ":") {
		splitHostPort := strings.Split(hostport, ":")
		host, port = splitHostPort[0], splitHostPort[1]
		_, error := strconv.Atoi(splitHostPort[1])
		if error != nil {
			return fmt.Errorf("illegal url format")
		}
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
	connectionTimeout := ldapConfs.LdapConnectionTimeout
	goldap.DefaultTimeout = time.Duration(connectionTimeout) * time.Second

	switch protocol {
	case "ldap":
		ldap, err = goldap.Dial("tcp", fmt.Sprintf("%s:%s", host, port))
	case "ldaps":
		ldap, err = goldap.DialTLS("tcp", fmt.Sprintf("%s:%s", host, port), &tls.Config{InsecureSkipVerify: true})
	}

	if err != nil {
		return err
	}
	defer ldap.Close()

	ldapSearchDn := ldapConfs.LdapSearchDn
	if ldapSearchDn != "" {
		log.Debug("Search DN: ", ldapSearchDn)
		ldapSearchPassword := ldapConfs.LdapSearchPassword
		err = ldap.Bind(ldapSearchDn, ldapSearchPassword)
		if err != nil {
			log.Debug("Bind search dn error", err)
			return err
		}
	}

	return nil

}
