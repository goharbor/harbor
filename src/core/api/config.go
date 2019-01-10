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
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/config/metadata"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security/secret"
	"github.com/goharbor/harbor/src/common/utils/log"
	corecfg "github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/filter"
	"net/http"
	"strings"
)

// ConfigAPI ...
type ConfigAPI struct {
	BaseController
	cfgManager *config.CfgManager
}

// Prepare validates the user
func (c *ConfigAPI) Prepare() {
	c.BaseController.Prepare()
	c.cfgManager = config.NewDBCfgManager()
	if !c.SecurityCtx.IsAuthenticated() {
		c.HandleUnauthorized()
		return
	}

	// Only internal container can access /api/internal/configurations
	if strings.EqualFold(c.Ctx.Request.RequestURI, "/api/internal/configurations") {
		if _, ok := c.Ctx.Request.Context().Value(filter.SecurCtxKey).(*secret.SecurityContext); !ok {
			c.HandleUnauthorized()
			return
		}
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
	configs := c.cfgManager.GetUserCfgs()
	log.Infof("current configs %+v", configs)
	m, err := convertForGet(configs)
	if err != nil {
		log.Errorf("failed to convert configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
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
	c.DecodeJSONReq(&m)
	err := c.cfgManager.Load()
	isSysErr := false
	if err != nil {
		isSysErr = true
	}
	err = c.cfgManager.ValidateCfg(m)
	if err != nil {
		if isSysErr {
			log.Errorf("failed to validate configurations: %v", err)
			c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
		}

		c.CustomAbort(http.StatusBadRequest, err.Error())
	}

	if err := c.cfgManager.UpdateConfig(m); err != nil {
		log.Errorf("failed to upload configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	if err := c.cfgManager.Load(); err != nil {
		log.Errorf("failed to load configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}

	// Everything is ok, detect the configurations to confirm if the option we are caring is changed.
	if err := watchConfigChanges(m); err != nil {
		log.Errorf("Failed to watch configuration change with error: %s\n", err)
	}
}

// Reset system configurations
func (c *ConfigAPI) Reset() {
	if err := corecfg.Reset(); err != nil {
		log.Errorf("failed to reset configurations: %v", err)
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
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
