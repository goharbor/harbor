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
	"strings"

	"errors"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/config/metadata"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/common/utils/log"
	corecfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/filter"
)

// ConfigAPI ...
type ConfigAPI struct {
	BaseController
	cfgManager *config.CfgManager
}

// Prepare validates the user
func (c *ConfigAPI) Prepare() {
	c.BaseController.Prepare()
	c.cfgManager = corecfg.GetCfgManager()
	if !c.SecurityCtx.IsAuthenticated() {
		c.SendUnAuthorizedError(errors.New("UnAuthorized"))
		return
	}

	// Only internal container can access /api/internal/configurations
	if strings.EqualFold(c.Ctx.Request.RequestURI, "/api/internal/configurations") {
		if _, ok := c.Ctx.Request.Context().Value(filter.SecurCtxKey).(*secret.SecurityContext); !ok {
			c.SendUnAuthorizedError(errors.New("UnAuthorized"))
			return
		}
	}

	if !c.SecurityCtx.IsSysAdmin() && !c.SecurityCtx.IsSolutionUser() {
		c.SendForbiddenError(errors.New(c.SecurityCtx.GetUsername()))
		return
	}

}

type value struct {
	Value    interface{} `json:"value"`
	Editable bool        `json:"editable"`
}

// Get returns configurations
func (c *ConfigAPI) Get() {
	configs := c.cfgManager.GetUserCfgs()
	m, err := convertForGet(configs)
	if err != nil {
		log.Errorf("failed to convert configurations: %v", err)
		c.SendInternalServerError(errors.New(""))
		return
	}

	c.Data["json"] = m
	c.ServeJSON()
}

// GetInternalConfig returns internal configurations
func (c *ConfigAPI) GetInternalConfig() {

	configs := c.cfgManager.GetAll()
	c.Data["json"] = configs
	c.ServeJSON()
}

// Put updates configurations
func (c *ConfigAPI) Put() {
	m := map[string]interface{}{}
	if err := c.DecodeJSONReq(&m); err != nil {
		c.SendBadRequestError(err)
		return
	}
	err := c.cfgManager.Load()
	if err != nil {
		log.Errorf("failed to get configurations: %v", err)
		c.SendInternalServerError(errors.New(""))
		return
	}
	isSysErr, err := c.validateCfg(m)
	if err != nil {
		if isSysErr {
			log.Errorf("failed to validate configurations: %v", err)
			c.SendInternalServerError(errors.New(""))
			return
		}

		c.SendBadRequestError(err)
		return

	}

	if err := c.cfgManager.UpdateConfig(m); err != nil {
		log.Errorf("failed to upload configurations: %v", err)
		c.SendInternalServerError(errors.New(""))
		return
	}
}

func (c *ConfigAPI) validateCfg(cfgs map[string]interface{}) (bool, error) {
	flag, err := authModeCanBeModified()
	if err != nil {
		return true, err
	}
	if !flag {
		if failedKeys := checkUnmodifiable(c.cfgManager, cfgs, common.AUTHMode); len(failedKeys) > 0 {
			return false, fmt.Errorf("the keys %v can not be modified as new users have been inserted into database", failedKeys)
		}
	}
	err = c.cfgManager.ValidateCfg(cfgs)
	return false, err
}

func checkUnmodifiable(mgr *config.CfgManager, cfgs map[string]interface{}, keys ...string) (failed []string) {
	if mgr == nil || cfgs == nil || keys == nil {
		return
	}
	for _, k := range keys {
		v := mgr.Get(k).GetString()
		if nv, ok := cfgs[k]; ok {
			if v != fmt.Sprintf("%v", nv) {
				failed = append(failed, k)
			}
		}
	}
	return
}

// delete sensitive attrs and add editable field to every attr
func convertForGet(cfg map[string]interface{}) (map[string]*value, error) {
	result := map[string]*value{}

	mList := metadata.Instance().GetAll()

	for _, item := range mList {
		if _, ok := item.ItemType.(*metadata.PasswordType); ok {
			delete(cfg, item.Name)
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
