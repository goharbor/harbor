//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package config

import (
	"context"
	"fmt"
	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/config/store"
	"github.com/goharbor/harbor/src/pkg/config/validate"
	"os"
)

// CfgManager ... Configure Manager
type CfgManager struct {
	Store *store.ConfigStore
}

var validateRules = []validate.Rule{
	&validate.LdapGroupValidateRule{},
}

// LoadDefault ...
func (c *CfgManager) LoadDefault() {
	// Init Default Value
	itemArray := metadata.Instance().GetAll()
	for _, item := range itemArray {
		if len(item.DefaultValue) > 0 {
			cfgValue, err := metadata.NewCfgValue(item.Name, item.DefaultValue)
			if err != nil {
				log.Errorf("LoadDefault failed, config item, key: %v,  err: %v", item.Name, err)
				continue
			}
			c.Store.Set(item.Name, *cfgValue)
		}
	}
}

// LoadSystemConfigFromEnv ...
func (c *CfgManager) LoadSystemConfigFromEnv() {
	itemArray := metadata.Instance().GetAll()
	// Init System Value
	for _, item := range itemArray {
		if item.Scope == metadata.SystemScope && len(item.EnvKey) > 0 {
			if envValue, ok := os.LookupEnv(item.EnvKey); ok {
				configValue, err := metadata.NewCfgValue(item.Name, envValue)
				if err != nil {
					log.Errorf("LoadSystemConfigFromEnv failed, config item, key: %v,  err: %v", item.Name, err)
					continue
				}
				c.Store.Set(item.Name, *configValue)
			}
		}
	}
}

// GetAll get all settings.
func (c *CfgManager) GetAll(ctx context.Context) map[string]interface{} {
	resultMap := map[string]interface{}{}
	if err := c.Store.Load(ctx); err != nil {
		log.Errorf("AllConfigs failed, error %v", err)
		return resultMap
	}
	metaDataList := metadata.Instance().GetAll()
	for _, item := range metaDataList {
		cfgValue, err := c.Store.GetAnyType(item.Name)
		if err != nil {
			if err != metadata.ErrValueNotSet {
				log.Errorf("Failed to get Value of key %v, error %v", item.Name, err)
			}
			continue
		}
		resultMap[item.Name] = cfgValue
	}
	return resultMap
}

// GetUserCfgs retrieve all user configs
func (c *CfgManager) GetUserCfgs(ctx context.Context) map[string]interface{} {
	resultMap := map[string]interface{}{}
	if err := c.Store.Load(ctx); err != nil {
		log.Errorf("UserConfigs failed, error %v", err)
		return resultMap
	}
	metaDataList := metadata.Instance().GetAll()
	for _, item := range metaDataList {
		if item.Scope == metadata.UserScope {
			cfgValue, err := c.Store.GetAnyType(item.Name)
			if err != nil {
				if err == metadata.ErrValueNotSet {
					if _, ok := item.ItemType.(*metadata.StringType); ok {
						cfgValue = ""
					}
					if _, ok := item.ItemType.(*metadata.NonEmptyStringType); ok {
						cfgValue = ""
					}
				} else {
					log.Errorf("Failed to get Value of key %v, error %v", item.Name, err)
					continue
				}
			}
			resultMap[item.Name] = cfgValue
		}
	}
	return resultMap
}

// Load load configuration from storage, like database or redis
func (c *CfgManager) Load(ctx context.Context) error {
	return c.Store.Load(ctx)
}

// Save - Save all current configuration to storage
func (c *CfgManager) Save(ctx context.Context) error {
	return c.Store.Save(ctx)
}

// Get ...
func (c *CfgManager) Get(ctx context.Context, key string) *metadata.ConfigureValue {
	configValue, err := c.Store.Get(key)
	if err != nil {
		log.Debugf("failed to get key %v, error: %v, maybe default value not defined before get", key, err)
		configValue = &metadata.ConfigureValue{}
	}
	return configValue
}

// Set ...
func (c *CfgManager) Set(ctx context.Context, key string, value interface{}) {
	configValue, err := metadata.NewCfgValue(key, utils.GetStrValueOfAnyType(value))
	if err != nil {
		log.Errorf("error when setting key: %v,  error %v", key, err)
		return
	}
	c.Store.Set(key, *configValue)
}

// GetDatabaseCfg - Get database configurations
func (c *CfgManager) GetDatabaseCfg() *models.Database {
	ctx := context.Background()
	return &models.Database{
		Type: c.Get(ctx, common.DatabaseType).GetString(),
		PostGreSQL: &models.PostGreSQL{
			Host:         c.Get(ctx, common.PostGreSQLHOST).GetString(),
			Port:         c.Get(ctx, common.PostGreSQLPort).GetInt(),
			Username:     c.Get(ctx, common.PostGreSQLUsername).GetString(),
			Password:     c.Get(ctx, common.PostGreSQLPassword).GetString(),
			Database:     c.Get(ctx, common.PostGreSQLDatabase).GetString(),
			SSLMode:      c.Get(ctx, common.PostGreSQLSSLMode).GetString(),
			MaxIdleConns: c.Get(ctx, common.PostGreSQLMaxIdleConns).GetInt(),
			MaxOpenConns: c.Get(ctx, common.PostGreSQLMaxOpenConns).GetInt(),
		},
	}
}

// UpdateConfig - Update config Store with a specified configuration and also save updated configure.
func (c *CfgManager) UpdateConfig(ctx context.Context, cfgs map[string]interface{}) error {
	return c.Store.Update(ctx, cfgs)
}

// ValidateCfg validate config by metadata. return the first error if exist.
func (c *CfgManager) ValidateCfg(ctx context.Context, cfgs map[string]interface{}) error {
	for key, value := range cfgs {
		item, exist := metadata.Instance().GetByName(key)
		if !exist {
			return errors.New(fmt.Sprintf("invalid config, item not defined in metadatalist, %v", key))
		}
		if item.Scope == metadata.SystemScope {
			return errors.New(fmt.Sprintf("system config items cannot be updated, item: %v", key))
		}
		strVal := utils.GetStrValueOfAnyType(value)
		_, err := metadata.NewCfgValue(key, strVal)
		if err != nil {
			return errors.Wrap(err, "item name "+key)
		}
	}

	for _, r := range validateRules {
		if err := r.Validate(ctx, c, cfgs); err != nil {
			return err
		}
	}
	return nil
}
