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

package user

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
)

// fakeRedis is a minimal cache.Cache standing in for Harbor's Redis cache.
type fakeRedis struct {
	mu      sync.Mutex
	entries map[string]time.Duration
	failing bool
	saves   int
	lookups int
}

func newFakeRedis() *fakeRedis { return &fakeRedis{entries: map[string]time.Duration{}} }

func (f *fakeRedis) Contains(_ context.Context, key string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lookups++
	if f.failing {
		return false // lib/cache/redis returns false on any error
	}
	_, ok := f.entries[key]
	return ok
}

func (f *fakeRedis) Save(_ context.Context, key string, _ any, exp ...time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.saves++
	if f.failing {
		return errors.New("redis unavailable")
	}
	var ttl time.Duration
	if len(exp) > 0 {
		ttl = exp[0]
	}
	f.entries[key] = ttl
	return nil
}

func (f *fakeRedis) Delete(_ context.Context, key string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.entries, key)
	return nil
}
func (f *fakeRedis) Fetch(context.Context, string, any) error             { return cache.ErrNotFound }
func (f *fakeRedis) Ping(context.Context) error                           { return nil }
func (f *fakeRedis) Scan(context.Context, string) (cache.Iterator, error) { return nil, nil }

func newTestRedisVerifyCache(backend cache.Cache, secret string) *redisVerifyCache {
	c := newRedisVerifyCache(time.Minute)
	c.cache = func() cache.Cache { return backend }
	c.secretKey = func() (string, error) { return secret, nil }
	return c
}

func TestRedisVerifyCacheDisabled(t *testing.T) {
	if newRedisVerifyCache(0) != nil {
		t.Fatal("zero TTL should disable the Redis cache")
	}
	var c *redisVerifyCache
	c.remember(testCtx, "pw", "s", "v", "h")
	if c.verified(testCtx, "pw", "s", "v", "h") {
		t.Fatal("nil Redis cache reported a hit")
	}
}

func TestRedisVerifyCacheHitAndMiss(t *testing.T) {
	r := newFakeRedis()
	c := newTestRedisVerifyCache(r, "secret-key-16chr")
	if c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("empty cache reported a hit")
	}
	c.remember(testCtx, "pw", "salt", "v", "hash")
	if !c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("remembered credential was not a hit")
	}
	for _, tc := range [][4]string{
		{"wrong", "salt", "v", "hash"},
		{"pw", "newsalt", "v", "hash"},
		{"pw", "salt", "other", "hash"},
		{"pw", "salt", "v", "newhash"},
	} {
		if c.verified(testCtx, tc[0], tc[1], tc[2], tc[3]) {
			t.Errorf("unexpected hit for %v", tc)
		}
	}
}

func TestRedisVerifyCacheEntryFormat(t *testing.T) {
	r := newFakeRedis()
	c := newTestRedisVerifyCache(r, "secret-key-16chr")
	c.remember(testCtx, "Harbor12345", "salt", "pbkdf2_sha256", "storedhash")
	if len(r.entries) != 1 {
		t.Fatalf("expected one entry, got %d", len(r.entries))
	}
	for k, ttl := range r.entries {
		if !strings.HasPrefix(k, redisVerifyKeyPrefix) || len(k) != len(redisVerifyKeyPrefix)+64 {
			t.Errorf("unexpected key format %q", k)
		}
		for _, secret := range []string{"Harbor12345", "salt", "storedhash"} {
			if strings.Contains(k, secret) {
				t.Errorf("key %q leaks %q", k, secret)
			}
		}
		if ttl != time.Minute {
			t.Errorf("entry saved with TTL %s, want 1m", ttl)
		}
	}
}

// Two core instances with the same Harbor secret key share entries; an
// instance with a different key cannot use them.
func TestRedisVerifyCacheSharedAcrossInstances(t *testing.T) {
	r := newFakeRedis()
	podA := newTestRedisVerifyCache(r, "secret-key-16chr")
	podB := newTestRedisVerifyCache(r, "secret-key-16chr")
	other := newTestRedisVerifyCache(r, "another-key-16ch")

	podA.remember(testCtx, "pw", "salt", "v", "hash")
	if !podB.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("instance sharing the secret key did not see the entry")
	}
	if other.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("instance with a different secret key matched the entry")
	}
}

func TestRedisVerifyCacheBackendFailureIsMiss(t *testing.T) {
	r := newFakeRedis()
	c := newTestRedisVerifyCache(r, "secret-key-16chr")
	c.remember(testCtx, "pw", "salt", "v", "hash")
	r.failing = true
	if c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("Redis failure must be treated as a miss")
	}
	c.remember(testCtx, "pw2", "salt", "v", "hash") // must not panic
}

func TestRedisVerifyCacheNoBackend(t *testing.T) {
	c := newTestRedisVerifyCache(nil, "secret-key-16chr")
	c.cache = func() cache.Cache { return nil }
	c.remember(testCtx, "pw", "salt", "v", "hash")
	if c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("cache without a backend reported a hit")
	}
}

func TestRedisVerifyCacheSecretKeyUnavailable(t *testing.T) {
	r := newFakeRedis()
	c := newTestRedisVerifyCache(r, "")
	c.secretKey = func() (string, error) { return "", errors.New("not ready") }
	c.remember(testCtx, "pw", "salt", "v", "hash")
	if len(r.entries) != 0 || r.saves != 0 {
		t.Fatal("nothing should be written without a secret key")
	}
	if c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("cache without a secret key reported a hit")
	}
	// Once the key becomes available, the cache starts working.
	c.secretKey = func() (string, error) { return "secret-key-16chr", nil }
	c.remember(testCtx, "pw", "salt", "v", "hash")
	if !c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("cache did not recover after the secret key became available")
	}
}

// TestDeriveRedisVerifyKeyIsHKDF checks the derived key against an independent
// RFC 5869 HKDF-SHA256 computation (zero-length salt, one output block).
func TestDeriveRedisVerifyKeyIsHKDF(t *testing.T) {
	const secret = "0123456789abcdef"
	got, err := deriveRedisVerifyKey(secret)
	if err != nil {
		t.Fatalf("deriveRedisVerifyKey: %v", err)
	}
	if len(got) != sha256.Size {
		t.Fatalf("key length = %d, want %d", len(got), sha256.Size)
	}

	// Extract: PRK = HMAC-SHA256(salt = HashLen zero bytes, IKM = secret).
	extract := hmac.New(sha256.New, make([]byte, sha256.Size))
	extract.Write([]byte(secret))
	prk := extract.Sum(nil)
	// Expand: T(1) = HMAC-SHA256(PRK, info || 0x01).
	expand := hmac.New(sha256.New, prk)
	expand.Write([]byte(redisVerifyKeyContext))
	expand.Write([]byte{0x01})
	want := expand.Sum(nil)

	if !bytes.Equal(got, want) {
		t.Fatalf("derived key does not match RFC 5869 HKDF-SHA256")
	}

	other, err := deriveRedisVerifyKey(secret + "x")
	if err != nil {
		t.Fatalf("deriveRedisVerifyKey: %v", err)
	}
	if bytes.Equal(got, other) {
		t.Fatalf("different secrets produced the same key")
	}
}
