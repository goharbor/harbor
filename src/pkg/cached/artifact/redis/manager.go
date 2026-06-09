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
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/lib/retry"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/cached"
)

// cleanupCacheTimeout bounds a single cache-eviction attempt after the
// enclosing transaction has committed. Cache invalidation is best-effort:
// any stale entry expires via the configured TTL, and the next read from
// the database repopulates it. We must not sit spinning in a retry loop
// for the full default timeout.
const cleanupCacheTimeout = 3 * time.Second

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
	// Defer cache invalidation until after the enclosing transaction
	// commits. Running Redis calls inside the transaction would keep the
	// Postgres row locks held across network round-trips; on a canceled
	// request ctx, the retry loop would spin for a full minute while the
	// transaction sits idle in ClientRead, blocking concurrent deletes.
	// When there is no enclosing transaction, AfterCommit runs the hook
	// synchronously on the caller's goroutine.
	m.scheduleCleanUp(ctx, art)
	return nil
}

func (m *Manager) Update(ctx context.Context, artifact *artifact.Artifact, props ...string) error {
	// pass on update operation
	if err := m.delegator.Update(ctx, artifact, props...); err != nil {
		return err
	}
	// Same rationale as Delete: cache eviction must not hold the tx open.
	m.scheduleCleanUp(ctx, artifact)
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

// scheduleCleanUp registers the cache invalidation for art to run after the
// enclosing transaction commits. The closure captures a value copy of the
// fields it needs so the hook is independent of the request context and any
// mutations to art after registration.
func (m *Manager) scheduleCleanUp(ctx context.Context, art *artifact.Artifact) {
	// Capture by value — the caller's art pointer may be reused or mutated.
	id := art.ID
	repo := art.RepositoryName
	digest := art.Digest
	orm.AfterCommit(ctx, func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupCacheTimeout)
		defer cancel()
		m.cleanUpKeys(cleanupCtx, id, repo, digest)
	})
}

// cleanUpKeys best-effort deletes the cache entries for a given artifact.
// Called after the enclosing transaction has committed with a detached
// context, so it neither holds row locks nor observes a canceled request.
func (m *Manager) cleanUpKeys(ctx context.Context, id int64, repo, digest string) {
	// clean index by id
	idIdx, err := m.keyBuilder.Format("id", id)
	if err != nil {
		log.Errorf("format artifact id key error: %v", err)
	} else {
		// retry to avoid dirty data; capped retry + ctx-aware so we cannot
		// block indefinitely on Redis hiccups.
		if err = retry.Retry(
			func() error { return m.CacheClient(ctx).Delete(ctx, idIdx) },
			retry.Context(ctx),
			retry.Timeout(cleanupCacheTimeout),
		); err != nil {
			log.Warningf("delete artifact cache key %s error: %v", idIdx, err)
		}
	}

	// clean index by digest
	digestIdx, err := m.keyBuilder.Format("repository", repo, "digest", digest)
	if err != nil {
		log.Errorf("format artifact digest key error: %v", err)
	} else {
		if err = retry.Retry(
			func() error { return m.CacheClient(ctx).Delete(ctx, digestIdx) },
			retry.Context(ctx),
			retry.Timeout(cleanupCacheTimeout),
		); err != nil {
			log.Warningf("delete artifact cache key %s error: %v", digestIdx, err)
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
	//   1. cleanUp keys synchronously (we're not holding a transaction here —
	//      the UpdatePullTime pathway is called from a background goroutine,
	//      not inside orm.WithTransaction)
	//   2. re-get
	cleanupCtx, cancel := context.WithTimeout(context.Background(), cleanupCacheTimeout)
	defer cancel()
	m.cleanUpKeys(cleanupCtx, art.ID, art.RepositoryName, art.Digest)

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
