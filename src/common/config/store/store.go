// Package store is only used in the internal implement of manager, not a public api.
package store

import (
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/config/metadata"
	"github.com/goharbor/harbor/src/common/config/store/driver"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/log"
	"sync"
)

// ConfigStore - the config data store
type ConfigStore struct {
	cfgDriver driver.Driver
	cfgValues sync.Map
}

// NewConfigStore create config store
func NewConfigStore(cfgDriver driver.Driver) *ConfigStore {
	return &ConfigStore{cfgDriver: cfgDriver}
}

// Get - Get config data from current store
func (c *ConfigStore) Get(key string) (*metadata.ConfigureValue, error) {
	if value, ok := c.cfgValues.Load(key); ok {
		if result, ok := value.(metadata.ConfigureValue); ok {
			return &result, nil
		}
		return nil, errors.New("data in config store is not a ConfigureValue type")
	}
	return nil, metadata.ErrValueNotSet
}

// GetAnyType get interface{} type for config items
func (c *ConfigStore) GetAnyType(key string) (interface{}, error) {
	if value, ok := c.cfgValues.Load(key); ok {
		if result, ok := value.(metadata.ConfigureValue); ok {
			return result.GetAnyType()
		}
		return nil, errors.New("data in config store is not a ConfigureValue type")
	}
	return nil, metadata.ErrValueNotSet
}

// Set - Set configure value in store, not saved to config driver
func (c *ConfigStore) Set(key string, value metadata.ConfigureValue) error {
	c.cfgValues.Store(key, value)
	return nil
}

// Load - Load data from driver, all user config in the store will be refreshed
func (c *ConfigStore) Load() error {
	if c.cfgDriver == nil {
		return errors.New("failed to load store, cfgDriver is nil")
	}
	cfgs, err := c.cfgDriver.Load()
	if err != nil {
		return err
	}
	for key, value := range cfgs {
		cfgValue := metadata.ConfigureValue{}
		strValue := fmt.Sprintf("%v", value)
		err = cfgValue.Set(key, strValue)
		if err != nil {
			log.Errorf("error when loading data item, key %v, value %v, error %v", key, value, err)
			continue
		}
		c.cfgValues.Store(key, cfgValue)
	}
	return nil
}

// Save - Save all data in current store
func (c *ConfigStore) Save() error {
	cfgMap := map[string]interface{}{}
	c.cfgValues.Range(func(key, value interface{}) bool {
		keyStr := fmt.Sprintf("%v", key)
		if configValue, ok := value.(metadata.ConfigureValue); ok {
			valueStr := configValue.Value
			if _, ok := metadata.Instance().GetByName(keyStr); ok {
				cfgMap[keyStr] = valueStr
			} else {
				log.Errorf("failed to get metadata for key %v", keyStr)
			}
		}
		return true
	})

	if c.cfgDriver == nil {
		return errors.New("failed to save store, cfgDriver is nil")
	}

	return c.cfgDriver.Save(cfgMap)
}

// Update - Only update specified settings in cfgMap in store and driver
func (c *ConfigStore) Update(cfgMap map[string]interface{}) error {
	// Update to store
	for key, value := range cfgMap {
		configValue, err := metadata.NewCfgValue(key, utils.GetStrValueOfAnyType(value))
		if err != nil {
			log.Warningf("error %v, skip to update configure item, key:%v ", err, key)
			delete(cfgMap, key)
			continue
		}
		c.Set(key, *configValue)
	}
	// Update to driver
	return c.cfgDriver.Save(cfgMap)
}
