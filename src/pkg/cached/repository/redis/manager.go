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
	"github.com/goharbor/harbor/src/pkg/cached"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/pkg/repository/model"
)

var _ CachedManager = &manager{}

// CachedManager is the interface combines raw resource manager and cached manager for better extension.
type CachedManager interface {
	// Manager is the raw resource manager.
	repository.Manager
	// Manager is the common interface for resource cache.
	cached.Manager
}

// manager is the cached manager implemented by redis.
type manager struct {
	// delegator delegates the raw crud to DAO.
	delegator repository.Manager
	// client returns the redis cache client.
	client func() libcache.Cache
	// keyBuilder builds cache object key.
	keyBuilder *cached.ObjectKey
	// lifetime is the cache life time.
	lifetime time.Duration
}

// NewManager returns the redis cache manager.
func NewManager(m repository.Manager) *manager {
	return &manager{
		delegator:  m,
		client:     func() libcache.Cache { return libcache.Default() },
		keyBuilder: cached.NewObjectKey(cached.ResourceTypeRepository),
		lifetime:   time.Duration(config.CacheExpireHours()) * time.Hour,
	}
}

func (m *manager) Count(ctx context.Context, query *q.Query) (int64, error) {
	return m.delegator.Count(ctx, query)
}

func (m *manager) List(ctx context.Context, query *q.Query) ([]*model.RepoRecord, error) {
	return m.delegator.List(ctx, query)
}

func (m *manager) Create(ctx context.Context, repo *model.RepoRecord) (int64, error) {
	return m.delegator.Create(ctx, repo)
}

func (m *manager) NonEmptyRepos(ctx context.Context) ([]*model.RepoRecord, error) {
	return m.delegator.NonEmptyRepos(ctx)
}

func (m *manager) Get(ctx context.Context, id int64) (*model.RepoRecord, error) {
	key, err := m.keyBuilder.Format("id", id)
	if err != nil {
		return nil, err
	}

	repo := &model.RepoRecord{}
	if err = m.client().Fetch(ctx, key, repo); err == nil {
		return repo, nil
	}

	log.Debugf("get repository %d from cache error: %v, will query from database.", id, err)

	repo, err = m.delegator.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err = m.client().Save(ctx, key, repo, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save repository %s to cache error: %v", repo.Name, err)
	}

	return repo, nil
}

func (m *manager) GetByName(ctx context.Context, name string) (*model.RepoRecord, error) {
	key, err := m.keyBuilder.Format("name", name)
	if err != nil {
		return nil, err
	}

	repo := &model.RepoRecord{}
	if err = m.client().Fetch(ctx, key, repo); err == nil {
		return repo, nil
	}

	repo, err = m.delegator.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}

	if err = m.client().Save(ctx, key, repo, m.lifetime); err != nil {
		// log error if save to cache failed
		log.Debugf("save repository %s to cache error: %v", repo.Name, err)
	}

	return repo, nil
}

func (m *manager) Delete(ctx context.Context, id int64) error {
	repo, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	// pass on delete operation
	if err := m.delegator.Delete(ctx, id); err != nil {
		return err
	}
	// clean cache
	m.cleanUp(ctx, repo)
	return nil
}

func (m *manager) Update(ctx context.Context, repo *model.RepoRecord, props ...string) error {
	// pass on update operation
	if err := m.delegator.Update(ctx, repo, props...); err != nil {
		return err
	}
	// clean cache
	m.cleanUp(ctx, repo)
	return nil
}

func (m *manager) AddPullCount(ctx context.Context, id int64, count uint64) error {
	repo, err := m.Get(ctx, id)
	if err != nil {
		return err
	}
	// pass on update operation
	if err = m.delegator.AddPullCount(ctx, id, count); err != nil {
		return err
	}
	// refresh cache
	m.refreshCache(ctx, repo)
	return nil
}

// cleanUp cleans up data in cache.
func (m *manager) cleanUp(ctx context.Context, repo *model.RepoRecord) {
	// clean index by id
	idIdx, err := m.keyBuilder.Format("id", repo.RepositoryID)
	if err != nil {
		log.Errorf("format repository id key error: %v", err)
	} else {
		// retry to avoid dirty data
		if err = retry.Retry(func() error { return m.client().Delete(ctx, idIdx) }); err != nil {
			log.Errorf("delete repository cache key %s error: %v", idIdx, err)
		}
	}

	// clean index by name
	nameIdx, err := m.keyBuilder.Format("name", repo.Name)
	if err != nil {
		log.Errorf("format repository name key error: %v", err)
	} else {
		if err = retry.Retry(func() error { return m.client().Delete(ctx, nameIdx) }); err != nil {
			log.Errorf("delete repository cache key %s error: %v", nameIdx, err)
		}
	}
}

// refreshCache refreshes cache.
func (m *manager) refreshCache(ctx context.Context, repo *model.RepoRecord) {
	// refreshCache used for AddPullCount, because we have a background goroutine to
	// update per repo's pull_count in period time, in that case, we don't want to lose
	// cache every fixed interval, so prefer to use refreshCache instead of cleanUp.
	// no need to consider lock because we only have one goroutine do this work one by one.

	// refreshCache includes 2 steps:
	//   1. cleanUp
	//   2. re-get
	m.cleanUp(ctx, repo)

	var err error
	// re-get by id
	_, err = m.Get(ctx, repo.RepositoryID)
	if err != nil {
		log.Errorf("refresh cache by repository id %d error: %v", repo.RepositoryID, err)
	}
	// re-get by name
	_, err = m.GetByName(ctx, repo.Name)
	if err != nil {
		log.Errorf("refresh cache by repository name %s error: %v", repo.Name, err)
	}
}

func (m *manager) ResourceType(ctx context.Context) string {
	return cached.ResourceTypeRepository
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
