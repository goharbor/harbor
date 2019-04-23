// Copyright 2018 Project Harbor Authors
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

package api

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/models"
	ldapUtils "github.com/goharbor/harbor/src/common/utils/ldap"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/auth"

	"errors"
	"strings"

	"github.com/goharbor/harbor/src/core/config"
	goldap "gopkg.in/ldap.v2"
)

// LdapAPI handles requesst to /api/ldap/ping /api/ldap/user/search /api/ldap/user/import
type LdapAPI struct {
	BaseController
	ldapConfig *ldapUtils.Session
}

const (
	pingErrorMessage       = "LDAP connection test failed"
	loadSystemErrorMessage = "Can't load system configuration!"
	canNotOpenLdapSession  = "Can't open LDAP session!"
	searchLdapFailMessage  = "LDAP search failed!"
	importUserError        = "Found internal error when importing LDAP user!"
)

// Prepare ...
func (l *LdapAPI) Prepare() {
	l.BaseController.Prepare()
	if !l.SecurityCtx.IsAuthenticated() {
		l.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}
	if !l.SecurityCtx.IsSysAdmin() {
		l.SendForbiddenError(errors.New(l.SecurityCtx.GetUsername()))
		return
	}

	// check the auth_mode except ping
	if strings.EqualFold(l.Ctx.Request.RequestURI, "/api/ldap/ping") {
		return
	}
	authMode, err := config.AuthMode()
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("Can't load system configuration, error: %v", err))
		return
	}
	if authMode != "ldap_auth" {
		l.SendInternalServerError(errors.New("system auth_mode isn't ldap_auth, please check configuration"))
		return
	}
	ldapCfg, err := ldapUtils.LoadSystemLdapConfig()
	if err != nil {
		l.SendInternalServerError(fmt.Errorf("Can't load system configuration, error: %v", err))
		return
	}
	l.ldapConfig = ldapCfg

}

// Ping ...
func (l *LdapAPI) Ping() {
	var ldapConfs = models.LdapConf{
		LdapConnectionTimeout: 5,
	}
	var err error

	l.Ctx.Input.CopyBody(1 << 32)

	if string(l.Ctx.Input.RequestBody) == "" {
		ldapSession := *l.ldapConfig
		err = ldapSession.ConnectionTest()
	} else {
		isValid, err := l.DecodeJSONReqAndValidate(&ldapConfs)
		if !isValid {
			l.SendBadRequestError(err)
			return
		}
		err = ldapUtils.ConnectionTestWithConfig(ldapConfs)
	}

	if err != nil {
		l.SendInternalServerError(fmt.Errorf("LDAP connect fail, error: %v", err))
		return
	}
}

// Search ...
func (l *LdapAPI) Search() {
	var err error
	var ldapUsers []models.LdapUser
	ldapSession := *l.ldapConfig
	if err = ldapSession.Open(); err != nil {
		l.SendInternalServerError(fmt.Errorf("can't Open LDAP session, error: %v", err))
		return
	}
	defer ldapSession.Close()

	searchName := l.GetString("username")

	ldapUsers, err = ldapSession.SearchUser(searchName)

	if err != nil {
		l.SendInternalServerError(fmt.Errorf("LDAP search fail, error: %v", err))
		return
	}

	l.Data["json"] = ldapUsers
	l.ServeJSON()

}

// ImportUser ...
func (l *LdapAPI) ImportUser() {
	var ldapImportUsers models.LdapImportUser
	var ldapFailedImportUsers []models.LdapFailedImportUser

	isValid, err := l.DecodeJSONReqAndValidate(&ldapImportUsers)
	if !isValid {
		l.SendBadRequestError(err)
		return
	}

	ldapFailedImportUsers, err = importUsers(ldapImportUsers.LdapUIDList, l.ldapConfig)

	if err != nil {
		l.SendInternalServerError(fmt.Errorf("LDAP import user fail, error: %v", err))
		return
	}

	if len(ldapFailedImportUsers) > 0 {
		// Some user require json format response.
		l.SendNotFoundError(errors.New("ldap user is not found"))
		l.Data["json"] = ldapFailedImportUsers
		l.ServeJSON()
		return
	}

}

func importUsers(ldapImportUsers []string, ldapConfig *ldapUtils.Session) ([]models.LdapFailedImportUser, error) {
	var failedImportUser []models.LdapFailedImportUser
	var u models.LdapFailedImportUser

	ldapSession := *ldapConfig
	if err := ldapSession.Open(); err != nil {
		log.Errorf("Can't connect to LDAP, error: %v", err)
	}
	defer ldapSession.Close()

	for _, tempUID := range ldapImportUsers {
		u.UID = tempUID
		u.Error = ""

		if u.UID == "" {
			u.Error = "empty_uid"
			failedImportUser = append(failedImportUser, u)
			continue
		}

		if u.Error != "" {
			failedImportUser = append(failedImportUser, u)
			continue
		}

		ldapUsers, err := ldapSession.SearchUser(u.UID)
		if err != nil {
			u.UID = tempUID
			u.Error = "failed_search_user"
			failedImportUser = append(failedImportUser, u)
			log.Errorf("Invalid LDAP search request for %s, error: %v", tempUID, err)
			continue
		}

		if ldapUsers == nil || len(ldapUsers) <= 0 {
			u.UID = tempUID
			u.Error = "unknown_user"
			failedImportUser = append(failedImportUser, u)
			continue
		}

		var user models.User

		user.Username = ldapUsers[0].Username
		user.Realname = ldapUsers[0].Realname
		user.Email = ldapUsers[0].Email
		err = auth.OnBoardUser(&user)

		if err != nil || user.UserID <= 0 {
			u.UID = tempUID
			u.Error = err.Error()
			failedImportUser = append(failedImportUser, u)
			log.Errorf("Can't import user %s, error: %s", tempUID, u.Error)
		}

	}

	return failedImportUser, nil
}

// SearchGroup ... Search LDAP by groupname
func (l *LdapAPI) SearchGroup() {
	var ldapGroups []models.LdapGroup
	var err error
	searchName := l.GetString("groupname")
	groupDN := l.GetString("groupdn")
	ldapSession := *l.ldapConfig
	ldapSession.Open()
	defer ldapSession.Close()

	// Search LDAP group by groupName or group DN
	if len(searchName) > 0 {
		ldapGroups, err = ldapSession.SearchGroupByName(searchName)
		if err != nil {
			l.SendInternalServerError(fmt.Errorf("can't search LDAP group by name, error: %v", err))
			return
		}
	} else if len(groupDN) > 0 {
		if _, err := goldap.ParseDN(groupDN); err != nil {
			l.SendBadRequestError(fmt.Errorf("invalid DN: %v", err))
			return
		}
		ldapGroups, err = ldapSession.SearchGroupByDN(groupDN)
		if err != nil {
			// OpenLDAP usually return an error if DN is not found
			l.SendNotFoundError(fmt.Errorf("search LDAP group fail, error: %v", err))
			return
		}
	}
	if len(ldapGroups) == 0 {
		l.SendNotFoundError(errors.New("No ldap group found"))
		return
	}
	l.Data["json"] = ldapGroups
	l.ServeJSON()
}
