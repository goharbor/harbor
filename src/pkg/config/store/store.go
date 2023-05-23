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

// Package store is only used in the internal implement of manager, not a public api.
package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/log"
)

// ConfigStore - the config data store
type ConfigStore struct {
	cfgDriver Driver
	cfgValues sync.Map
}

// NewConfigStore create config store
func NewConfigStore(cfgDriver Driver) *ConfigStore {
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
func (c *ConfigStore) Load(ctx context.Context) error {
	if c.cfgDriver == nil {
		return errors.New("failed to load store, cfgDriver is nil")
	}
	cfgs, err := c.cfgDriver.Load(ctx)
	if err != nil {
		return err
	}
	for key, value := range cfgs {
		strValue, err := ToString(value)
		if err != nil {
			log.Errorf("failed to transform the value from driver to string, key: %s, value: %v, error: %v", key, value, err)
			continue
		}
		cfgValue := metadata.ConfigureValue{}
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
func (c *ConfigStore) Save(ctx context.Context) error {
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

	return c.cfgDriver.Save(ctx, cfgMap)
}

// Update - Only update specified settings in cfgMap in store and driver
func (c *ConfigStore) Update(ctx context.Context, cfgMap map[string]interface{}) error {
	// Update to store
	for key, value := range cfgMap {
		configValue, err := metadata.NewCfgValue(key, utils.GetStrValueOfAnyType(value))
		if err != nil {
			log.Warningf("error %v, skip to update configure item, key:%v ", err, key)
			delete(cfgMap, key)
			continue
		}
		if err := c.Set(key, *configValue); err != nil {
			log.Warningf("failed to update configure item, key=%s, error: %v", key, err)
			continue
		}
	}
	// Update to driver
	return c.cfgDriver.Save(ctx, cfgMap)
}

// ToString ...
func ToString(value interface{}) (string, error) {
	if value == nil {
		return "nil", nil
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Map, reflect.Array, reflect.Slice, reflect.Struct:
		d, err := json.Marshal(value)
		if err != nil {
			return "", err
		}
		return string(d), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.String:
		return value.(string), nil
	default:
		return fmt.Sprintf("%v", value), nil
	}
}
