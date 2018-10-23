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
	"net/http"
	"reflect"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
)

// ConfigAPI ...
type ConfigAPI struct {
	BaseController
}

// Prepare validates the user
func (c *ConfigAPI) Prepare() {
	c.BaseController.Prepare()
	if !c.SecurityCtx.IsAuthenticated() {
		c.HandleUnauthorized()
		return
	}
	if !c.SecurityCtx.IsSysAdmin() && !c.SecurityCtx.IsSolutionUser() {
		c.HandleForbidden(c.SecurityCtx.GetUsername())
		return
	}
}

type value struct {
	Value    interface{} `json:"value"`
	Editable bool        `json:"editable"`
}

// Get returns configurations
func (c *ConfigAPI) Get() {
	configs, err := config.GetSystemCfg()
	if err != nil {
		log.Errorf("failed to get configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	cfgs := map[string]interface{}{}
	for _, k := range common.HarborValidKeys {
		if v, ok := configs[k]; ok {
			cfgs[k] = v
		}
	}

	m, err := convertForGet(cfgs)
	if err != nil {
		log.Errorf("failed to convert configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	c.Data["json"] = m
	c.ServeJSON()
}

// GetInternalConfig returns internal configurations
func (c *ConfigAPI) GetInternalConfig() {
	configs, err := config.GetSystemCfg()
	if err != nil {
		log.Errorf("failed to get configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	c.Data["json"] = configs
	c.ServeJSON()
}

// Put updates configurations
func (c *ConfigAPI) Put() {
	m := map[string]interface{}{}
	c.DecodeJSONReq(&m)

	cfg := map[string]interface{}{}
	for _, k := range common.HarborValidKeys {
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

	if err := config.Upload(cfg); err != nil {
		log.Errorf("failed to upload configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if err := config.Load(); err != nil {
		log.Errorf("failed to load configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	// Everything is ok, detect the configurations to confirm if the option we are caring is changed.
	if err := watchConfigChanges(cfg); err != nil {
		log.Errorf("Failed to watch configuration change with error: %s\n", err)
	}
}

// Reset system configurations
func (c *ConfigAPI) Reset() {
	if err := config.Reset(); err != nil {
		log.Errorf("failed to reset configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
}

func validateCfg(c map[string]interface{}) (bool, error) {
	strMap := map[string]string{}
	for k := range common.HarborStringKeysMap {
		if _, ok := c[k]; !ok {
			continue
		}
		if _, ok := c[k].(string); !ok {
			return false, fmt.Errorf("Invalid value type, expected string, key: %s, value: %v, type: %v", k, c[k], reflect.TypeOf(c[k]))
		}
		strMap[k] = c[k].(string)
	}
	numMap := map[string]int{}
	for k := range common.HarborNumKeysMap {
		if _, ok := c[k]; !ok {
			continue
		}
		if _, ok := c[k].(float64); !ok {
			return false, fmt.Errorf("Invalid value type, expected float64, key: %s, value: %v, type: %v", k, c[k], reflect.TypeOf(c[k]))
		}
		numMap[k] = int(c[k].(float64))
	}
	boolMap := map[string]bool{}
	for k := range common.HarborBoolKeysMap {
		if _, ok := c[k]; !ok {
			continue
		}
		if _, ok := c[k].(bool); !ok {
			return false, fmt.Errorf("Invalid value type, expected bool, key: %s, value: %v, type: %v", k, c[k], reflect.TypeOf(c[k]))
		}
		boolMap[k] = c[k].(bool)
	}

	mode, err := config.AuthMode()
	if err != nil {
		return true, err
	}

	if value, ok := strMap[common.AUTHMode]; ok {
		if value != common.DBAuth && value != common.LDAPAuth && value != common.UAAAuth {
			return false, fmt.Errorf("invalid %s, shoud be one of %s, %s, %s", common.AUTHMode, common.DBAuth, common.LDAPAuth, common.UAAAuth)
		}
		flag, err := authModeCanBeModified()
		if err != nil {
			return true, err
		}
		if mode != value && !flag {
			return false, fmt.Errorf("%s can not be modified as new users have been inserted into database", common.AUTHMode)
		}
		mode = value
	}

	if mode == common.LDAPAuth {
		ldapConf, err := config.LDAPConf()
		if err != nil {
			return true, err
		}

		if len(ldapConf.LdapURL) == 0 {
			if _, ok := strMap[common.LDAPURL]; !ok {
				return false, fmt.Errorf("%s is missing", common.LDAPURL)
			}
		}

		if len(ldapConf.LdapBaseDn) == 0 {
			if _, ok := strMap[common.LDAPBaseDN]; !ok {
				return false, fmt.Errorf("%s is missing", common.LDAPBaseDN)
			}
		}
		if len(ldapConf.LdapUID) == 0 {
			if _, ok := strMap[common.LDAPUID]; !ok {
				return false, fmt.Errorf("%s is missing", common.LDAPUID)
			}
		}
		if ldapConf.LdapScope == 0 {
			if _, ok := numMap[common.LDAPScope]; !ok {
				return false, fmt.Errorf("%s is missing", common.LDAPScope)
			}
		}
	}

	if ldapURL, ok := strMap[common.LDAPURL]; ok && len(ldapURL) == 0 {
		return false, fmt.Errorf("%s is empty", common.LDAPURL)
	}
	if baseDN, ok := strMap[common.LDAPBaseDN]; ok && len(baseDN) == 0 {
		return false, fmt.Errorf("%s is empty", common.LDAPBaseDN)
	}
	if uID, ok := strMap[common.LDAPUID]; ok && len(uID) == 0 {
		return false, fmt.Errorf("%s is empty", common.LDAPUID)
	}
	if scope, ok := numMap[common.LDAPScope]; ok &&
		scope != common.LDAPScopeBase &&
		scope != common.LDAPScopeOnelevel &&
		scope != common.LDAPScopeSubtree {
		return false, fmt.Errorf("invalid %s, should be %d, %d or %d",
			common.LDAPScope,
			common.LDAPScopeBase,
			common.LDAPScopeOnelevel,
			common.LDAPScopeSubtree)
	}
	for k, n := range numMap {
		if n < 0 {
			return false, fmt.Errorf("invalid %s: %d", k, n)
		}
		if (k == common.EmailPort ||
			k == common.PostGreSQLPort) && n > 65535 {
			return false, fmt.Errorf("invalid %s: %d", k, n)
		}
	}

	if crt, ok := strMap[common.ProjectCreationRestriction]; ok &&
		crt != common.ProCrtRestrEveryone &&
		crt != common.ProCrtRestrAdmOnly {
		return false, fmt.Errorf("invalid %s, should be %s or %s",
			common.ProjectCreationRestriction,
			common.ProCrtRestrAdmOnly,
			common.ProCrtRestrEveryone)
	}
	return false, nil
}

// delete sensitive attrs and add editable field to every attr
func convertForGet(cfg map[string]interface{}) (map[string]*value, error) {
	result := map[string]*value{}

	for _, k := range common.HarborPasswordKeys {
		if _, ok := cfg[k]; ok {
			delete(cfg, k)
		}
	}

	if _, ok := cfg[common.ScanAllPolicy]; !ok {
		cfg[common.ScanAllPolicy] = models.DefaultScanAllPolicy
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
