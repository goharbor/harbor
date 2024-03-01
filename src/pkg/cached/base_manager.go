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

package cached

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/orm"
)

// innerCache is the default cache client,
// actually it is a wrapper for cache.LayerCache().
var innerCache cache.Cache = &cacheClient{}

// cacheClient is a interceptor for cache.CacheLayer, in order to implement specific
// case for cache layer.
type cacheClient struct{}

func (*cacheClient) Contains(ctx context.Context, key string) bool {
	return cache.LayerCache().Contains(ctx, key)
}

func (*cacheClient) Delete(ctx context.Context, key string) error {
	return cache.LayerCache().Delete(ctx, key)
}

func (*cacheClient) Fetch(ctx context.Context, key string, value interface{}) error {
	return cache.LayerCache().Fetch(ctx, key, value)
}

func (*cacheClient) Ping(ctx context.Context) error {
	return cache.LayerCache().Ping(ctx)
}

func (*cacheClient) Save(ctx context.Context, key string, value interface{}, expiration ...time.Duration) error {
	// intercept here
	// it should ignore save cache if this request is wrapped by orm.Transaction,
	// because if tx rollback, we can not rollback cache,
	// identify whether in transaction by checking the commitedKey in context.
	// commitedKey is a context value which be injected in the transaction middleware.
	if orm.HasCommittedKey(ctx) {
		return nil
	}

	return cache.LayerCache().Save(ctx, key, value, expiration...)
}

func (*cacheClient) Scan(ctx context.Context, match string) (cache.Iterator, error) {
	return cache.LayerCache().Scan(ctx, match)
}

var _ Manager = &BaseManager{}

// BaseManager is the base manager for cache and implement the cache manager interface.
type BaseManager struct {
	resourceType string
	cacheClient  cache.Cache
}

// NewBaseManager returns a instance of base manager.
func NewBaseManager(resourceType string) *BaseManager {
	return &BaseManager{
		resourceType: resourceType,
		cacheClient:  innerCache,
	}
}

// WithCacheClient can override the default cache client.
func (bm *BaseManager) WithCacheClient(cc cache.Cache) *BaseManager {
	bm.cacheClient = cc
	return bm
}

// CacheClient returns the cache client.
func (bm *BaseManager) CacheClient(_ context.Context) cache.Cache {
	return bm.cacheClient
}

// ResourceType returns the resource type.
func (bm *BaseManager) ResourceType(_ context.Context) string {
	return bm.resourceType
}

// CountCache returns current this resource occupied cache count.
func (bm *BaseManager) CountCache(ctx context.Context) (int64, error) {
	var count int64
	// prefix is resource type
	iter, err := bm.CacheClient(ctx).Scan(ctx, bm.ResourceType(ctx))
	if err != nil {
		return 0, err
	}

	for iter.Next(ctx) {
		count++
	}

	return count, nil
}

// DeleteCache deletes specific cache by key.
func (bm *BaseManager) DeleteCache(ctx context.Context, key string) error {
	return bm.CacheClient(ctx).Delete(ctx, key)
}

// FlushAll flush this resource's all cache.
func (bm *BaseManager) FlushAll(ctx context.Context) error {
	// prefix is resource type
	iter, err := bm.CacheClient(ctx).Scan(ctx, bm.ResourceType(ctx))
	if err != nil {
		return err
	}

	var errs errors.Errors
	for iter.Next(ctx) {
		if err = bm.CacheClient(ctx).Delete(ctx, iter.Val()); err != nil {
			errs = append(errs, err)
		}
	}

	if errs.Len() > 0 {
		return errs
	}

	return nil
}
