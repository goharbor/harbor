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

package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/api"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

var (
	// valid keys of configurations which user can modify
	validKeys = []string{
		common.ExtEndpoint,
		common.AUTHMode,
		common.DatabaseType,
		common.MySQLHost,
		common.MySQLPort,
		common.MySQLUsername,
		common.MySQLPassword,
		common.MySQLDatabase,
		common.SQLiteFile,
		common.SelfRegistration,
		common.LDAPURL,
		common.LDAPSearchDN,
		common.LDAPSearchPwd,
		common.LDAPBaseDN,
		common.LDAPUID,
		common.LDAPFilter,
		common.LDAPScope,
		common.LDAPTimeout,
		common.TokenServiceURL,
		common.RegistryURL,
		common.EmailHost,
		common.EmailPort,
		common.EmailUsername,
		common.EmailPassword,
		common.EmailFrom,
		common.EmailSSL,
		common.EmailIdentity,
		common.ProjectCreationRestriction,
		common.VerifyRemoteCert,
		common.MaxJobWorkers,
		common.TokenExpiration,
		common.CfgExpiration,
		common.JobLogDir,
		common.AdminInitialPassword,
	}

	numKeys = []string{
		common.EmailPort,
		common.LDAPScope,
		common.LDAPTimeout,
		common.MySQLPort,
		common.MaxJobWorkers,
		common.TokenExpiration,
		common.CfgExpiration,
	}

	boolKeys = []string{
		common.EmailSSL,
		common.SelfRegistration,
		common.VerifyRemoteCert,
	}

	passwordKeys = []string{
		common.AdminInitialPassword,
		common.EmailPassword,
		common.LDAPSearchPwd,
		common.MySQLPassword,
	}
)

// ConfigAPI ...
type ConfigAPI struct {
	api.BaseAPI
}

// Prepare validates the user
func (c *ConfigAPI) Prepare() {
	userID := c.ValidateUser()
	isSysAdmin, err := dao.IsAdminRole(userID)
	if err != nil {
		log.Errorf("failed to check the role of user: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if !isSysAdmin {
		c.CustomAbort(http.StatusForbidden, http.StatusText(http.StatusForbidden))
	}
}

type value struct {
	Value    interface{} `json:"value"`
	Editable bool        `json:"editable"`
}

// Get returns configurations
func (c *ConfigAPI) Get() {
	cfg, err := config.GetSystemCfg()
	if err != nil {
		log.Errorf("failed to get configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	m, err := convertForGet(cfg)
	if err != nil {
		log.Errorf("failed to convert configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	c.Data["json"] = m
	c.ServeJSON()
}

// Put updates configurations
func (c *ConfigAPI) Put() {
	m := map[string]string{}
	c.DecodeJSONReq(&m)

	cfg := map[string]string{}
	for _, k := range validKeys {
		if v, ok := m[k]; ok {
			cfg[k] = v
		}
	}

	isSysErr, err := validateCfg(cfg)

	if err != nil {
		if isSysErr {
			log.Errorf("failed to validate configurations: %v", err)
			c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.CustomAbort(http.StatusBadRequest, err.Error())
	}

	if value, ok := cfg[common.AUTHMode]; ok {
		mode, err := config.AuthMode()
		if err != nil {
			log.Errorf("failed to get auth mode: %v", err)
			c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		if mode != value {
			flag, err := authModeCanBeModified()
			if err != nil {
				log.Errorf("failed to determine whether auth mode can be modified: %v", err)
				c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
			}

			if !flag {
				c.CustomAbort(http.StatusBadRequest,
					fmt.Sprintf("%s can not be modified as new users have been inserted into database",
						common.AUTHMode))
			}
		}
	}

	result, err := convertForPut(cfg)
	if err != nil {
		log.Errorf("failed to convert configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if err := config.Upload(result); err != nil {
		log.Errorf("failed to upload configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if err := config.Load(); err != nil {
		log.Errorf("failed to load configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// Reset system configurations
func (c *ConfigAPI) Reset() {
	if err := config.Reset(); err != nil {
		log.Errorf("failed to reset configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func validateCfg(c map[string]string) (bool, error) {
	isSysErr := false

	mode, err := config.AuthMode()
	if err != nil {
		isSysErr = true
		return isSysErr, err
	}

	if value, ok := c[common.AUTHMode]; ok {
		if value != common.DBAuth && value != common.LDAPAuth {
			return isSysErr, fmt.Errorf("invalid %s, shoud be %s or %s", common.AUTHMode, common.DBAuth, common.LDAPAuth)
		}
		mode = value
	}

	if mode == common.LDAPAuth {
		ldap, err := config.LDAP()
		if err != nil {
			isSysErr = true
			return isSysErr, err
		}

		if len(ldap.URL) == 0 {
			if _, ok := c[common.LDAPURL]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", common.LDAPURL)
			}
		}

		if len(ldap.BaseDN) == 0 {
			if _, ok := c[common.LDAPBaseDN]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", common.LDAPBaseDN)
			}
		}
		if len(ldap.UID) == 0 {
			if _, ok := c[common.LDAPUID]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", common.LDAPUID)
			}
		}
		if ldap.Scope == 0 {
			if _, ok := c[common.LDAPScope]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", common.LDAPScope)
			}
		}
	}

	if ldapURL, ok := c[common.LDAPURL]; ok && len(ldapURL) == 0 {
		return isSysErr, fmt.Errorf("%s is empty", common.LDAPURL)
	}
	if baseDN, ok := c[common.LDAPBaseDN]; ok && len(baseDN) == 0 {
		return isSysErr, fmt.Errorf("%s is empty", common.LDAPBaseDN)
	}
	if uID, ok := c[common.LDAPUID]; ok && len(uID) == 0 {
		return isSysErr, fmt.Errorf("%s is empty", common.LDAPUID)
	}
	if scope, ok := c[common.LDAPScope]; ok &&
		scope != common.LDAPScopeBase &&
		scope != common.LDAPScopeOnelevel &&
		scope != common.LDAPScopeSubtree {
		return isSysErr, fmt.Errorf("invalid %s, should be %s, %s or %s",
			common.LDAPScope,
			common.LDAPScopeBase,
			common.LDAPScopeOnelevel,
			common.LDAPScopeSubtree)
	}

	for _, k := range boolKeys {
		v, ok := c[k]
		if !ok {
			continue
		}

		if v != "0" && v != "1" {
			return isSysErr, fmt.Errorf("%s should be %s or %s",
				k, "0", "1")
		}
	}

	for _, k := range numKeys {
		v, ok := c[k]
		if !ok {
			continue
		}

		n, err := strconv.Atoi(v)
		if err != nil || n < 0 {
			return isSysErr, fmt.Errorf("invalid %s: %s", k, v)
		}

		if (k == common.EmailPort ||
			k == common.MySQLPort) && n > 65535 {
			return isSysErr, fmt.Errorf("invalid %s: %s", k, v)
		}
	}

	if crt, ok := c[common.ProjectCreationRestriction]; ok &&
		crt != common.ProCrtRestrEveryone &&
		crt != common.ProCrtRestrAdmOnly {
		return isSysErr, fmt.Errorf("invalid %s, should be %s or %s",
			common.ProjectCreationRestriction,
			common.ProCrtRestrAdmOnly,
			common.ProCrtRestrEveryone)
	}

	return isSysErr, nil
}

//convert map[string]string to map[string]interface{}
func convertForPut(m map[string]string) (map[string]interface{}, error) {
	cfg := map[string]interface{}{}

	for k, v := range m {
		cfg[k] = v
	}

	for _, k := range numKeys {
		if _, ok := cfg[k]; !ok {
			continue
		}

		v, err := strconv.Atoi(cfg[k].(string))
		if err != nil {
			return nil, err
		}
		cfg[k] = v
	}

	for _, k := range boolKeys {
		if _, ok := cfg[k]; !ok {
			continue
		}

		cfg[k] = cfg[k] == "1"
	}

	return cfg, nil
}

// delete sensitive attrs and add editable field to every attr
func convertForGet(cfg map[string]interface{}) (map[string]*value, error) {
	result := map[string]*value{}

	for _, k := range passwordKeys {
		if _, ok := cfg[k]; ok {
			delete(cfg, k)
		}
	}

	for k, v := range cfg {
		result[k] = &value{
			Value:    v,
			Editable: true,
		}
	}

	flag, err := authModeCanBeModified()
	if err != nil {
		return nil, err
	}
	result[common.AUTHMode].Editable = flag

	return result, nil
}

func authModeCanBeModified() (bool, error) {
	return dao.AuthModeCanBeModified()
}
