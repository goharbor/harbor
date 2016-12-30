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

	//TODO filter attr in sys config

	c.Data["json"] = cfg
	c.ServeJSON()
}

// Put updates configurations
func (c *ConfigAPI) Put() {
	m := map[string]string{}
	c.DecodeJSONReq(&m)
	if err := validateCfg(m); err != nil {
		c.CustomAbort(http.StatusBadRequest, err.Error())
	}

	if value, ok := m[comcfg.AUTH_MODE]; ok {
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
						comcfg.AUTH_MODE))
			}
		}
	}

	log.Info(m)

	if err := config.Upload(m); err != nil {
		log.Errorf("failed to upload configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if err := config.Load(); err != nil {
		log.Errorf("failed to load configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

// TODO ldap timeout, scope value
func validateCfg(c map[string]string) error {
	if value, ok := c[comcfg.AUTH_MODE]; ok {
		if value != comcfg.DB_AUTH && value != comcfg.LDAP_AUTH {
			return fmt.Errorf("invalid %s, shoud be %s or %s", comcfg.AUTH_MODE, comcfg.DB_AUTH, comcfg.LDAP_AUTH)
		}

		if value == comcfg.LDAP_AUTH {
			if _, ok := c[comcfg.LDAP_URL]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAP_URL)
			}
			if _, ok := c[comcfg.LDAP_BASE_DN]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAP_BASE_DN)
			}
			if _, ok := c[comcfg.LDAP_UID]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAP_UID)
			}
			if _, ok := c[comcfg.LDAP_SCOPE]; !ok {
				return fmt.Errorf("%s is missing", comcfg.LDAP_SCOPE)
			}
		}
	}

	if ldapURL, ok := c[comcfg.LDAP_URL]; ok && len(ldapURL) == 0 {
		return fmt.Errorf("%s is empty", comcfg.LDAP_URL)
	}
	if baseDN, ok := c[comcfg.LDAP_BASE_DN]; ok && len(baseDN) == 0 {
		return fmt.Errorf("%s is empty", comcfg.LDAP_BASE_DN)
	}
	if uID, ok := c[comcfg.LDAP_UID]; ok && len(uID) == 0 {
		return fmt.Errorf("%s is empty", comcfg.LDAP_UID)
	}
	if scope, ok := c[comcfg.LDAP_SCOPE]; ok &&
		scope != comcfg.LDAP_SCOPE_BASE &&
		scope != comcfg.LDAP_SCOPE_ONELEVEL &&
		scope != comcfg.LDAP_SCOPE_SUBTREE {
		return fmt.Errorf("invalid %s, should be %s, %s or %s",
			comcfg.LDAP_SCOPE,
			comcfg.LDAP_SCOPE_BASE,
			comcfg.LDAP_SCOPE_ONELEVEL,
			comcfg.LDAP_SCOPE_SUBTREE)
	}

	if self, ok := c[comcfg.SELF_REGISTRATION]; ok &&
		self != "true" && self != "false" {
		return fmt.Errorf("%s should be %s or %s",
			comcfg.SELF_REGISTRATION, "true", "false")
	}

	if port, ok := c[comcfg.EMAIL_SERVER_PORT]; ok {
		if p, err := strconv.Atoi(port); err != nil || p < 0 || p > 65535 {
			return fmt.Errorf("invalid %s", comcfg.EMAIL_SERVER_PORT)
		}
	}

	if ssl, ok := c[comcfg.EMAIL_SSL]; ok && ssl != "true" && ssl != "false" {
		return fmt.Errorf("%s should be true or false", comcfg.EMAIL_SSL)
	}

	if crt, ok := c[comcfg.PROJECT_CREATION_RESTRICTION]; ok &&
		crt != comcfg.PRO_CRT_RESTR_EVERYONE &&
		crt != comcfg.PRO_CRT_RESTR_ADM_ONLY {
		return fmt.Errorf("invalid %s, should be %s or %s",
			comcfg.PROJECT_CREATION_RESTRICTION,
			comcfg.PRO_CRT_RESTR_ADM_ONLY,
			comcfg.PRO_CRT_RESTR_EVERYONE)
	}

	if verify, ok := c[comcfg.VERIFY_REMOTE_CERT]; ok && verify != "true" && verify != "false" {
		return fmt.Errorf("invalid %s, should be true or false", comcfg.VERIFY_REMOTE_CERT)
	}

	if worker, ok := c[comcfg.MAX_JOB_WORKERS]; ok {
		if w, err := strconv.Atoi(worker); err != nil || w <= 0 {
			return fmt.Errorf("invalid %s", comcfg.MAX_JOB_WORKERS)
		}
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
