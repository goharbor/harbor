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

package proxy

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/docker/distribution"
	"github.com/docker/distribution/manifest/manifestlist"
	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/lib"
	libCache "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/log"
)

const defaultHandler = "default"

const pendingManifestList = "pendingmanifestlist:"

type pendingManifestListState struct {
	Repository  string
	Digest      string
	Tag         string
	ContentType string
	Payload     []byte
}

// NewCacheHandlerRegistry ...
func NewCacheHandlerRegistry(local localInterface) map[string]ManifestCacheHandler {
	manListHandler := &ManifestListCache{
		local: local,
		cache: libCache.Default(),
	}
	manHandler := &ManifestCache{local: local, manifestListCache: manListHandler}
	registry := map[string]ManifestCacheHandler{
		manifestlist.MediaTypeManifestList: manListHandler,
		v1.MediaTypeImageIndex:             manListHandler,
		schema2.MediaTypeManifest:          manHandler,
		defaultHandler:                     manHandler,
	}
	return registry
}

// ManifestCacheHandler define how to cache manifest content
type ManifestCacheHandler interface {
	// CacheContent - cache the content of the manifest
	CacheContent(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, r RemoteInterface, contentType string)
}

// ManifestListCache handle Manifest list type and index type
type ManifestListCache struct {
	cache libCache.Cache
	local localInterface
}

// CacheContent ...
func (m *ManifestListCache) CacheContent(ctx context.Context, _ string, man distribution.Manifest, art lib.ArtifactInfo, _ RemoteInterface, contentType string) {
	_, payload, err := man.Payload()
	if err != nil {
		log.Errorf("failed to get payload, error %v", err)
		return
	}
	if len(getReference(art)) == 0 {
		log.Errorf("failed to get reference, reference is empty, skip to cache manifest list")
		return
	}
	// cache key should contain digest if digest exist
	if len(art.Digest) == 0 {
		art.Digest = string(digest.FromBytes(payload))
	}
	key := manifestListKey(art.Repository, art)
	log.Debugf("cache manifest list with key=cache:%v", key)
	if err := m.cache.Save(ctx, manifestListContentTypeKey(art.Repository, art), contentType, manifestListCacheInterval); err != nil {
		log.Errorf("failed to cache content type, error %v", err)
	}
	if err := m.cache.Save(ctx, key, payload, manifestListCacheInterval); err != nil {
		log.Errorf("failed to cache payload, error %v", err)
	}
	if err := m.push(ctx, art, man, contentType); err != nil {
		log.Errorf("error when push manifest list to local :%v", err)
	}
}

// cacheTrimmedDigest - cache the change Trimmed Digest for controller.EnsureTag when digest is changed
func (m *ManifestListCache) cacheTrimmedDigest(ctx context.Context, originDig string, newDig string) {
	if m.cache == nil || len(originDig) == 0 {
		return
	}
	key := TrimmedManifestlist + originDig
	err := m.cache.Save(ctx, key, newDig)
	if err != nil {
		log.Warningf("failed to cache the trimmed manifest, err %v", err)
		return
	}
	log.Debugf("Saved key:%v, value:%v", key, newDig)
}

func (m *ManifestListCache) getTrimmedDigest(ctx context.Context, originDig string) (string, error) {
	if m.cache == nil || len(originDig) == 0 {
		return "", nil
	}
	var trimmedDigest string
	err := m.cache.Fetch(ctx, TrimmedManifestlist+originDig, &trimmedDigest)
	if errors.Is(err, libCache.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return trimmedDigest, nil
}

func pendingManifestListKey(repo string, childDigest string, parentDigest string) string {
	return pendingManifestList + repo + ":" + childDigest + ":" + parentDigest
}

func (m *ManifestListCache) registerPendingManifestList(ctx context.Context, art lib.ArtifactInfo, man distribution.Manifest, contentType string) {
	if m.cache == nil || len(art.Digest) == 0 {
		return
	}
	_, payload, err := man.Payload()
	if err != nil {
		log.Errorf("failed to get payload for pending manifest list, error %v", err)
		return
	}
	state := pendingManifestListState{
		Repository:  art.Repository,
		Digest:      art.Digest,
		Tag:         art.Tag,
		ContentType: contentType,
		Payload:     payload,
	}
	for _, ref := range man.References() {
		key := pendingManifestListKey(art.Repository, string(ref.Digest), art.Digest)
		if err := m.cache.Save(ctx, key, state, manifestListCacheInterval); err != nil {
			log.Errorf("failed to save pending manifest list, key %v, error %v", key, err)
		}
	}
}

func (m *ManifestListCache) deletePendingManifestList(ctx context.Context, art lib.ArtifactInfo, man distribution.Manifest) {
	if m.cache == nil || len(art.Digest) == 0 {
		return
	}
	for _, ref := range man.References() {
		key := pendingManifestListKey(art.Repository, string(ref.Digest), art.Digest)
		if err := m.cache.Delete(ctx, key); err != nil {
			log.Warningf("failed to delete pending manifest list, key %v, error %v", key, err)
		}
	}
}

func (m *ManifestListCache) reconcilePendingManifestLists(ctx context.Context, repo string, childDigest string) {
	if m.cache == nil || len(childDigest) == 0 {
		return
	}
	iter, err := m.cache.Scan(ctx, pendingManifestList+repo+":"+childDigest+":")
	if err != nil {
		log.Errorf("failed to scan pending manifest list for repo %v digest %v, error %v", repo, childDigest, err)
		return
	}
	for iter.Next(ctx) {
		key := iter.Val()
		var state pendingManifestListState
		if err := m.cache.Fetch(ctx, key, &state); err != nil {
			if !errors.Is(err, libCache.ErrNotFound) {
				log.Errorf("failed to fetch pending manifest list, key %v, error %v", key, err)
			}
			continue
		}
		man, _, err := distribution.UnmarshalManifest(state.ContentType, state.Payload)
		if err != nil {
			log.Errorf("failed to unmarshal pending manifest list, key %v, error %v", key, err)
			continue
		}
		art := lib.ArtifactInfo{Repository: state.Repository, Digest: state.Digest, Tag: state.Tag}
		if err := m.push(ctx, art, man, state.ContentType); err != nil {
			log.Errorf("failed to reconcile pending manifest list, key %v, error %v", key, err)
		}
	}
}

func (m *ManifestListCache) updateManifestList(ctx context.Context, repo string, manifest distribution.Manifest) (distribution.Manifest, error) {
	switch v := manifest.(type) {
	case *manifestlist.DeserializedManifestList:
		existMans := make([]manifestlist.ManifestDescriptor, 0)
		for _, ma := range v.Manifests {
			art := lib.ArtifactInfo{Repository: repo, Digest: string(ma.Digest)}
			a, err := m.local.GetManifest(ctx, art)
			if err != nil {
				return nil, err
			}
			if a != nil {
				existMans = append(existMans, ma)
			}
		}
		return manifestlist.FromDescriptors(existMans)
	}
	return nil, fmt.Errorf("current manifest list type is unknown, manifest type[%T], content [%+v]", manifest, manifest)
}

func (m *ManifestListCache) push(ctx context.Context, art lib.ArtifactInfo, man distribution.Manifest, contentType string) error {
	// For manifest list, it might include some different manifest
	// it will wait and check for 30 mins, if all depend manifests are ready then push it
	// if time exceed, then push a updated manifest list which contains existing manifest
	repo := art.Repository
	reference := getReference(art)
	var newMan distribution.Manifest
	var err error
	for range maxManifestListWait {
		log.Debugf("waiting for the manifest ready, repo %v, tag:%v", repo, reference)
		time.Sleep(time.Duration(sleepIntervalSec) * time.Second)
		newMan, err = m.updateManifestList(ctx, repo, man)
		if err != nil {
			return err
		}
		if len(newMan.References()) == len(man.References()) {
			break
		}
	}
	if len(newMan.References()) == 0 {
		m.registerPendingManifestList(ctx, art, man, contentType)
		log.Debugf("manifest list is pending because no child manifest is ready yet, repository: %v, reference: %v", repo, reference)
		return nil
	}
	_, pl, err := newMan.Payload()
	if err != nil {
		log.Errorf("failed to get payload, error %v", err)
		return err
	}
	log.Debugf("The manifest list payload: %v", string(pl))
	newDig := digest.FromBytes(pl)
	previousTrimmedDigest, err := m.getTrimmedDigest(ctx, art.Digest)
	if err != nil {
		return err
	}
	m.cacheTrimmedDigest(ctx, art.Digest, string(newDig))
	err = m.local.PushManifest(repo, string(newDig), newMan)
	if err != nil {
		log.Errorf("failed to push manifest list, error: %v", err)
		return err
	}
	if len(art.Tag) > 0 {
		err = m.local.PushManifest(repo, art.Tag, newMan)
		if err != nil {
			log.Errorf("failed to push manifest list tag, error: %v", err)
			return err
		}
	}
	if len(newMan.References()) < len(man.References()) {
		m.registerPendingManifestList(ctx, art, man, contentType)
		log.Debugf("push manifest list partially, repository: %v, reference: %v, digest: %v", repo, reference, newDig)
	} else {
		m.deletePendingManifestList(ctx, art, man)
		log.Debugf("push manifest list successfully, repository: %v, reference: %v, digest: %v", repo, reference, newDig)
	}
	if len(previousTrimmedDigest) > 0 && previousTrimmedDigest != string(newDig) {
		m.local.DeleteManifest(repo, previousTrimmedDigest)
	}
	log.Debug("update artifact pull time to avoid it is removed by GC before the manifest list is pushed to local")
	artForPullTime := lib.ArtifactInfo{Repository: repo, Digest: string(newDig), Tag: art.Tag}
	if err := m.local.UpdatePullTime(ctx, artForPullTime); err != nil {
		log.Errorf("failed to update pull time for artifact %v:%v, error: %v", artForPullTime.Repository, getReference(artForPullTime), err)
	}
	return nil
}

// ManifestCache default Manifest handler
type ManifestCache struct {
	local             localInterface
	manifestListCache *ManifestListCache
}

// CacheContent ...
func (m *ManifestCache) CacheContent(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, r RemoteInterface, _ string) {
	var waitBlobs []distribution.Descriptor
	for n := range maxManifestWait {
		time.Sleep(time.Duration(sleepIntervalSec) * time.Second)
		waitBlobs = m.local.CheckDependencies(ctx, art.Repository, man)
		if len(waitBlobs) == 0 {
			break
		}
		log.Debugf("Current n=%v artifact: %v:%v", n, art.Repository, art.Tag)
	}
	if len(waitBlobs) > 0 {
		// docker client will skip to pull layers exist in local
		// these blobs are not exist in the proxy server
		// it will cause the manifest dependency check always fail
		// need to push these blobs before push manifest to avoid failure
		log.Debug("Waiting blobs not empty, push it to local repo directly")
		for _, desc := range waitBlobs {
			err := m.putBlobToLocal(remoteRepo, art.Repository, desc, r)
			if err != nil {
				log.Errorf("Failed to push blob to local repo, error: %v", err)
				return
			}
		}
	}

	digestAvailable, err := m.push(art, man)
	if err != nil {
		log.Errorf("error occurred on manifest push to local: %v", err)
		if !digestAvailable {
			return
		}
	}
	if m.manifestListCache != nil && len(art.Digest) > 0 {
		m.manifestListCache.reconcilePendingManifestLists(ctx, art.Repository, art.Digest)
	}
}

func (m *ManifestCache) push(art lib.ArtifactInfo, man distribution.Manifest) (bool, error) {
	errs := []error{}
	digestAvailable := false
	if len(art.Digest) > 0 {
		err := m.local.PushManifest(art.Repository, art.Digest, man)
		if err != nil {
			log.Errorf("failed to push manifest referencing digest, tag: %v, digest: %v, error %v", art.Tag, art.Digest, err)
			errs = append(errs, err)
		} else {
			digestAvailable = true
		}
	}
	if len(art.Tag) > 0 {
		err := m.local.PushManifest(art.Repository, art.Tag, man)
		if err != nil {
			log.Errorf("failed to push manifest referencing tag, tag: %v, digest: %v, error %v", art.Tag, art.Digest, err)
			errs = append(errs, err)
		}
	}
	return digestAvailable, errors.Join(errs...)
}

func (m *ManifestCache) putBlobToLocal(remoteRepo string, localRepo string, desc distribution.Descriptor, r RemoteInterface) error {
	log.Debugf("Put blob to local registry!, sourceRepo:%v, localRepo:%v, digest: %v", remoteRepo, localRepo, desc.Digest)
	_, bReader, err := r.BlobReader(remoteRepo, string(desc.Digest))
	if err != nil {
		log.Errorf("failed to create blob reader, error %v", err)
		return err
	}
	defer bReader.Close()
	err = m.local.PushBlob(localRepo, desc, bReader)
	return err
}
