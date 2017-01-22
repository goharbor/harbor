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

// Get returns configurations
func (c *ConfigAPI) Get() {
	cfg, err := config.GetSystemCfg()
	if err != nil {
		log.Errorf("failed to get configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if cfg.Database.MySQL != nil {
		cfg.Database.MySQL.Password = ""
	}

	cfg.InitialAdminPwd = ""
	cfg.SecretKey = ""

	m := map[string]interface{}{}
	m["config"] = cfg

	editable, err := dao.AuthModeCanBeModified()
	if err != nil {
		log.Errorf("failed to determinie whether auth mode can be modified: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	m["auth_mode_editable"] = editable

	c.Data["json"] = m
	c.ServeJSON()
}

// Put updates configurations
func (c *ConfigAPI) Put() {
	m := map[string]string{}
	c.DecodeJSONReq(&m)
	if err := validateCfg(m); err != nil {
		c.CustomAbort(http.StatusBadRequest, err.Error())
	}

	if value, ok := m[comcfg.AUTHMode]; ok {
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

	if err := config.Upload(m); err != nil {
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

/*
func convert() ([]*models.Config, error) {
	cfgs := []*models.Config{}
	var err error
	pwdKeys := []string{config.LDAP_SEARCH_PWD, config.EMAIL_PWD}
	for _, pwdKey := range pwdKeys {
		if pwd, ok := c[pwdKey]; ok && len(pwd) != 0 {
			c[pwdKey], err = utils.ReversibleEncrypt(pwd, ui_cfg.SecretKey())
			if err != nil {
				return nil, err
			}
		}
	}

	for _, key := range configKeys {
		if value, ok := c[key]; ok {
			cfgs = append(cfgs, &models.Config{
				Key:   key,
				Value: value,
			})
		}
	}

	return cfgs, nil
}
*/
/*
//[]*models.Config >> cfgForGet
func convert(cfg *config.Configuration) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	for _, config := range configs {
		cfg[config.Key] = &value{
			Value:    config.Value,
			Editable: true,
		}
	}

	dels := []string{config.LDAP_SEARCH_PWD, config.EMAIL_PWD}
	for _, del := range dels {
		if _, ok := cfg[del]; ok {
			delete(cfg, del)
		}
	}

	flag, err := authModeCanBeModified()
	if err != nil {
		return nil, err
	}
	cfg[config.AUTH_MODE].Editable = flag

	return cfgForGet(cfg), nil
}
*/
func authModeCanBeModified() (bool, error) {
	return dao.AuthModeCanBeModified()
}
