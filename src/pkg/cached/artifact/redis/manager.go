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
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/cached"
)

var _ CachedManager = &manager{}

// CachedManager is the interface combines raw resource manager and cached manager for better extension.
type CachedManager interface {
	// Manager is the raw resource manager.
	artifact.Manager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// manager is the cached manager implemented by redis.
type manager struct {
	// delegator delegates the raw crud to DAO.
	delegator artifact.Manager
	// client returns the redis cache client.
	client func() libcache.Cache
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager(m artifact.Manager) *manager {
	return &manager{
		delegator:  m,
		client:     func() libcache.Cache { return libcache.Default() },
		keyBuilder: cached.NewObjectKey(cached.ResourceTypeArtifact),
		lifetime:   time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.delegator.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*artifact.Artifact, error) {
	return m.delegator.List(ctx, query)
}

func (m *manager) Create(ctx context.Context, artifact *artifact.Artifact) (int64, error) {
	return m.delegator.Create(ctx, artifact)
}

func (m *manager) ListReferences(ctx context.Context, query *q.Query) ([]*artifact.Reference, error) {
	return m.delegator.ListReferences(ctx, query)
}

func (m *manager) DeleteReference(ctx context.Context, id int64) error {
	return m.delegator.DeleteReference(ctx, id)
}

func (m *manager) Get(ctx context.Context, id int64) (*artifact.Artifact, error) {
	key, err := m.keyBuilder.Format("id", id)
	if err != nil {
		return nil, err
	}

	art := &artifact.Artifact{}
	if err = m.client().Fetch(ctx, key, art); err == nil {
		return art, nil
	}

	log.Debugf("get artifact %d from cache error: %v, will query from database.", id, err)

	art, err = m.delegator.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = m.client().Save(ctx, key, art, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save artifact %s to cache error: %v", art.String(), err)
	}

	return art, nil
}

func (m *manager) GetByDigest(ctx context.Context, repository, digest string) (*artifact.Artifact, error) {
	key, err := m.keyBuilder.Format("digest", digest)
	if err != nil {
		return nil, err
	}

	art := &artifact.Artifact{}
	if err = m.client().Fetch(ctx, key, art); err == nil {
		return art, nil
	}

	art, err = m.delegator.GetByDigest(ctx, repository, digest)
	if err != nil {
		return nil, err
	}

	if err = m.client().Save(ctx, key, art, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save artifact %s to cache error: %v", art.String(), err)
	}

	return art, nil
}

func (m *manager) Delete(ctx context.Context, id int64) error {
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

func (m *manager) Update(ctx context.Context, artifact *artifact.Artifact, props ...string) error {
	// pass on update operation
	if err := m.delegator.Update(ctx, artifact, props...); err != nil {
		return err
	}
	// clean cache
	m.cleanUp(ctx, artifact)
	return nil
}

func (m *manager) UpdatePullTime(ctx context.Context, id int64, pullTime time.Time) error {
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
func (m *manager) cleanUp(ctx context.Context, art *artifact.Artifact) {
	// clean index by id
	idIdx, err := m.keyBuilder.Format("id", art.ID)
	if err != nil {
		log.Errorf("format artifact id key error: %v", err)
	} else {
		// retry to avoid dirty data
		if err = retry.Retry(func() error { return m.client().Delete(ctx, idIdx) }); err != nil {
			log.Errorf("delete artifact cache key %s error: %v", idIdx, err)
		}
	}

	// clean index by digest
	digestIdx, err := m.keyBuilder.Format("digest", art.Digest)
	if err != nil {
		log.Errorf("format artifact digest key error: %v", err)
	} else {
		if err = retry.Retry(func() error { return m.client().Delete(ctx, digestIdx) }); err != nil {
			log.Errorf("delete artifact cache key %s error: %v", digestIdx, err)
		}
	}
}

// refreshCache refreshes cache.
func (m *manager) refreshCache(ctx context.Context, art *artifact.Artifact) {
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

func (m *manager) ResourceType(ctx context.Context) string {
	return cached.ResourceTypeArtifact
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
