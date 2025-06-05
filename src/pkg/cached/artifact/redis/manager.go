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
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/cached"
)

var _ CachedManager = &Manager{}

// CachedManager is the interface combines raw resource Manager and cached Manager for better extension.
type CachedManager interface {
	// Manager is the raw resource Manager.
	artifact.Manager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// Manager is the cached manager implemented by redis.
type Manager struct {
	*cached.BaseManager
	// delegator delegates the raw crud to DAO.
	delegator artifact.Manager
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager(m artifact.Manager) *Manager {
	return &Manager{
		BaseManager: cached.NewBaseManager(cached.ResourceTypeArtifact),
		delegator:   m,
		keyBuilder:  cached.NewObjectKey(cached.ResourceTypeArtifact),
		lifetime:    time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *Manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.delegator.Count(ctx, query)
}

func (m *Manager) List(ctx context.Context, query *q.Query) ([]*artifact.Artifact, error) {
	return m.delegator.List(ctx, query)
}

func (m *Manager) ListWithLatest(ctx context.Context, query *q.Query) ([]*artifact.Artifact, error) {
	return m.delegator.ListWithLatest(ctx, query)
}

func (m *Manager) Create(ctx context.Context, artifact *artifact.Artifact) (int64, error) {
	return m.delegator.Create(ctx, artifact)
}

func (m *Manager) ListReferences(ctx context.Context, query *q.Query) ([]*artifact.Reference, error) {
	return m.delegator.ListReferences(ctx, query)
}

func (m *Manager) DeleteReference(ctx context.Context, id int64) error {
	return m.delegator.DeleteReference(ctx, id)
}

func (m *Manager) Get(ctx context.Context, id int64) (*artifact.Artifact, error) {
	key, err := m.keyBuilder.Format("id", id)
	if err != nil {
		return nil, err
	}

	art := &artifact.Artifact{}
	if err = m.CacheClient(ctx).Fetch(ctx, key, art); err == nil {
		return art, nil
	}

	log.Debugf("get artifact %d from cache error: %v, will query from database.", id, err)

	art, err = m.delegator.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = m.CacheClient(ctx).Save(ctx, key, art, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save artifact %s to cache error: %v", art.String(), err)
	}

	return art, nil
}

func (m *Manager) GetByDigest(ctx context.Context, repository, digest string) (*artifact.Artifact, error) {
	key, err := m.keyBuilder.Format("repository", repository, "digest", digest)
	if err != nil {
		return nil, err
	}

	art := &artifact.Artifact{}
	if err = m.CacheClient(ctx).Fetch(ctx, key, art); err == nil {
		return art, nil
	}

	art, err = m.delegator.GetByDigest(ctx, repository, digest)
	if err != nil {
		return nil, err
	}

	if err = m.CacheClient(ctx).Save(ctx, key, art, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save artifact %s to cache error: %v", art.String(), err)
	}

	return art, nil
}

func (m *Manager) Delete(ctx context.Context, id int64) error {
	art, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	// pass on delete operation
	if err := m.delegator.Delete(ctx, id); err != nil {
		return err
	}
	// clean cache
	m.cleanUp(ctx, art)
	return nil
}

func (m *Manager) Update(ctx context.Context, artifact *artifact.Artifact, props ...string) error {
	// pass on update operation
	if err := m.delegator.Update(ctx, artifact, props...); err != nil {
		return err
	}
	// clean cache
	m.cleanUp(ctx, artifact)
	return nil
}

func (m *Manager) UpdatePullTime(ctx context.Context, id int64, pullTime time.Time) error {
	art, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	// pass on updatePullTime operation
	if err = m.delegator.UpdatePullTime(ctx, id, pullTime); err != nil {
		return err
	}
	// refresh cache
	m.refreshCache(ctx, art)
	return nil
}

// cleanUp cleans up data in cache.
func (m *Manager) cleanUp(ctx context.Context, art *artifact.Artifact) {
	// clean index by id
	idIdx, err := m.keyBuilder.Format("id", art.ID)
	if err != nil {
		log.Errorf("format artifact id key error: %v", err)
	} else {
		// retry to avoid dirty data
		if err = retry.Retry(func() error { return m.CacheClient(ctx).Delete(ctx, idIdx) }); err != nil {
			log.Errorf("delete artifact cache key %s error: %v", idIdx, err)
		}
	}

	// clean index by digest
	digestIdx, err := m.keyBuilder.Format("repository", art.RepositoryName, "digest", art.Digest)
	if err != nil {
		log.Errorf("format artifact digest key error: %v", err)
	} else {
		if err = retry.Retry(func() error { return m.CacheClient(ctx).Delete(ctx, digestIdx) }); err != nil {
			log.Errorf("delete artifact cache key %s error: %v", digestIdx, err)
		}
	}
}

// refreshCache refreshes cache.
func (m *Manager) refreshCache(ctx context.Context, art *artifact.Artifact) {
	// refreshCache used for UpdatePullTime, because we have a background goroutine to
	// update per artifact pull_time in period time, in that case, we don't want to lose
	// cache every fixed interval, so prefer to use refreshCache instead of cleanUp.
	// no need to consider lock because we only have one goroutine do this work one by one.

	// refreshCache includes 2 steps:
	//   1. cleanUp
	//   2. re-get
	m.cleanUp(ctx, art)

	var err error
	// re-get by id
	_, err = m.Get(ctx, art.ID)
	if err != nil {
		log.Errorf("refresh cache by artifact id %d error: %v", art.ID, err)
	}
	// re-get by digest
	_, err = m.GetByDigest(ctx, art.RepositoryName, art.Digest)
	if err != nil {
		log.Errorf("refresh cache by artifact digest %s error: %v", art.Digest, err)
	}
}
