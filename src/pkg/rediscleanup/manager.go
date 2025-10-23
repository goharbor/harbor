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

package rediscleanup

import (
	"context"

	"github.com/go-redis/redis/v8"

	"github.com/goharbor/harbor/src/lib/log"
	libredis "github.com/goharbor/harbor/src/lib/redis"
)

var (
	// Mgr default redis cleanup manager
	Mgr = NewManager()
)

// Manager interface provide the cleanup functions for Redis cache
type Manager interface {
	// CleanupInvalidBlobSizeKeys removes all zero or negative blob size keys from Redis
	CleanupInvalidBlobSizeKeys(ctx context.Context) error
}

type manager struct{}

// NewManager returns redis cleanup manager
func NewManager() Manager {
	return &manager{}
}

// CleanupInvalidBlobSizeKeys removes all zero or negative blob size keys from Redis
func (m *manager) CleanupInvalidBlobSizeKeys(ctx context.Context) error {
	rc, err := libredis.GetRegistryClient()
	if err != nil {
		return err
	}

	// Scan for all upload:*:size keys
	uploadKeys, err := rc.Keys(ctx, "upload:*:size").Result()
	if err != nil {
		return err
	}

	// Scan for all blobs::* keys
	blobKeys, err := rc.Keys(ctx, "blobs::*").Result()
	if err != nil {
		return err
	}

	cleanedCount := 0

	// Clean up upload:*:size keys
	for _, key := range uploadKeys {
		size, err := rc.Get(ctx, key).Int64()
		if err != nil {
			if err == redis.Nil {
				continue // Key doesn't exist anymore, skip
			}
			log.Errorf("failed to get blob size for key %s during cleanup, error: %v", key, err)
			continue
		}

		// If we find a zero or negative size, delete it
		if size <= 0 {
			if err := rc.Del(ctx, key).Err(); err != nil {
				log.Errorf("failed to delete invalid blob size key %s during cleanup, error: %v", key, err)
			} else {
				cleanedCount++
				log.Infof("cleaned up invalid upload key %s with value %d", key, size)
			}
		}
	}

	// Clean up blobs::* keys with zero size
	for _, key := range blobKeys {
		size, err := rc.HGet(ctx, key, "size").Result()
		if err != nil {
			if err == redis.Nil {
				continue // Key doesn't exist anymore, skip
			}
			log.Errorf("failed to get blob size for key %s during cleanup, error: %v", key, err)
			continue
		}

		// If we find a zero size, delete the entire key
		if size == "0" {
			if err := rc.Del(ctx, key).Err(); err != nil {
				log.Errorf("failed to delete zero-sized blob key %s during cleanup, error: %v", key, err)
			} else {
				cleanedCount++
				log.Infof("cleaned up zero-sized blob key %s", key)
			}
		}
	}

	if cleanedCount > 0 {
		log.Infof("cleanup completed: removed %d invalid blob size keys from Redis", cleanedCount)
	}

	return nil
}
