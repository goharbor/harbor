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
	//"strings"

	"github.com/vmware/harbor/src/common/api"
	comcfg "github.com/vmware/harbor/src/common/config"
	"github.com/vmware/harbor/src/common/dao"
	//"github.com/vmware/harbor/src/common/models"
	//"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	"github.com/vmware/harbor/src/ui/config"
)

// keys of attrs which user can modify
var validKeys = []string{
	comcfg.AUTHMode,
	comcfg.EmailFrom,
	comcfg.EmailHost,
	comcfg.EmailIdentity,
	comcfg.EmailPassword,
	comcfg.EmailPort,
	comcfg.EmailSSL,
	comcfg.EmailUsername,
	comcfg.LDAPBaseDN,
	comcfg.LDAPFilter,
	comcfg.LDAPScope,
	comcfg.LDAPSearchDN,
	comcfg.LDAPSearchPwd,
	comcfg.LDAPTimeout,
	comcfg.LDAPUID,
	comcfg.LDAPURL,
	comcfg.ProjectCreationRestriction,
	comcfg.SelfRegistration,
	comcfg.VerifyRemoteCert,
}

var numKeys = []string{
	comcfg.EmailPort,
	comcfg.LDAPScope,
	comcfg.LDAPTimeout,
}

var boolKeys = []string{
	comcfg.EmailSSL,
	comcfg.SelfRegistration,
	comcfg.VerifyRemoteCert,
}

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

	if err := validateCfg(cfg); err != nil {
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

func validateCfg(c map[string]string) error {
	if value, ok := c[comcfg.AUTHMode]; ok {
		if value != comcfg.DBAuth && value != comcfg.LDAPAuth {
			return fmt.Errorf("invalid %s, shoud be %s or %s", comcfg.AUTHMode, comcfg.DBAuth, comcfg.LDAPAuth)
		}

		if value == comcfg.LDAPAuth {
			if _, ok := c[comcfg.LDAPURL]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAPURL)
			}
			if _, ok := c[comcfg.LDAPBaseDN]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAPBaseDN)
			}
			if _, ok := c[comcfg.LDAPUID]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAPUID)
			}
			if _, ok := c[comcfg.LDAPScope]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAPScope)
			}
		}
	}

	if ldapURL, ok := c[comcfg.LDAPURL]; ok && len(ldapURL) == 0 {
		return fmt.Errorf("%s is empty", comcfg.LDAPURL)
	}
	if baseDN, ok := c[comcfg.LDAPBaseDN]; ok && len(baseDN) == 0 {
		return fmt.Errorf("%s is empty", comcfg.LDAPBaseDN)
	}
	if uID, ok := c[comcfg.LDAPUID]; ok && len(uID) == 0 {
		return fmt.Errorf("%s is empty", comcfg.LDAPUID)
	}
	if scope, ok := c[comcfg.LDAPScope]; ok &&
		scope != comcfg.LDAPScopeBase &&
		scope != comcfg.LDAPScopeOnelevel &&
		scope != comcfg.LDAPScopeSubtree {
		return fmt.Errorf("invalid %s, should be %s, %s or %s",
			comcfg.LDAPScope,
			comcfg.LDAPScopeBase,
			comcfg.LDAPScopeOnelevel,
			comcfg.LDAPScopeSubtree)
	}
	if timeout, ok := c[comcfg.LDAPTimeout]; ok {
		if t, err := strconv.Atoi(timeout); err != nil || t < 0 {
			return fmt.Errorf("invalid %s", comcfg.LDAPTimeout)
		}
	}

	if self, ok := c[comcfg.SelfRegistration]; ok &&
		self != "0" && self != "1" {
		return fmt.Errorf("%s should be %s or %s",
			comcfg.SelfRegistration, "0", "1")
	}

	if port, ok := c[comcfg.EmailPort]; ok {
		if p, err := strconv.Atoi(port); err != nil || p < 0 || p > 65535 {
			return fmt.Errorf("invalid %s", comcfg.EmailPort)
		}
	}

	if ssl, ok := c[comcfg.EmailSSL]; ok && ssl != "0" && ssl != "1" {
		return fmt.Errorf("%s should be %s or %s", comcfg.EmailSSL, "0", "1")
	}

	if crt, ok := c[comcfg.ProjectCreationRestriction]; ok &&
		crt != comcfg.ProCrtRestrEveryone &&
		crt != comcfg.ProCrtRestrAdmOnly {
		return fmt.Errorf("invalid %s, should be %s or %s",
			comcfg.ProjectCreationRestriction,
			comcfg.ProCrtRestrAdmOnly,
			comcfg.ProCrtRestrEveryone)
	}

	if verify, ok := c[comcfg.VerifyRemoteCert]; ok && verify != "0" && verify != "1" {
		return fmt.Errorf("invalid %s, should be %s or %s",
			comcfg.VerifyRemoteCert, "0", "1")
	}

	return nil
}

//encode passwords and convert map[string]string to map[string]interface{}
func convertForPut(m map[string]string) (map[string]interface{}, error) {
	cfg := map[string]interface{}{}

	/*
		pwdKeys := []string{config.LDAP_SEARCH_PWD, config.EMAIL_PWD}
		for _, pwdKey := range pwdKeys {
			if pwd, ok := c[pwdKey]; ok && len(pwd) != 0 {
				c[pwdKey], err = utils.ReversibleEncrypt(pwd, ui_cfg.SecretKey())
				if err != nil {
					return nil, err
				}
			}
		}
	*/
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

	dels := []string{
		comcfg.AdminInitialPassword,
		comcfg.EmailPassword,
		comcfg.LDAPSearchPwd,
		comcfg.MySQLPassword,
		comcfg.SecretKey}
	for _, del := range dels {
		if _, ok := cfg[del]; ok {
			delete(cfg, del)
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
