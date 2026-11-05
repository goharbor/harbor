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
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// redisVerifyKeyPrefix namespaces entries inside Harbor's cache; the cache
	// itself adds its own "cache:" prefix.
	redisVerifyKeyPrefix = "local-password-verify:"
	// redisVerifyKeyContext separates this HMAC key from any other use of
	// Harbor's secret key.
	redisVerifyKeyContext = "harbor:local-password-verify-cache:v1"
	// redisVerifyTimeout bounds each Redis call so a slow or unreachable Redis
	// degrades to the normal PBKDF2 verification instead of stalling logins.
	redisVerifyTimeout = 200 * time.Millisecond
)

// redisVerifyCache is the credentialCache backed by Harbor's shared Redis
// cache, so every harbor-core instance can reuse a verification made by any
// other instance.
//
// Security properties, in addition to those of verifyCache:
//   - Redis stores only HMAC-SHA256 fingerprints with a TTL, never passwords,
//     salts or hashes. The HMAC key is derived from Harbor's secret key, which
//     every core instance already shares and which is not stored in Redis.
//   - Someone who can read Redis alone cannot test password guesses. Someone
//     who has both Redis and Harbor's secret key could test guesses for live
//     entries at HMAC speed instead of PBKDF2 speed, so Redis should be
//     protected like the rest of Harbor's internal services.
//   - Redis errors and timeouts are treated as misses: verification falls back
//     to PBKDF2 and never succeeds because of the cache.
type redisVerifyCache struct {
	ttl       time.Duration
	timeout   time.Duration
	cache     func() cache.Cache
	secretKey func() (string, error)

	mu     sync.Mutex
	key    []byte
	warned bool
}

// newRedisVerifyCache returns a Redis-backed cache, or nil when ttl is not
// positive.
func newRedisVerifyCache(ttl time.Duration) *redisVerifyCache {
	if ttl <= 0 {
		return nil
	}
	return &redisVerifyCache{
		ttl:       ttl,
		timeout:   redisVerifyTimeout,
		cache:     cache.Default,
		secretKey: config.SecretKey,
	}
}

// hmacKey derives the shared HMAC key from Harbor's secret key on first use.
// Failures are not memoized, so a key that is unavailable early in startup is
// picked up later; until then the cache simply misses.
func (c *redisVerifyCache) hmacKey() []byte {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.key != nil {
		return c.key
	}
	secret, err := c.secretKey()
	if err != nil || secret == "" {
		if !c.warned {
			log.Warningf("password verify cache: secret key unavailable, Redis cache disabled until it is: %v", err)
			c.warned = true
		}
		return nil
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(redisVerifyKeyContext))
	c.key = mac.Sum(nil)
	return c.key
}

func (c *redisVerifyCache) entryKey(password, salt, version, storedHash string) (string, bool) {
	key := c.hmacKey()
	if key == nil {
		return "", false
	}
	fp := credentialFingerprint(key, password, salt, version, storedHash)
	return redisVerifyKeyPrefix + hex.EncodeToString(fp[:]), true
}

func (c *redisVerifyCache) verified(ctx context.Context, password, salt, version, storedHash string) bool {
	if c == nil {
		return false
	}
	backend := c.cache()
	if backend == nil {
		return false
	}
	k, ok := c.entryKey(password, salt, version, storedHash)
	if !ok {
		return false
	}
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	// Contains returns false on any Redis error, which falls back to PBKDF2.
	return backend.Contains(ctx, k)
}

func (c *redisVerifyCache) remember(ctx context.Context, password, salt, version, storedHash string) {
	if c == nil {
		return
	}
	backend := c.cache()
	if backend == nil {
		return
	}
	k, ok := c.entryKey(password, salt, version, storedHash)
	if !ok {
		return
	}
	// The password is already verified; do not let a cancelled request stop
	// the write, but keep it bounded.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), c.timeout)
	defer cancel()
	if err := backend.Save(ctx, k, 1, c.ttl); err != nil {
		log.Debugf("password verify cache: failed to save entry to Redis: %v", err)
	}
}
