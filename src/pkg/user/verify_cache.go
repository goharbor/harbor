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
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"os"
	"strconv"
	"sync"
	"time"
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
)

// verifyCache remembers recent successful local password verifications so that
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

// newVerifyCacheFromEnv builds the cache using verifyCacheTTLEnv, falling back
// to defaultVerifyCacheTTL when the variable is unset or invalid.
func newVerifyCacheFromEnv() *verifyCache {
	ttl := defaultVerifyCacheTTL
	if v, ok := os.LookupEnv(verifyCacheTTLEnv); ok {
		if d, err := time.ParseDuration(v); err == nil {
			ttl = d
		}
	}
	return newVerifyCache(ttl, defaultVerifyCacheMaxEntries)
}

func (c *verifyCache) fingerprint(password, salt, version, storedHash string) [sha256.Size]byte {
	mac := hmac.New(sha256.New, c.key)
	for _, field := range []string{version, salt, storedHash, password} {
		// Length-prefix each field so different splits cannot collide.
		mac.Write([]byte(strconv.Itoa(len(field)) + ":" + field))
	}
	var out [sha256.Size]byte
	copy(out[:], mac.Sum(nil))
	return out
}

// verified reports whether this exact password was successfully verified
// against this exact stored credential within the TTL.
func (c *verifyCache) verified(password, salt, version, storedHash string) bool {
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
func (c *verifyCache) remember(password, salt, version, storedHash string) {
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
