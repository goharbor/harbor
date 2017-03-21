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

	"github.com/vmware/harbor/src/common/api"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

var (
	// valid keys of configurations which user can modify
	validKeys = []string{
		comcfg.ExtEndpoint,
		comcfg.AUTHMode,
		comcfg.DatabaseType,
		comcfg.MySQLHost,
		comcfg.MySQLPort,
		comcfg.MySQLUsername,
		comcfg.MySQLPassword,
		comcfg.MySQLDatabase,
		comcfg.SQLiteFile,
		comcfg.SelfRegistration,
		comcfg.LDAPURL,
		comcfg.LDAPSearchDN,
		comcfg.LDAPSearchPwd,
		comcfg.LDAPBaseDN,
		comcfg.LDAPUID,
		comcfg.LDAPFilter,
		comcfg.LDAPScope,
		comcfg.LDAPTimeout,
		comcfg.TokenServiceURL,
		comcfg.RegistryURL,
		comcfg.EmailHost,
		comcfg.EmailPort,
		comcfg.EmailUsername,
		comcfg.EmailPassword,
		comcfg.EmailFrom,
		comcfg.EmailSSL,
		comcfg.EmailIdentity,
		comcfg.ProjectCreationRestriction,
		comcfg.VerifyRemoteCert,
		comcfg.MaxJobWorkers,
		comcfg.TokenExpiration,
		comcfg.CfgExpiration,
		comcfg.JobLogDir,
		comcfg.UseCompressedJS,
		comcfg.AdminInitialPassword,
	}

	numKeys = []string{
		comcfg.EmailPort,
		comcfg.LDAPScope,
		comcfg.LDAPTimeout,
		comcfg.MySQLPort,
		comcfg.MaxJobWorkers,
		comcfg.TokenExpiration,
		comcfg.CfgExpiration,
	}

	boolKeys = []string{
		comcfg.EmailSSL,
		comcfg.SelfRegistration,
		comcfg.VerifyRemoteCert,
		comcfg.UseCompressedJS,
	}

	passwordKeys = []string{
		comcfg.AdminInitialPassword,
		comcfg.EmailPassword,
		comcfg.LDAPSearchPwd,
		comcfg.MySQLPassword,
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

	if value, ok := cfg[comcfg.AUTHMode]; ok {
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
						comcfg.AUTHMode))
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

	if value, ok := c[comcfg.AUTHMode]; ok {
		if value != comcfg.DBAuth && value != comcfg.LDAPAuth {
			return isSysErr, fmt.Errorf("invalid %s, shoud be %s or %s", comcfg.AUTHMode, comcfg.DBAuth, comcfg.LDAPAuth)
		}
		mode = value
	}

	if mode == comcfg.LDAPAuth {
		ldap, err := config.LDAP()
		if err != nil {
			isSysErr = true
			return isSysErr, err
		}

		if len(ldap.URL) == 0 {
			if _, ok := c[comcfg.LDAPURL]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", comcfg.LDAPURL)
			}
		}

		if len(ldap.BaseDN) == 0 {
			if _, ok := c[comcfg.LDAPBaseDN]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", comcfg.LDAPBaseDN)
			}
		}
		if len(ldap.UID) == 0 {
			if _, ok := c[comcfg.LDAPUID]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", comcfg.LDAPUID)
			}
		}
		if ldap.Scope == 0 {
			if _, ok := c[comcfg.LDAPScope]; !ok {
				return isSysErr, fmt.Errorf("%s is missing", comcfg.LDAPScope)
			}
		}
	}

	if ldapURL, ok := c[comcfg.LDAPURL]; ok && len(ldapURL) == 0 {
		return isSysErr, fmt.Errorf("%s is empty", comcfg.LDAPURL)
	}
	if baseDN, ok := c[comcfg.LDAPBaseDN]; ok && len(baseDN) == 0 {
		return isSysErr, fmt.Errorf("%s is empty", comcfg.LDAPBaseDN)
	}
	if uID, ok := c[comcfg.LDAPUID]; ok && len(uID) == 0 {
		return isSysErr, fmt.Errorf("%s is empty", comcfg.LDAPUID)
	}
	if scope, ok := c[comcfg.LDAPScope]; ok &&
		scope != comcfg.LDAPScopeBase &&
		scope != comcfg.LDAPScopeOnelevel &&
		scope != comcfg.LDAPScopeSubtree {
		return isSysErr, fmt.Errorf("invalid %s, should be %s, %s or %s",
			comcfg.LDAPScope,
			comcfg.LDAPScopeBase,
			comcfg.LDAPScopeOnelevel,
			comcfg.LDAPScopeSubtree)
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

		if (k == comcfg.EmailPort ||
			k == comcfg.MySQLPort) && n > 65535 {
			return isSysErr, fmt.Errorf("invalid %s: %s", k, v)
		}
	}

	if crt, ok := c[comcfg.ProjectCreationRestriction]; ok &&
		crt != comcfg.ProCrtRestrEveryone &&
		crt != comcfg.ProCrtRestrAdmOnly {
		return isSysErr, fmt.Errorf("invalid %s, should be %s or %s",
			comcfg.ProjectCreationRestriction,
			comcfg.ProCrtRestrAdmOnly,
			comcfg.ProCrtRestrEveryone)
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
	result[comcfg.AUTHMode].Editable = flag

	return result, nil
}

func authModeCanBeModified() (bool, error) {
	return dao.AuthModeCanBeModified()
}
