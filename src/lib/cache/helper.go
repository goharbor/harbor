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

package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/log"
)

var (
	fetchOrSaveMu = keyMutex{m: &sync.Map{}}
)

// FetchOrSave retrieves the value for the key if present in the cache.
// Otherwise, it saves the value from the builder and retrieves the value for the key again.
func FetchOrSave(ctx context.Context, c Cache, key string, value interface{}, builder func() (interface{}, error), expiration ...time.Duration) error {
	err := c.Fetch(ctx, key, value)
	// value found from the cache
	if err == nil {
		return nil
	}
	// internal error
	if !errors.Is(err, ErrNotFound) {
		return err
	}

	// lock the key in cache and try to build the value for the key
	lockKey := fmt.Sprintf("%p:%s", c, key)
	fetchOrSaveMu.Lock(lockKey)

	defer fetchOrSaveMu.Unlock(lockKey)

	// fetch again to avoid to build the value multi-times
	err = c.Fetch(ctx, key, value)
	if err == nil {
		return nil
	}
	// internal error
	if !errors.Is(err, ErrNotFound) {
		return err
	}

	val, err := builder()
	if err != nil {
		return err
	}

	if err := c.Save(ctx, key, val, expiration...); err != nil {
		log.Warningf("failed to save value to cache, error: %v", err)

		// save the val to cache failed, copy it to the value using the default codec
		data, err := codec.Encode(val)
		if err != nil {
			return err
		}

		return codec.Decode(data, value)
	}

	return c.Fetch(ctx, key, value) // after the building, fetch value again
}
