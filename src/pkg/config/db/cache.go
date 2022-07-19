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

package db

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/pkg/config/store"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/log"
)

const cacheKey = "cfgs"

// Cache - Used to load/save configuration with cache
type Cache struct {
	cache  cache.Cache
	driver store.Driver
}

// Load - load config from database, only user setting will be load from database.
func (d *Cache) Load(ctx context.Context) (map[string]interface{}, error) {
	f := func() (interface{}, error) {
		return d.driver.Load(ctx)
	}

	result := map[string]interface{}{}

	// let the cache expired after one minute
	// because the there no way to rollback the config items been saved when invalidate the cache failed
	if err := cache.FetchOrSave(ctx, d.cache, cacheKey, &result, f, time.Minute); err != nil {
		return nil, err
	}

	return result, nil
}

// Save - Only save user config items in the cfgs map
func (d *Cache) Save(ctx context.Context, cfg map[string]interface{}) error {
	if err := d.driver.Save(ctx, cfg); err != nil {
		return err
	}

	if err := d.cache.Delete(ctx, cacheKey); err != nil {
		log.Warningf("failed to invalidate the cache of the configurations immediately, error: %v", err)
	}

	return nil
}

// NewCacheDriver returns driver with cache
func NewCacheDriver(cache cache.Cache, driver store.Driver) store.Driver {
	return &Cache{
		cache:  cache,
		driver: driver,
	}
}
