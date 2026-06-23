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
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/cache"
	_ "github.com/goharbor/harbor/src/lib/cache/memory" // register the in-memory provider
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/role/model"
)

// stubController is a hand-rolled inner Controller for decorator tests; it counts
// calls and returns a canned role. (No generated mock exists for this interface —
// the file under src/testing/controller/role is a mislabeled project mock — and a
// local stub keeps the test in package role so it can assert cache internals.)
type stubController struct {
	getCalls    int32
	createCalls int32
	updateCalls int32
	deleteCalls int32
}

func (s *stubController) Get(_ context.Context, id int64, _ *Option) (*Role, error) {
	atomic.AddInt32(&s.getCalls, 1)
	return &Role{Role: model.Role{ID: id, Name: "r"}}, nil
}
func (s *stubController) Count(context.Context, *q.Query) (int64, error) { return 0, nil }
func (s *stubController) Create(context.Context, *Role) (int64, error) {
	atomic.AddInt32(&s.createCalls, 1)
	return 1, nil
}
func (s *stubController) Delete(context.Context, int64, ...*Option) error {
	atomic.AddInt32(&s.deleteCalls, 1)
	return nil
}
func (s *stubController) Update(context.Context, *Role, *Option) error {
	atomic.AddInt32(&s.updateCalls, 1)
	return nil
}
func (s *stubController) List(context.Context, *q.Query, *Option) ([]*Role, error) {
	return nil, nil // warm sees no roles
}

func (s *stubController) gets() int32 { return atomic.LoadInt32(&s.getCalls) }

func withPermission() *Option { return &Option{WithPermission: true} }

func newMemoryCache(t *testing.T) cache.Cache {
	t.Helper()
	c, err := cache.New(cache.Memory)
	require.NoError(t, err)
	return c
}

// L1 hit within the window does not delegate to the inner controller again.
func TestCacheL1Hit(t *testing.T) {
	inner := &stubController{}
	c := newCachingControllerWith(inner, time.Minute, -1) // L1 on, Redis off
	ctx := context.Background()

	_, err := c.Get(ctx, 1, withPermission())
	require.NoError(t, err)
	_, err = c.Get(ctx, 1, withPermission())
	require.NoError(t, err)

	assert.Equal(t, int32(1), inner.gets(), "second read should be served from L1")
}

// Once the L1 entry's window passes, the next read re-delegates to the inner.
func TestCacheL1Expiry(t *testing.T) {
	inner := &stubController{}
	c := newCachingControllerWith(inner, 20*time.Millisecond, -1) // L1 on (short), Redis off
	ctx := context.Background()

	_, err := c.Get(ctx, 1, withPermission())
	require.NoError(t, err)
	time.Sleep(40 * time.Millisecond)
	_, err = c.Get(ctx, 1, withPermission())
	require.NoError(t, err)

	assert.Equal(t, int32(2), inner.gets(), "read after L1 expiry should re-delegate")
}

// An L2 (Redis) hit serves the value and promotes it to L1 without delegating.
func TestCacheL2HitPromotes(t *testing.T) {
	inner := &stubController{}
	c := newCachingControllerWith(inner, time.Minute, time.Minute)
	c.redisCache = newMemoryCache(t)
	ctx := context.Background()

	// First read populates L2 + L1.
	_, err := c.Get(ctx, 1, withPermission())
	require.NoError(t, err)
	require.Equal(t, int32(1), inner.gets())

	// Drop the L1 entry so the next read must consult L2.
	c.local.Delete(int64(1))

	_, err = c.Get(ctx, 1, withPermission())
	require.NoError(t, err)
	assert.Equal(t, int32(1), inner.gets(), "L2 hit should avoid delegating")

	_, ok := c.local.Load(int64(1))
	assert.True(t, ok, "L2 hit should promote back into L1")
}

// Reads without permissions are never cached — they always delegate.
func TestCacheGetWithoutPermissionBypasses(t *testing.T) {
	inner := &stubController{}
	c := newCachingControllerWith(inner, time.Minute, time.Minute)
	c.redisCache = newMemoryCache(t)
	ctx := context.Background()

	_, err := c.Get(ctx, 1, nil)
	require.NoError(t, err)
	_, err = c.Get(ctx, 1, &Option{WithPermission: false})
	require.NoError(t, err)

	assert.Equal(t, int32(2), inner.gets(), "non-permission reads should always delegate")
	_, ok := c.local.Load(int64(1))
	assert.False(t, ok, "non-permission reads should not populate L1")
}

// Update/Delete/Create invalidate so the next read re-delegates.
func TestCacheWritesInvalidate(t *testing.T) {
	for _, tc := range []struct {
		name  string
		write func(c *cachingController, ctx context.Context) error
	}{
		{"update", func(c *cachingController, ctx context.Context) error {
			return c.Update(ctx, &Role{Role: model.Role{ID: 1}}, withPermission())
		}},
		{"delete", func(c *cachingController, ctx context.Context) error {
			return c.Delete(ctx, 1)
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			inner := &stubController{}
			c := newCachingControllerWith(inner, time.Minute, time.Minute)
			c.redisCache = newMemoryCache(t)
			ctx := context.Background()

			_, err := c.Get(ctx, 1, withPermission()) // cache it
			require.NoError(t, err)
			require.Equal(t, int32(1), inner.gets())

			require.NoError(t, tc.write(c, ctx))
			_, ok := c.local.Load(int64(1))
			require.False(t, ok, "write should drop the L1 entry")

			_, err = c.Get(ctx, 1, withPermission())
			require.NoError(t, err)
			assert.Equal(t, int32(2), inner.gets(), "read after a write should re-delegate")
		})
	}
}

// L1 disabled (ttl <= 0): every read consults L2; the inner runs only on the L2 miss.
func TestCacheL1Disabled(t *testing.T) {
	inner := &stubController{}
	c := newCachingControllerWith(inner, -1, time.Minute) // L1 off, Redis on
	c.redisCache = newMemoryCache(t)
	ctx := context.Background()

	_, err := c.Get(ctx, 1, withPermission())
	require.NoError(t, err)
	_, err = c.Get(ctx, 1, withPermission())
	require.NoError(t, err)

	assert.Equal(t, int32(1), inner.gets(), "second read should hit L2, not delegate")
	_, ok := c.local.Load(int64(1))
	assert.False(t, ok, "nothing should be stored in L1 when it is disabled")
}

// Redis disabled (ttl <= 0): reads go L1 -> inner and nothing is written to Redis.
func TestCacheRedisDisabled(t *testing.T) {
	inner := &stubController{}
	c := newCachingControllerWith(inner, time.Minute, -1) // L1 on, Redis off
	mc := newMemoryCache(t)
	c.redisCache = mc // present but must be ignored because L2 is disabled
	ctx := context.Background()

	_, err := c.Get(ctx, 1, withPermission())
	require.NoError(t, err)
	_, err = c.Get(ctx, 1, withPermission())
	require.NoError(t, err)

	assert.Equal(t, int32(1), inner.gets(), "second read should hit L1")
	assert.False(t, mc.Contains(ctx, roleCacheKey(1)), "nothing should be written to Redis when L2 is disabled")
}

// Both layers disabled (-1, -1): full bypass — every read delegates to the inner.
func TestCacheBothDisabled(t *testing.T) {
	inner := &stubController{}
	c := newCachingControllerWith(inner, -1, -1)
	c.redisCache = newMemoryCache(t) // present but must be ignored (L2 disabled)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := c.Get(ctx, 1, withPermission())
		require.NoError(t, err)
	}

	assert.Equal(t, int32(3), inner.gets(), "with both layers off every read should delegate")
	_, ok := c.local.Load(int64(1))
	assert.False(t, ok, "with L1 off nothing should be populated in L1")
	assert.False(t, c.redisCache.Contains(ctx, roleCacheKey(1)), "with L2 off nothing should be written to Redis")
}

// envDuration parses seconds, duration strings, and the <=0 disable sentinel.
func TestEnvDuration(t *testing.T) {
	const name = "ROLE_CACHE_TEST_DURATION"

	t.Run("unset uses default", func(t *testing.T) {
		assert.Equal(t, 5*time.Second, envDuration(name, 5*time.Second))
	})
	t.Run("integer seconds", func(t *testing.T) {
		t.Setenv(name, "30")
		assert.Equal(t, 30*time.Second, envDuration(name, time.Second))
	})
	t.Run("duration string", func(t *testing.T) {
		t.Setenv(name, "30m")
		assert.Equal(t, 30*time.Minute, envDuration(name, time.Second))
	})
	t.Run("negative disables", func(t *testing.T) {
		t.Setenv(name, "-1")
		assert.Equal(t, time.Duration(-1), envDuration(name, time.Second))
	})
	t.Run("invalid uses default", func(t *testing.T) {
		t.Setenv(name, "not-a-duration")
		assert.Equal(t, 7*time.Second, envDuration(name, 7*time.Second))
	})
}
