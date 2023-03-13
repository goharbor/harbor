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

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/cached"
)

var _ CachedManager = &Manager{}

// ManifestManager is the Manager for manifest.
type ManifestManager interface {
	// Save manifest to cache.
	Save(ctx context.Context, digest string, manifest []byte) error
	// Get manifest from cache.
	Get(ctx context.Context, digest string) ([]byte, error)
	// Delete manifest from cache.
	Delete(ctx context.Context, digest string) error
}

// CachedManager is the interface combines raw resource Manager and cached Manager for better extension.
type CachedManager interface {
	// ManifestManager is the Manager for manifest.
	ManifestManager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// Manager is the cached manager implemented by redis.
type Manager struct {
	*cached.BaseManager
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager() *Manager {
	return &Manager{
		BaseManager: cached.NewBaseManager(cached.ResourceTypeManifest),
		keyBuilder:  cached.NewObjectKey(cached.ResourceTypeManifest),
		lifetime:    time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *Manager) Save(ctx context.Context, digest string, manifest []byte) error {
	key, err := m.keyBuilder.Format("digest", digest)
	if err != nil {
		return err
	}

	return m.CacheClient(ctx).Save(ctx, key, manifest, m.lifetime)
}

func (m *Manager) Get(ctx context.Context, digest string) ([]byte, error) {
	key, err := m.keyBuilder.Format("digest", digest)
	if err != nil {
		return nil, err
	}

	var manifest []byte
	if err = m.CacheClient(ctx).Fetch(ctx, key, &manifest); err == nil {
		return manifest, nil
	}

	return nil, err
}

func (m *Manager) Delete(ctx context.Context, digest string) error {
	key, err := m.keyBuilder.Format("digest", digest)
	if err != nil {
		return err
	}

	return retry.Retry(func() error { return m.CacheClient(ctx).Delete(ctx, key) })
}
