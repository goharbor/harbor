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
	"context"
	"testing"
	"time"
)

var testCtx = context.Background()

func TestVerifyCacheNilIsSafe(t *testing.T) {
	var c *verifyCache
	c.remember(testCtx, "pw", "salt", "pbkdf2_sha256", "hash")
	if c.verified(testCtx, "pw", "salt", "pbkdf2_sha256", "hash") {
		t.Fatal("nil cache must never report a verified password")
	}
}

func TestNewVerifyCacheDisabled(t *testing.T) {
	if newVerifyCache(0, 10) != nil {
		t.Error("zero TTL should disable the cache")
	}
	if newVerifyCache(time.Minute, 0) != nil {
		t.Error("zero size should disable the cache")
	}
}

func TestVerifyCacheHitAndMiss(t *testing.T) {
	c := newVerifyCache(time.Minute, 10)
	if c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("empty cache reported a hit")
	}
	c.remember(testCtx, "pw", "salt", "v", "hash")
	if !c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("remembered credential was not a hit")
	}
	// Any differing field must miss.
	cases := [][4]string{
		{"wrong", "salt", "v", "hash"},
		{"pw", "newsalt", "v", "hash"},
		{"pw", "salt", "other", "hash"},
		{"pw", "salt", "v", "newhash"},
		// Field boundaries must not be ambiguous.
		{"wpw", "salt", "v", "hash"},
		{"pw", "sal", "tv", "hash"},
	}
	for _, tc := range cases {
		if c.verified(testCtx, tc[0], tc[1], tc[2], tc[3]) {
			t.Errorf("unexpected hit for %v", tc)
		}
	}
}

func TestVerifyCacheExpiry(t *testing.T) {
	c := newVerifyCache(time.Minute, 10)
	now := time.Unix(1000, 0)
	c.now = func() time.Time { return now }
	c.remember(testCtx, "pw", "salt", "v", "hash")
	now = now.Add(59 * time.Second)
	if !c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("entry expired too early")
	}
	now = now.Add(time.Second)
	if c.verified(testCtx, "pw", "salt", "v", "hash") {
		t.Fatal("entry did not expire")
	}
	if len(c.entries) != 0 {
		t.Fatal("expired entry was not removed")
	}
}

func TestVerifyCacheBounded(t *testing.T) {
	c := newVerifyCache(time.Minute, 2)
	now := time.Unix(1000, 0)
	c.now = func() time.Time { return now }
	c.remember(testCtx, "a", "s", "v", "h")
	c.remember(testCtx, "b", "s", "v", "h")
	c.remember(testCtx, "c", "s", "v", "h") // full: not cached, live entries kept
	if len(c.entries) != 2 {
		t.Fatalf("cache exceeded its bound: %d entries", len(c.entries))
	}
	if c.verified(testCtx, "c", "s", "v", "h") {
		t.Fatal("entry beyond the bound should not be cached")
	}
	if !c.verified(testCtx, "a", "s", "v", "h") || !c.verified(testCtx, "b", "s", "v", "h") {
		t.Fatal("live entries were evicted")
	}
	// Once entries expire there is room again.
	now = now.Add(2 * time.Minute)
	c.remember(testCtx, "c", "s", "v", "h")
	if !c.verified(testCtx, "c", "s", "v", "h") {
		t.Fatal("expired entries were not purged to make room")
	}
}

func TestVerifyCacheKeyIsPerInstance(t *testing.T) {
	a := newVerifyCache(time.Minute, 10)
	b := newVerifyCache(time.Minute, 10)
	if a.fingerprint("pw", "s", "v", "h") == b.fingerprint("pw", "s", "v", "h") {
		t.Fatal("two caches produced the same fingerprint; HMAC key is not random")
	}
}

func TestNewCredentialCacheFromEnv(t *testing.T) {
	t.Setenv(verifyCacheTypeEnv, "")
	t.Setenv(verifyCacheTTLEnv, "0")
	if newCredentialCacheFromEnv() != nil {
		t.Error("TTL 0 should disable the cache")
	}
	t.Setenv(verifyCacheTTLEnv, "30s")
	if c, ok := newCredentialCacheFromEnv().(*verifyCache); !ok || c.ttl != 30*time.Second {
		t.Error("TTL override was not applied to the default memory cache")
	}
	t.Setenv(verifyCacheTTLEnv, "not-a-duration")
	if c, ok := newCredentialCacheFromEnv().(*verifyCache); !ok || c.ttl != defaultVerifyCacheTTL {
		t.Error("invalid TTL should fall back to the default")
	}

	t.Setenv(verifyCacheTTLEnv, "")
	t.Setenv(verifyCacheTypeEnv, "memory")
	if _, ok := newCredentialCacheFromEnv().(*verifyCache); !ok {
		t.Error("memory type should build the in-memory cache")
	}
	t.Setenv(verifyCacheTypeEnv, "Redis")
	if c, ok := newCredentialCacheFromEnv().(*redisVerifyCache); !ok || c.ttl != defaultVerifyCacheTTL {
		t.Error("redis type should build the Redis cache")
	}
	t.Setenv(verifyCacheTypeEnv, "none")
	if newCredentialCacheFromEnv() != nil {
		t.Error("none type should disable the cache")
	}
	t.Setenv(verifyCacheTypeEnv, "bogus")
	if _, ok := newCredentialCacheFromEnv().(*verifyCache); !ok {
		t.Error("unknown type should fall back to the memory cache")
	}
	t.Setenv(verifyCacheTypeEnv, "redis")
	t.Setenv(verifyCacheTTLEnv, "0")
	if newCredentialCacheFromEnv() != nil {
		t.Error("TTL 0 should disable the Redis cache too")
	}
}
