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
	"maps"
	"reflect"
	"strconv"
	"sync"
	"sync/atomic"

	beegoorm "github.com/beego/beego/v2/client/orm"

	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/lib/config/metadata"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
)

type valueMap = map[string]metadata.ConfigureValue

// ConfigStore - the config data store
//
// The value map is replaced as a whole so readers never mix old and new settings.
type ConfigStore struct {
	cfgDriver Driver
	mu        sync.Mutex // serializes writers
	cfgValues atomic.Pointer[valueMap]
	// only grows, so a slow Load cannot replace newer values
	mergedRevision atomic.Uint64
	// a Save forces the next Load to merge again, undoing a failed or rolled-back save
	localWrites  atomic.Uint64
	mergedWrites atomic.Uint64
	// a Load that read the driver while an Update was published may hold older values
	publishedUpdates atomic.Uint64
}

// NewConfigStore create config store
func NewConfigStore(cfgDriver Driver) *ConfigStore {
	return &ConfigStore{cfgDriver: cfgDriver}
}

func (c *ConfigStore) values() valueMap {
	if m := c.cfgValues.Load(); m != nil {
		return *m
	}
	return nil
}

// Callers must hold c.mu.
func (c *ConfigStore) update(fn func(next valueMap)) {
	next := maps.Clone(c.values())
	if next == nil {
		next = valueMap{}
	}
	fn(next)
	c.cfgValues.Store(&next)
}

// Get - Get config data from current store
func (c *ConfigStore) Get(key string) (*metadata.ConfigureValue, error) {
	if value, ok := c.values()[key]; ok {
		return &value, nil
	}
	return nil, metadata.ErrValueNotSet
}

// GetFromDriver ...
func (c *ConfigStore) GetFromDriver(ctx context.Context, key string) (map[string]any, error) {
	if c.cfgDriver == nil {
		return nil, errors.New("failed to load store, cfgDriver is nil")
	}
	cfgs, err := c.cfgDriver.Get(ctx, key)
	if err != nil {
		return nil, err
	}
	return cfgs, nil
}

// GetAnyType get any type for config items
func (c *ConfigStore) GetAnyType(key string) (any, error) {
	if value, ok := c.values()[key]; ok {
		return value.GetAnyType()
	}
	return nil, metadata.ErrValueNotSet
}

// Set - Set configure value in store, not saved to config driver
func (c *ConfigStore) Set(key string, value metadata.ConfigureValue) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.update(func(next valueMap) { next[key] = value })
	return nil
}

// Load - Load data from driver, all user config in the store will be refreshed
func (c *ConfigStore) Load(ctx context.Context) error {
	if c.cfgDriver == nil {
		return errors.New("failed to load store, cfgDriver is nil")
	}
	var revision uint64
	if r, ok := c.cfgDriver.(Revisioned); ok {
		revision = r.Revision()
	}
	writes := c.localWrites.Load()
	if revision != 0 && revision == c.mergedRevision.Load() && writes == c.mergedWrites.Load() {
		return nil
	}
	updates := c.publishedUpdates.Load()
	cfgs, err := c.cfgDriver.Load(ctx)
	if err != nil {
		return err
	}
	loaded := make(valueMap, len(cfgs))
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
		loaded[key] = cfgValue
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if revision != 0 && revision < c.mergedRevision.Load() {
		// a concurrent Load already published newer values
		return nil
	}
	if updates != c.publishedUpdates.Load() {
		// the next Load merges again
		return nil
	}
	c.update(func(next valueMap) { maps.Copy(next, loaded) })
	c.mergedRevision.Store(revision)
	c.mergedWrites.Store(writes)
	return nil
}

// Save - Save all data in current store
func (c *ConfigStore) Save(ctx context.Context) error {
	cfgMap := map[string]any{}
	for keyStr, configValue := range c.values() {
		if _, ok := metadata.Instance().GetByName(keyStr); ok {
			cfgMap[keyStr] = configValue.Value
		} else {
			log.Errorf("failed to get metadata for key %v", keyStr)
		}
	}

	if c.cfgDriver == nil {
		return errors.New("failed to save store, cfgDriver is nil")
	}

	err := c.cfgDriver.Save(ctx, cfgMap)
	c.localWrites.Add(1)
	return err
}

// Update - Only update specified settings in cfgMap in store and driver
//
// Inside a transaction the values arrive through the commit announcement so readers
// never see rolled-back settings.
func (c *ConfigStore) Update(ctx context.Context, cfgMap map[string]any) error {
	updated := valueMap{}
	for key, value := range cfgMap {
		configValue, err := metadata.NewCfgValue(key, utils.GetStrValueOfAnyType(value))
		if err != nil {
			log.Warningf("error %v, skip to update configure item, key:%v ", err, key)
			delete(cfgMap, key)
			continue
		}
		updated[key] = *configValue
	}
	if err := c.cfgDriver.Save(ctx, cfgMap); err != nil {
		return err
	}
	if inTransaction(ctx) {
		return nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.update(func(next valueMap) { maps.Copy(next, updated) })
	c.publishedUpdates.Add(1)
	return nil
}

func inTransaction(ctx context.Context) bool {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return false
	}
	_, ok := o.(beegoorm.TxOrmer)
	return ok
}

// ToString ...
func ToString(value any) (string, error) {
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
