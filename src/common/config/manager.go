// Copyright Project Harbor Authors
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

package config

import (
	"fmt"
	"os"
	"sync"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/config/metadata"
	"github.com/goharbor/harbor/src/common/config/store"
	"github.com/goharbor/harbor/src/common/config/store/driver"
	"github.com/goharbor/harbor/src/common/http/modifier/auth"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// CfgManager ... Configure Manager
type CfgManager struct {
	store *store.ConfigStore
}

// NewDBCfgManager - create DB config manager
func NewDBCfgManager() *CfgManager {
	manager := &CfgManager{store: store.NewConfigStore(&driver.Database{})}
	// load default value
	manager.loadDefault()
	// load system config from env
	manager.loadSystemConfigFromEnv()
	return manager
}

// NewRESTCfgManager - create REST config manager
func NewRESTCfgManager(configURL, secret string) *CfgManager {
	secAuth := auth.NewSecretAuthorizer(secret)
	manager := &CfgManager{store: store.NewConfigStore(driver.NewRESTDriver(configURL, secAuth))}
	return manager
}

// InMemoryDriver driver for unit testing
type InMemoryDriver struct {
	sync.Mutex
	cfgMap map[string]interface{}
}

// Load load data from driver, for example load from database,
// it should be invoked before get any user scope config
// for system scope config, because it is immutable, no need to call this method
func (d *InMemoryDriver) Load() (map[string]interface{}, error) {
	d.Lock()
	defer d.Unlock()
	res := make(map[string]interface{})
	for k, v := range d.cfgMap {
		res[k] = v
	}
	return res, nil
}

// Save only save user config setting to driver, for example: database, REST
func (d *InMemoryDriver) Save(cfg map[string]interface{}) error {
	d.Lock()
	defer d.Unlock()
	for k, v := range cfg {
		d.cfgMap[k] = v
	}
	return nil
}

// NewInMemoryManager create a manager for unit testing, doesn't involve database or REST
func NewInMemoryManager() *CfgManager {
	manager := &CfgManager{store: store.NewConfigStore(&InMemoryDriver{cfgMap: map[string]interface{}{}})}
	// load default value
	manager.loadDefault()
	// load system config from env
	manager.loadSystemConfigFromEnv()
	return manager
}

// loadDefault ...
func (c *CfgManager) loadDefault() {
	// Init Default Value
	itemArray := metadata.Instance().GetAll()
	for _, item := range itemArray {
		if len(item.DefaultValue) > 0 {
			cfgValue, err := metadata.NewCfgValue(item.Name, item.DefaultValue)
			if err != nil {
				log.Errorf("loadDefault failed, config item, key: %v,  err: %v", item.Name, err)
				continue
			}
			c.store.Set(item.Name, *cfgValue)
		}
	}
}

// loadSystemConfigFromEnv ...
func (c *CfgManager) loadSystemConfigFromEnv() {
	itemArray := metadata.Instance().GetAll()
	// Init System Value
	for _, item := range itemArray {
		if item.Scope == metadata.SystemScope && len(item.EnvKey) > 0 {
			if envValue, ok := os.LookupEnv(item.EnvKey); ok {
				configValue, err := metadata.NewCfgValue(item.Name, envValue)
				if err != nil {
					log.Errorf("loadSystemConfigFromEnv failed, config item, key: %v,  err: %v", item.Name, err)
					continue
				}
				c.store.Set(item.Name, *configValue)
			}
		}
	}
}

// GetAll get all settings.
func (c *CfgManager) GetAll() map[string]interface{} {
	resultMap := map[string]interface{}{}
	if err := c.store.Load(); err != nil {
		log.Errorf("GetAll failed, error %v", err)
		return resultMap
	}
	metaDataList := metadata.Instance().GetAll()
	for _, item := range metaDataList {
		cfgValue, err := c.store.GetAnyType(item.Name)
		if err != nil {
			if err != metadata.ErrValueNotSet {
				log.Errorf("Failed to get value of key %v, error %v", item.Name, err)
			}
			continue
		}
		resultMap[item.Name] = cfgValue
	}
	return resultMap
}

// GetUserCfgs retrieve all user configs
func (c *CfgManager) GetUserCfgs() map[string]interface{} {
	resultMap := map[string]interface{}{}
	if err := c.store.Load(); err != nil {
		log.Errorf("GetUserCfgs failed, error %v", err)
		return resultMap
	}
	metaDataList := metadata.Instance().GetAll()
	for _, item := range metaDataList {
		if item.Scope == metadata.UserScope {
			cfgValue, err := c.store.GetAnyType(item.Name)
			if err != nil {
				if err == metadata.ErrValueNotSet {
					if _, ok := item.ItemType.(*metadata.StringType); ok {
						cfgValue = ""
					}
					if _, ok := item.ItemType.(*metadata.NonEmptyStringType); ok {
						cfgValue = ""
					}
				} else {
					log.Errorf("Failed to get value of key %v, error %v", item.Name, err)
					continue
				}
			}
			resultMap[item.Name] = cfgValue
		}
	}
	return resultMap
}

// Load load configuration from storage, like database or redis
func (c *CfgManager) Load() error {
	return c.store.Load()
}

// Save - Save all current configuration to storage
func (c *CfgManager) Save() error {
	return c.store.Save()
}

// Get ...
func (c *CfgManager) Get(key string) *metadata.ConfigureValue {
	configValue, err := c.store.Get(key)
	if err != nil {
		log.Errorf("failed to get key %v, error: %v", key, err)
		configValue = &metadata.ConfigureValue{}
	}
	return configValue
}

// Set ...
func (c *CfgManager) Set(key string, value interface{}) {
	configValue, err := metadata.NewCfgValue(key, utils.GetStrValueOfAnyType(value))
	if err != nil {
		log.Errorf("error when setting key: %v,  error %v", key, err)
		return
	}
	c.store.Set(key, *configValue)
}

// GetDatabaseCfg - Get database configurations
func (c *CfgManager) GetDatabaseCfg() *models.Database {
	return &models.Database{
		Type: c.Get(common.DatabaseType).GetString(),
		PostGreSQL: &models.PostGreSQL{
			Host:         c.Get(common.PostGreSQLHOST).GetString(),
			Port:         c.Get(common.PostGreSQLPort).GetInt(),
			Username:     c.Get(common.PostGreSQLUsername).GetString(),
			Password:     c.Get(common.PostGreSQLPassword).GetString(),
			Database:     c.Get(common.PostGreSQLDatabase).GetString(),
			SSLMode:      c.Get(common.PostGreSQLSSLMode).GetString(),
			MaxIdleConns: c.Get(common.PostGreSQLMaxIdleConns).GetInt(),
			MaxOpenConns: c.Get(common.PostGreSQLMaxOpenConns).GetInt(),
		},
	}
}

// UpdateConfig - Update config store with a specified configuration and also save updated configure.
func (c *CfgManager) UpdateConfig(cfgs map[string]interface{}) error {
	return c.store.Update(cfgs)
}

// ValidateCfg validate config by metadata. return the first error if exist.
func (c *CfgManager) ValidateCfg(cfgs map[string]interface{}) error {
	for key, value := range cfgs {
		strVal := utils.GetStrValueOfAnyType(value)
		_, err := metadata.NewCfgValue(key, strVal)
		if err != nil {
			return fmt.Errorf("%v, item name: %v", err, key)
		}
	}
	return nil
}

// DumpTrace dump all configurations
func (c *CfgManager) DumpTrace() {
	cfgs := c.GetAll()
	for k, v := range cfgs {
		log.Info(k, ":=", v)
	}
}
