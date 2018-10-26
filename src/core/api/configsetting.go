package api

import (
	"net/http"

	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/config/client/db"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// ConfigSettingAPI ...
type ConfigSettingAPI struct {
	BaseController
	ConfigClient config.Client
}

// Prepare validates the user
func (c *ConfigSettingAPI) Prepare() {
	c.BaseController.Prepare()
	if !c.SecurityCtx.IsAuthenticated() {
		c.HandleUnauthorized()
		return
	}
	if !c.SecurityCtx.IsSysAdmin() && !c.SecurityCtx.IsSolutionUser() {
		c.HandleForbidden(c.SecurityCtx.GetUsername())
		return
	}
	c.ConfigClient = db.NewDBConfigureStore()
}

// GetConfigByGroup ...
func (c *ConfigSettingAPI) GetConfigByGroup() {
	group := c.GetStringFromPath(":group")
	if len(group) == 0 {
		log.Error("failed to get group")
		c.CustomAbort(http.StatusInternalServerError, http.StatusText(http.StatusInternalServerError))
	}
	configList, err := c.ConfigClient.GetSettingByGroup(group)
	log.Errorf("Found items %v", len(configList))
	if err != nil {
		log.Error("failed to get setting by group %v", err)
	}
	c.Data["json"] = configList
	c.ServeJSON()
	return
}
