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

package redis

import (
	"context"
	"time"

	libcache "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/cached"
)

var _ CachedManager = &manager{}

// ManifestManager is the manager for manifest.
type ManifestManager interface {
	// Save manifest to cache.
	Save(ctx context.Context, digest string, manifest []byte) error
	// Get manifest from cache.
	Get(ctx context.Context, digest string) ([]byte, error)
	// Delete manifest from cache.
	Delete(ctx context.Context, digest string) error
}

// CachedManager is the interface combines raw resource manager and cached manager for better extension.
type CachedManager interface {
	// ManifestManager is the manager for manifest.
	ManifestManager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// manager is the cached manager implemented by redis.
type manager struct {
	// client returns the redis cache client.
	client func() libcache.Cache
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager() *manager {
	return &manager{
		client:     func() libcache.Cache { return libcache.Default() },
		keyBuilder: cached.NewObjectKey(cached.ResourceTypeManifest),
		lifetime:   time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *manager) Save(ctx context.Context, digest string, manifest []byte) error {
	key, err := m.keyBuilder.Format("digest", digest)
	if err != nil {
		return err
	}

	return m.client().Save(ctx, key, manifest, m.lifetime)
}

func (m *manager) Get(ctx context.Context, digest string) ([]byte, error) {
	key, err := m.keyBuilder.Format("digest", digest)
	if err != nil {
		return nil, err
	}

	var manifest []byte
	if err = m.client().Fetch(ctx, key, &manifest); err == nil {
		return manifest, nil
	}

	return nil, err
}

func (m *manager) Delete(ctx context.Context, digest string) error {
	key, err := m.keyBuilder.Format("digest", digest)
	if err != nil {
		return err
	}

	return retry.Retry(func() error { return m.client().Delete(ctx, key) })
}

func (m *manager) ResourceType(ctx context.Context) string {
	return cached.ResourceTypeManifest
}

func (m *manager) CountCache(ctx context.Context) (int64, error) {
	// prefix is resource type
	keys, err := m.client().Keys(ctx, m.ResourceType(ctx))
	if err != nil {
		return 0, err
	}

	return int64(len(keys)), nil
}

func (m *manager) DeleteCache(ctx context.Context, key string) error {
	return m.client().Delete(ctx, key)
}

func (m *manager) FlushAll(ctx context.Context) error {
	// prefix is resource type
	keys, err := m.client().Keys(ctx, m.ResourceType(ctx))
	if err != nil {
		return err
	}

	var errs errors.Errors
	for _, key := range keys {
		if err = m.client().Delete(ctx, key); err != nil {
			errs = append(errs, err)
		}
	}

	if errs.Len() > 0 {
		return errs
	}

	return nil
}
