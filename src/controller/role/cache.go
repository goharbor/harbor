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

package role

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
)

const (
	// defaultL1MemoryTTL is how long an entry in the process-local (L1) cache is
	// trusted before it is re-validated against L2/DB. It is the dominant control
	// on freshness: a change made on another node is picked up within this window
	// (the shared L2 entry is deleted on write, so the next L1 refresh re-reads it).
	defaultL1MemoryTTL = time.Second
	// defaultL2RedisTTL is the disabled sentinel: Redis (L2) is OFF by default.
	// RBAC is a per-request hot path, so we keep Redis off it unless an operator
	// explicitly opts in via ROLE_CACHE_L2_REDIS_TTL — avoiding a new Redis
	// availability/latency dependency on authz (see goharbor/harbor#23335, where a
	// per-request Redis cache path exhausted the DB pool and hung core). L1's
	// in-memory TTL bounds staleness without Redis. When enabled, 30m is the
	// suggested value; it then doubles as the backstop for out-of-band DB changes
	// (e.g. a direct SQL UPDATE to role_permission that bypasses the controller).
	defaultL2RedisTTL = time.Duration(-1)

	// Env vars (see envDuration for the accepted format and the <=0 "disable" rule).
	envL1MemoryTTL = "ROLE_CACHE_L1_MEMORY_TTL"
	envL2RedisTTL  = "ROLE_CACHE_L2_REDIS_TTL"
)

// errRoleNotFound signals that the inner controller has no such role. It is used
// only to keep a missing role out of the cache (FetchOrSave must not persist a nil
// value); callers translate it back to the (nil, nil) not-found contract.
var errRoleNotFound = errors.New("role not found")

// l1Entry is the process-local (L1) cache value: the role and the absolute
// instant after which it must be re-validated against L2/DB.
type l1Entry struct {
	role      *Role
	expiresAt time.Time
}

// cachingController decorates a role Controller with a two-level, staleness-bounded
// permission cache:
//
//	L1 (memory, fresh within l1MemoryTTL) -> L2 (Redis, TTL l2RedisTTL) -> inner (DB)
//
// It implements Controller and embeds the inner controller, so Count/List pass
// straight through; only Get (with permissions) is served from the cache, and
// Create/Update/Delete invalidate it after the inner write succeeds. Either layer
// is disabled by setting its TTL <= 0 (env value -1); with both disabled every
// read goes straight to the inner controller.
type cachingController struct {
	Controller // embedded inner (DB) controller — provides Count/List pass-through

	local       sync.Map      // id -> *l1Entry
	l1MemoryTTL time.Duration // <=0 => L1 disabled
	l2RedisTTL  time.Duration // <=0 => Redis (L2) disabled

	// redisCache is the injected L2 backend; tests set a fake here. Nil in prod,
	// where redis() uses the shared cache.Default().
	redisCache cache.Cache

	warmOnce sync.Once
}

// newCachingController wraps inner with a cache configured from the environment
// (read once at process start, consistent with the rest of Harbor's env config).
func newCachingController(inner Controller) *cachingController {
	return newCachingControllerWith(
		inner,
		envDuration(envL1MemoryTTL, defaultL1MemoryTTL),
		envDuration(envL2RedisTTL, defaultL2RedisTTL),
	)
}

// newCachingControllerWith wraps inner with explicit timings, so tests can
// exercise the layers without depending on import-time environment.
func newCachingControllerWith(inner Controller, l1MemoryTTL, l2RedisTTL time.Duration) *cachingController {
	return &cachingController{
		Controller:  inner,
		l1MemoryTTL: l1MemoryTTL,
		l2RedisTTL:  l2RedisTTL,
	}
}

// envDuration reads name as an integer number of seconds, or a Go duration
// string (e.g. "30m"). A value <= 0 means "disabled" and is returned as -1.
// On a parse error the default is used. This mirrors the convention already used
// by the Redis cache options parser (src/lib/cache/redis/util.go).
func envDuration(name string, def time.Duration) time.Duration {
	v := os.Getenv(name)
	if v == "" {
		return def
	}
	if i, err := strconv.Atoi(v); err == nil {
		if i <= 0 {
			return -1 // disabled
		}
		return time.Duration(i) * time.Second
	}
	if d, err := time.ParseDuration(v); err == nil {
		if d <= 0 {
			return -1 // disabled
		}
		return d
	}
	log.Errorf("failed to parse %s=%q, using default value %s", name, v, def)
	return def
}

func (c *cachingController) l1Enabled() bool { return c.l1MemoryTTL > 0 }
func (c *cachingController) l2Enabled() bool { return c.l2RedisTTL > 0 }

// redis returns the L2 backend, or nil when L2 is disabled: the injected
// redisCache if set (tests), else the shared cache.Default(). The default is
// resolved lazily (not stored at construction) because role.Ctl is built at
// import time, before cache.Initialize runs at startup.
func (c *cachingController) redis() cache.Cache {
	if !c.l2Enabled() {
		return nil
	}
	if c.redisCache != nil {
		return c.redisCache
	}
	return cache.Default()
}

func roleCacheKey(id int64) string {
	return fmt.Sprintf("role:%d", id)
}

// Get serves a populated role from L1 -> L2 -> inner (DB). Reads without
// permissions are not cacheable and go straight to the inner controller.
func (c *cachingController) Get(ctx context.Context, id int64, option *Option) (*Role, error) {
	if option == nil || !option.WithPermission {
		return c.Controller.Get(ctx, id, option)
	}
	c.maybeWarm()

	// L1: process-local, valid only within its TTL window.
	if c.l1Enabled() {
		if v, ok := c.local.Load(id); ok {
			if e := v.(*l1Entry); time.Now().Before(e.expiresAt) {
				// return a copy so callers cannot mutate the cached role
				return e.role.clone(), nil
			}
		}
	}

	// build reads the source of truth (DB). A missing role is reported as
	// errRoleNotFound so it is never written to the cache.
	build := func() (*Role, error) {
		r, err := c.Controller.Get(ctx, id, option)
		if err != nil {
			return nil, err
		}
		if r == nil {
			return nil, errRoleNotFound
		}
		return r, nil
	}

	// L2: Redis. FetchOrSave coalesces concurrent L1 misses for the same key
	// (singleflight => a single DB build) and populates Redis with a
	// non-cancelable context (l2RedisTTL). It runs only after an L1 miss, so the
	// per-request hot path stays L1-only. On any L2/Redis error we fall through
	// to a direct DB read, so a Redis hiccup never fails an authz check.
	if rc := c.redis(); rc != nil {
		var r Role
		err := cache.FetchOrSave(ctx, rc, roleCacheKey(id), &r, func() (any, error) {
			v, err := build()
			if err != nil {
				return nil, err
			}
			return v, nil
		}, c.l2RedisTTL)
		switch {
		case errors.Is(err, errRoleNotFound):
			return nil, nil
		case err == nil:
			c.storeL1(id, &r) // promote to L1
			return (&r).clone(), nil
		default:
			log.Warningf("role cache L2 path failed for id %d, falling back to DB: %v", id, err)
		}
	}

	// L2 disabled (or L2 errored above): read the DB directly and populate L1.
	r, err := build()
	if errors.Is(err, errRoleNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c.storeL1(id, r)
	return r.clone(), nil
}

// Create delegates to the inner controller, then invalidates the cache.
func (c *cachingController) Create(ctx context.Context, r *Role) (int64, error) {
	id, err := c.Controller.Create(ctx, r)
	if err != nil {
		return 0, err
	}
	c.invalidate(ctx, id)
	return id, nil
}

// Update delegates to the inner controller, then invalidates the cache.
func (c *cachingController) Update(ctx context.Context, r *Role, option *Option) error {
	if err := c.Controller.Update(ctx, r, option); err != nil {
		return err
	}
	// A successful inner Update guarantees a non-nil role with a valid ID (the
	// inner controller rejects a nil role). Guard defensively, but log rather than
	// silently skip: a missed invalidation would serve stale permissions.
	if r == nil {
		log.Warningf("role cache: Update succeeded with a nil role; cannot invalidate, permissions may be stale")
		return nil
	}
	c.invalidate(ctx, r.ID)
	return nil
}

// Delete delegates to the inner controller, then invalidates the cache.
func (c *cachingController) Delete(ctx context.Context, id int64, option ...*Option) error {
	if err := c.Controller.Delete(ctx, id, option...); err != nil {
		return err
	}
	c.invalidate(ctx, id)
	return nil
}

// invalidate drops the changed role from both layers so the next reader re-reads
// the DB. Other nodes refresh within l1MemoryTTL.
func (c *cachingController) invalidate(ctx context.Context, id int64) {
	if rc := c.redis(); rc != nil {
		_ = rc.Delete(ctx, roleCacheKey(id))
	}
	c.local.Delete(id)
}

// storeL1 caches a role in the process-local L1 with an absolute expiry. No-op
// when L1 is disabled.
func (c *cachingController) storeL1(id int64, r *Role) {
	if !c.l1Enabled() || r == nil {
		return
	}
	c.local.Store(id, &l1Entry{role: r, expiresAt: time.Now().Add(c.l1MemoryTTL)})
}

// maybeWarm fires the background warm once, if there is a cache layer to warm into.
func (c *cachingController) maybeWarm() {
	if !c.l1Enabled() && !c.l2Enabled() {
		return
	}
	c.warmOnce.Do(func() {
		go func() {
			defer func() {
				if rec := recover(); rec != nil {
					log.Warningf("role permission cache warm aborted: %v", rec)
				}
			}()
			c.warm(orm.Context())
		}()
	})
}

// warm loads all roles with their permissions (via the inner controller) into the
// enabled cache layers.
func (c *cachingController) warm(ctx context.Context) {
	roles, err := c.Controller.List(ctx, &q.Query{PageSize: -1}, &Option{WithPermission: true})
	if err != nil {
		log.Warningf("failed to warm role permission cache: %v", err)
		return
	}
	for _, r := range roles {
		if r == nil {
			continue
		}
		if rc := c.redis(); rc != nil {
			_ = rc.Save(ctx, roleCacheKey(r.ID), r, c.l2RedisTTL)
		}
		c.storeL1(r.ID, r)
	}
	log.Debugf("role permission cache warmed with %d roles", len(roles))
}
