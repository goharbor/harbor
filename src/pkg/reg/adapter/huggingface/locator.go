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

package huggingface

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/cache/memory"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/adapter/huggingface/synth"
)

const (
	modelTTL    = 7 * 24 * time.Hour
	snapshotTTL = 7 * 24 * time.Hour
	// refsTTL bounds Hub calls: one proxied HEAD alone calls ManifestExist twice, and every
	// digest miss would otherwise list the refs again.
	refsTTL   = 60 * time.Second
	digestTTL = 7 * 24 * time.Hour
	blobTTL   = 30 * 24 * time.Hour

	sweepInterval = 10 * time.Minute

	kindManifest = "manifest"
	kindConfig   = "config"
	kindLayer    = "layer"
)

// processCache backs the locator where no default cache is initialized (jobservice). The memory
// cache drops an expired entry only when it is read again or matched by Scan, so a periodic Scan
// keeps entries of commits that are never requested again from piling up.
var processCache = sync.OnceValue(func() cache.Cache {
	c, _ := memory.New(cache.Options{Codec: cache.DefaultCodec()})
	go func() {
		for range time.Tick(sweepInterval) {
			sweep(context.Background(), c)
		}
	}()
	return c
})

func sweep(ctx context.Context, c cache.Cache) {
	if _, err := c.Scan(ctx, keyPrefix); err != nil {
		log.Warningf("failed to sweep the hugging face cache: %v", err)
	}
}

func defaultCache() cache.Cache {
	if c := cache.Default(); c != nil {
		return c
	}
	return processCache()
}

// location is where the content of a digest comes from.
type location struct {
	ModelID string
	Commit  string
	Kind    string
	// Path and Size are set for layers.
	Path string
	Size int64
}

// locator holds all adapter state in one cache. Keys are scoped by registry ID so a registry
// without a token can never be served gated content fetched with another registry's token.
// Repositories stand in for model IDs in keys: the Hub is case-insensitive, and the canonical
// model ID is only known after the first Hub call.
type locator struct {
	cache      cache.Cache
	registryID int64
}

const keyPrefix = "huggingface:"

func (l *locator) prefix() string {
	return keyPrefix + strconv.FormatInt(l.registryID, 10) + ":"
}

func (l *locator) modelKey(repository string) string {
	return l.prefix() + "model:" + repository
}

func (l *locator) snapshotKey(repository, commit string) string {
	return l.prefix() + "snapshot:" + repository + ":" + commit
}

func (l *locator) refsKey(repository string) string {
	return l.prefix() + "refs:" + repository
}

// scannedKey marks that every ref head of a repository was synthesized and indexed, so a digest
// still missing from the index is not in any of them.
func (l *locator) scannedKey(repository string) string {
	return l.prefix() + "scanned:" + repository
}

func (l *locator) digestKey(repository string, d digest.Digest) string {
	return l.prefix() + "digest:" + repository + ":" + d.String()
}

// blobKey is unscoped: the git blob id is content-addressed and the value is only its sha256.
func (l *locator) blobKey(blobID string) string {
	return keyPrefix + "blob:" + blobID
}

// modelID returns the canonical model ID of a repository when known, else the repository.
// Calling the Hub with the canonical ID avoids a 307, which costs an API call of its own.
func (l *locator) modelID(ctx context.Context, repository string) string {
	var id string
	if err := l.cache.Fetch(ctx, l.modelKey(repository), &id); err != nil || id == "" {
		return repository
	}
	return id
}

func (l *locator) setModelID(ctx context.Context, repository, modelID string) {
	if err := l.cache.Save(ctx, l.modelKey(repository), modelID, modelTTL); err != nil {
		log.Warningf("failed to save %s to cache: %v", l.modelKey(repository), err)
	}
}

// locate returns the cached location of a digest, or cache.ErrNotFound.
func (l *locator) locate(ctx context.Context, repository string, d digest.Digest) (*location, error) {
	loc := &location{}
	if err := l.cache.Fetch(ctx, l.digestKey(repository, d), loc); err != nil {
		return nil, err
	}
	return loc, nil
}

func (l *locator) put(ctx context.Context, repository string, d digest.Digest, loc *location) {
	if err := l.cache.Save(ctx, l.digestKey(repository, d), loc, digestTTL); err != nil {
		log.Warningf("failed to save %s to cache: %v", l.digestKey(repository, d), err)
	}
}

func (l *locator) scanned(ctx context.Context, repository string) bool {
	return l.cache.Contains(ctx, l.scannedKey(repository))
}

func (l *locator) markScanned(ctx context.Context, repository string) {
	if err := l.cache.Save(ctx, l.scannedKey(repository), true, refsTTL); err != nil {
		log.Warningf("failed to save %s to cache: %v", l.scannedKey(repository), err)
	}
}

func (l *locator) forget(ctx context.Context, repository string, d digest.Digest) {
	if err := l.cache.Delete(ctx, l.digestKey(repository, d)); err != nil {
		log.Warningf("failed to delete %s from cache: %v", l.digestKey(repository, d), err)
	}
}

// index records every digest of an artifact. The manifest is written last, so its presence
// means the whole artifact is indexed.
func (l *locator) index(ctx context.Context, repository, modelID, commit string, art *synth.Artifact) {
	manifestKey := l.digestKey(repository, art.Digest)
	if l.cache.Contains(ctx, manifestKey) {
		return
	}
	save := func(key string, loc *location) bool {
		if err := l.cache.Save(ctx, key, loc, digestTTL); err != nil {
			log.Warningf("failed to save %s to cache: %v", key, err)
			return false
		}
		return true
	}
	for _, layer := range art.Layers {
		loc := &location{ModelID: modelID, Commit: commit, Kind: kindLayer, Path: layer.Path, Size: layer.Size}
		if !save(l.digestKey(repository, layer.Digest), loc) {
			return
		}
	}
	if !save(l.digestKey(repository, art.ConfigDigest), &location{ModelID: modelID, Commit: commit, Kind: kindConfig}) {
		return
	}
	save(manifestKey, &location{ModelID: modelID, Commit: commit, Kind: kindManifest})
}
