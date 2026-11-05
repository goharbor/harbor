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

package user // nolint:revive

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/lib/log"
)

const (
	// verifyCacheTTLEnv overrides how long a successful local password
	// verification is remembered. It takes a Go duration ("30s", "2m");
	// "0" disables the cache.
	verifyCacheTTLEnv = "LOCAL_PASSWORD_VERIFY_CACHE_TTL"
	// defaultVerifyCacheTTL keeps a verified credential for one minute. That is
	// long enough to absorb bursts of Basic-Auth API calls from the same client
	// while keeping the exposure window of a cached entry short.
	defaultVerifyCacheTTL = time.Minute
	// defaultVerifyCacheMaxEntries bounds memory use. When the cache is full and
	// nothing has expired, new verifications are simply not cached, which falls
	// back to the uncached behaviour rather than evicting live entries.
	defaultVerifyCacheMaxEntries = 10000

	// verifyCacheTypeEnv selects where successful verifications are remembered:
	//   - "memory" (default): per-process cache; each core pod verifies a
	//     credential once per TTL.
	//   - "redis": Harbor's shared Redis cache (cache.Default()); all core pods
	//     share entries, so a credential is verified once per TTL cluster-wide.
	//   - "none": disable caching.
	verifyCacheTypeEnv = "LOCAL_PASSWORD_VERIFY_CACHE_TYPE"
	verifyCacheMemory  = "memory"
	verifyCacheRedis   = "redis"
	verifyCacheNone    = "none"
)

// credentialCache remembers recent successful local password verifications.
// Implementations must only report a hit for the exact password and the exact
// stored credential (version, salt, hash) that were remembered, and must treat
// any internal error as a miss so callers fall back to the full verification.
type credentialCache interface {
	verified(ctx context.Context, password, salt, version, storedHash string) bool
	remember(ctx context.Context, password, salt, version, storedHash string)
}

// newCredentialCacheFromEnv builds the cache selected by verifyCacheTypeEnv with
// the TTL from verifyCacheTTLEnv. It returns nil when caching is disabled.
func newCredentialCacheFromEnv() credentialCache {
	ttl := defaultVerifyCacheTTL
	if v, ok := os.LookupEnv(verifyCacheTTLEnv); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			ttl = d
		} else {
			log.Warningf("invalid %s %q, using default %s", verifyCacheTTLEnv, v, defaultVerifyCacheTTL)
		}
	}
	typ := strings.ToLower(strings.TrimSpace(os.Getenv(verifyCacheTypeEnv)))
	switch typ {
	case verifyCacheNone:
		return nil
	case verifyCacheRedis:
		if c := newRedisVerifyCache(ttl); c != nil {
			return c
		}
		return nil
	case "", verifyCacheMemory:
	default:
		log.Warningf("unknown %s %q, using %q", verifyCacheTypeEnv, typ, verifyCacheMemory)
	}
	if c := newVerifyCache(ttl, defaultVerifyCacheMaxEntries); c != nil {
		return c
	}
	return nil
}

// credentialFingerprint is HMAC-SHA256 under key over the password version,
// salt, stored hash and candidate password. Each field is length-prefixed so
// different splits of the same bytes cannot collide.
func credentialFingerprint(key []byte, password, salt, version, storedHash string) [sha256.Size]byte {
	mac := hmac.New(sha256.New, key)
	for _, field := range []string{version, salt, storedHash, password} {
		mac.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}
	var out [sha256.Size]byte
	copy(out[:], mac.Sum(nil))
	return out
}

// verifyCache is the in-memory credentialCache. It remembers recent successful local password verifications so that
// clients using HTTP Basic Auth do not pay for a full PBKDF2 derivation
// (600,000 iterations for PasswordVersion pbkdf2_sha256) on every request.
//
// Security properties:
//   - Only successful verifications are cached. Wrong passwords always pay the
//     full PBKDF2 cost, so online guessing is not made cheaper.
//   - The cache key is HMAC-SHA256 under a random per-process key over the
//     password version, salt, stored hash and candidate password. Any change to
//     the stored credential (password change or reset gives a new salt and
//     hash) yields a different key, so the old password stops working at once.
//   - Plaintext passwords are never stored, and the HMAC key never leaves the
//     process memory. An attacker able to read core's memory could test guesses
//     against live entries at HMAC speed, but such an attacker can already read
//     plaintext credentials from in-flight requests; the short TTL and size
//     bound limit the exposure further.
//   - The user record is still loaded from the database on every request, so
//     deleted users are rejected as before.
type verifyCache struct {
	mu         sync.Mutex
	key        []byte
	ttl        time.Duration
	maxEntries int
	entries    map[[sha256.Size]byte]time.Time
	now        func() time.Time
}

// newVerifyCache returns a cache with the given TTL and size bound, or nil when
// the TTL or size is not positive (caching disabled). A nil *verifyCache is safe
// to use and caches nothing.
func newVerifyCache(ttl time.Duration, maxEntries int) *verifyCache {
	if ttl <= 0 || maxEntries <= 0 {
		return nil
	}
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		// Without a secret key the cache would weaken stored hashes; disable it.
		return nil
	}
	return &verifyCache{
		key:        key,
		ttl:        ttl,
		maxEntries: maxEntries,
		entries:    make(map[[sha256.Size]byte]time.Time),
		now:        time.Now,
	}
}

func (c *verifyCache) fingerprint(password, salt, version, storedHash string) [sha256.Size]byte {
	return credentialFingerprint(c.key, password, salt, version, storedHash)
}

// verified reports whether this exact password was successfully verified
// against this exact stored credential within the TTL.
func (c *verifyCache) verified(_ context.Context, password, salt, version, storedHash string) bool {
	if c == nil {
		return false
	}
	fp := c.fingerprint(password, salt, version, storedHash)
	c.mu.Lock()
	defer c.mu.Unlock()
	expiry, ok := c.entries[fp]
	if !ok {
		return false
	}
	if !c.now().Before(expiry) {
		delete(c.entries, fp)
		return false
	}
	return true
}

// remember records a successful verification. Call it only after the password
// has been checked against storedHash with the full PBKDF2 derivation.
func (c *verifyCache) remember(_ context.Context, password, salt, version, storedHash string) {
	if c == nil {
		return
	}
	fp := c.fingerprint(password, salt, version, storedHash)
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if _, exists := c.entries[fp]; !exists && len(c.entries) >= c.maxEntries {
		for k, expiry := range c.entries {
			if !now.Before(expiry) {
				delete(c.entries, k)
			}
		}
		if len(c.entries) >= c.maxEntries {
			return
		}
	}
	c.entries[fp] = now.Add(c.ttl)
}
