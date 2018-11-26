package db

import (
	"fmt"

	"github.com/goharbor/harbor/src/common/config"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// CoreConfigManager ... Wrap the configure driver to previous interface, used for remove adminserver container only
type CoreConfigManager struct {
	Driver *ConfigureDriver
}

// NewCoreConfigManager ...
func NewCoreConfigManager() *CoreConfigManager {
	return &CoreConfigManager{Driver: GetConfigureDriverInstance()}
}

// NewCoreConfigManagerFromArray ...
func NewCoreConfigManagerFromArray(items []config.Item) *CoreConfigManager {
	return &CoreConfigManager{Driver: NewDBConfigureStoreFromArray(items)}
}

// Load ...
func (ccm *CoreConfigManager) Load() (map[string]interface{}, error) {
	resultMap := map[string]interface{}{}
	ccm.Driver.Load()
	values, err := ccm.Driver.GetAllSettings()
	if err != nil {
		return resultMap, err
	}
	for _, item := range values {
		key := item.GetKey()
		itemMetaData, err := config.MetaData.GetConfigMetaData(key)
		if err != nil {
			log.Errorf("Can not get the metadata of current key:%v", key)
			continue
		}
		if itemMetaData.Type == config.StringType {
			resultMap[key] = item.GetString()
		} else if itemMetaData.Type == config.IntType {
			resultMap[key] = item.GetInt()
		} else if itemMetaData.Type == config.Int64Type {
			resultMap[key] = item.GetInt64()
		} else if itemMetaData.Type == config.BoolType {
			resultMap[key] = item.GetBool()
		} else if itemMetaData.Type == config.PasswordType {
			resultMap[key] = item.GetPassword()
		}
	}
	return resultMap, nil
}

// Get ... no cache temporary
func (ccm *CoreConfigManager) Get() (map[string]interface{}, error) {
	return ccm.Load()
}

// Upload ...
func (ccm *CoreConfigManager) Upload(cfgs map[string]interface{}) error {
	for key, value := range cfgs {
		err := ccm.Driver.UpdateConfigValue(key, fmt.Sprintf("%v", value))
		if err != nil {
			log.Errorf("Failed to update configure key %v, value %v, error %v", key, value, err)
		}
	}
	return ccm.Driver.Save()
}

// Reset ...
func (ccm *CoreConfigManager) Reset() error {
	ccm.Driver.Reset()
	return nil
}

// InitDatabaseAndConfigure - Initial database and configure
func InitDatabaseAndConfigure() {
	configManager := NewCoreConfigManager()
	database := configManager.Driver.GetDatabaseCfg()
	dao.PrepareDatabase(database)
	if err := configManager.Driver.Save(); err != nil {
		log.Fatalf("failed to save configuration: %v", err)
	}
	GetConfigureDriverInstance().Load()
}

// DumpConfigureTrace - Used to Dump config trace information
func DumpConfigureTrace() {
	cfgManager := NewCoreConfigManager()
	if cfg, err := cfgManager.Get(); err != nil {
		fmt.Printf("config settings,cfg:%v\n", cfg)
	} else {
		fmt.Printf("Failed to DumpConfigTrace, error: %v", err)
	}

}
