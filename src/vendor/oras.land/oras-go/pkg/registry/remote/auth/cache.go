/*
Copyright The ORAS Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package auth

import (
	"context"
	"strings"
	"sync"

	errdef "oras.land/oras-go/pkg/content"
	"oras.land/oras-go/pkg/registry/remote/internal/syncutil"
)

// DefaultCache is the sharable cache used by DefaultClient.
var DefaultCache Cache = NewCache()

// Cache caches the auth-scheme and auth-token for the "Authorization" header in
// accessing the remote registry.
// Precisely, the header is `Authorization: auth-scheme auth-token`.
// The `auth-token` is a generic term as `token68` in RFC 7235 section 2.1.
type Cache interface {
	// GetScheme returns the auth-scheme part cached for the given registry.
	// A single registry is assumed to have a consistent scheme.
	// If a registry has different schemes per path, the auth client is still
	// workable. However, the cache may not be effective as the cache cannot
	// correctly guess the scheme.
	GetScheme(ctx context.Context, registry string) (Scheme, error)

	// GetToken returns the auth-token part cached for the given registry of a
	// given scheme.
	// The underlying implementation MAY cache the token for all schemes for the
	// given registry.
	GetToken(ctx context.Context, registry string, scheme Scheme, key string) (string, error)

	// Set fetches the token using the given fetch function and caches the token
	// for the given scheme with the given key for the given registry.
	// The return values of the fetch function is returned by this function.
	// The underlying implementation MAY combine the fetch operation if the Set
	// function is invoked multiple times at the same time.
	Set(ctx context.Context, registry string, scheme Scheme, key string, fetch func(context.Context) (string, error)) (string, error)
}

// cacheEntry is a cache entry for a single registry.
type cacheEntry struct {
	scheme Scheme
	tokens sync.Map // map[string]string
}

// concurrentCache is a cache suitable for concurrent invocation.
type concurrentCache struct {
	status sync.Map // map[string]*syncutil.Once
	cache  sync.Map // map[string]*cacheEntry
}

// NewCache creates a new go-routine safe cache instance.
func NewCache() Cache {
	return &concurrentCache{}
}

// GetScheme returns the auth-scheme part cached for the given registry.
func (cc *concurrentCache) GetScheme(ctx context.Context, registry string) (Scheme, error) {
	entry, ok := cc.cache.Load(registry)
	if !ok {
		return SchemeUnknown, errdef.ErrNotFound
	}
	return entry.(*cacheEntry).scheme, nil
}

// GetToken returns the auth-token part cached for the given registry of a given
// scheme.
func (cc *concurrentCache) GetToken(ctx context.Context, registry string, scheme Scheme, key string) (string, error) {
	entryValue, ok := cc.cache.Load(registry)
	if !ok {
		return "", errdef.ErrNotFound
	}
	entry := entryValue.(*cacheEntry)
	if entry.scheme != scheme {
		return "", errdef.ErrNotFound
	}
	if token, ok := entry.tokens.Load(key); ok {
		return token.(string), nil
	}
	return "", errdef.ErrNotFound
}

// Set fetches the token using the given fetch function and caches the token
// for the given scheme with the given key for the given registry.
// Set combines the fetch operation if the Set is invoked multiple times at the
// same time.
func (cc *concurrentCache) Set(ctx context.Context, registry string, scheme Scheme, key string, fetch func(context.Context) (string, error)) (string, error) {
	// fetch token
	statusKey := strings.Join([]string{
		registry,
		scheme.String(),
		key,
	}, " ")
	statusValue, _ := cc.status.LoadOrStore(statusKey, syncutil.NewOnce())
	fetchOnce := statusValue.(*syncutil.Once)
	fetchedFirst, result, err := fetchOnce.Do(ctx, func() (interface{}, error) {
		return fetch(ctx)
	})
	if fetchedFirst {
		cc.status.Delete(statusKey)
	}
	if err != nil {
		return "", err
	}
	token := result.(string)
	if !fetchedFirst {
		return token, nil
	}

	// cache token
	newEntry := &cacheEntry{
		scheme: scheme,
	}
	entryValue, exists := cc.cache.LoadOrStore(registry, newEntry)
	entry := entryValue.(*cacheEntry)
	if exists && entry.scheme != scheme {
		// there is a scheme change, which is not expected in most scenarios.
		// force invalidating all previous cache.
		entry = newEntry
		cc.cache.Store(registry, entry)
	}
	entry.tokens.Store(key, token)

	return token, nil
}

// noCache is a cache implementation that does not do cache at all.
type noCache struct{}

// GetScheme always returns not found error as it has no cache.
func (noCache) GetScheme(ctx context.Context, registry string) (Scheme, error) {
	return SchemeUnknown, errdef.ErrNotFound
}

// GetToken always returns not found error as it has no cache.
func (noCache) GetToken(ctx context.Context, registry string, scheme Scheme, key string) (string, error) {
	return "", errdef.ErrNotFound
}

// Set calls fetch directly without caching.
func (noCache) Set(ctx context.Context, registry string, scheme Scheme, key string, fetch func(context.Context) (string, error)) (string, error) {
	return fetch(ctx)
}
