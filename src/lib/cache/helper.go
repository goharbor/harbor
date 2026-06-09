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
	"time"

	"github.com/goharbor/harbor/src/lib/log"
)

// FetchOrSave retrieves the value for the key if present in the cache.
// Otherwise, it saves the value from the builder and retrieves the value for the key again.
func FetchOrSave(ctx context.Context, c Cache, key string, value any, builder func() (any, error), expiration ...time.Duration) error {
	err := c.Fetch(ctx, key, value)
	// value found from the cache
	if err == nil {
		return nil
	}
	// internal error
	if !errors.Is(err, ErrNotFound) {
		return err
	}

	// Use singleflight to deduplicate concurrent builds for the same key: only the
	// first caller runs builder(), all concurrent callers share its result.
	groupKey := fmt.Sprintf("%p:%s", c, key)

	result, err, _ := fetchOrSaveGroup.Do(groupKey, func() (any, error) {
		val, err := builder()
		if err != nil {
			return nil, err
		}

		// Save with a non-cancelable context so a canceled request context (e.g. the
		// HTTP client disconnected) does not prevent the cache from being populated.
		saveCtx := context.WithoutCancel(ctx)
		if err := c.Save(saveCtx, key, val, expiration...); err != nil {
			log.Warningf("failed to save value to cache, error: %v", err)
		}

		return val, nil
	})
	if err != nil {
		return err
	}

	// Copy the shared result into the caller's value via the codec, so every caller
	// (the leader and all waiters) gets its own populated value.
	data, err := codec.Encode(result)
	if err != nil {
		return err
	}
	return codec.Decode(data, value)
}
