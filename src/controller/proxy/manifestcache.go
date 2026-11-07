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
	libErrors "github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
)

const defaultHandler = "default"

// NewCacheHandlerRegistry ...
func NewCacheHandlerRegistry(local localInterface) map[string]ManifestCacheHandler {
	manListHandler := &ManifestListCache{
		local: local,
		cache: libCache.Default(),
	}
	manHandler := &ManifestCache{local}
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
func (m *ManifestListCache) CacheContent(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, r RemoteInterface, contentType string) {
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
	if err := m.waitAndPush(ctx, remoteRepo, art, man, r); err != nil {
		log.Errorf("error when push manifest list to local :%v", err)
	}
}

func trimmedDigestKey(repo, dig string) string {
	return TrimmedManifestlist + repo + ":" + dig
}

// cacheTrimmedDigest - cache the change Trimmed Digest for controller.EnsureTag when digest is changed,
// and per repository for controller.manifestListTagged
func (m *ManifestListCache) cacheTrimmedDigest(ctx context.Context, repo, dig, newDig string) {
	if m.cache == nil {
		return
	}
	key := TrimmedManifestlist + dig
	if err := m.cache.Save(ctx, key, newDig); err != nil {
		log.Warningf("failed to cache the trimmed manifest, err %v", err)
		return
	}
	if err := m.cache.Save(ctx, trimmedDigestKey(repo, dig), newDig, manifestListCacheInterval); err != nil {
		log.Warningf("failed to cache the trimmed manifest, err %v", err)
		return
	}
	log.Debugf("Saved key:%v, value:%v", key, newDig)
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

func (m *ManifestListCache) waitAndPush(ctx context.Context, remoteRepo string, art lib.ArtifactInfo, man distribution.Manifest, r RemoteInterface) error {
	key := manifestListKey(art.Repository, art) + ":" + art.Tag
	if !inflightChecker.addRequest(key) {
		return nil
	}
	defer inflightChecker.removeRequest(key)
	// For manifest list, it might include some different manifest
	// it will wait and check for 30 mins, if all depend manifests are ready then push it
	// if time exceed, then push a updated manifest list which contains existing manifest
	newMan, err := m.updateManifestList(ctx, art.Repository, man)
	for i := 0; err == nil && len(newMan.References()) < len(man.References()) && i < maxManifestListWait; i++ {
		log.Debugf("waiting for the manifest ready, repo %v, reference:%v", art.Repository, getReference(art))
		time.Sleep(sleepIntervalSec * time.Second)
		newMan, err = m.updateManifestList(ctx, art.Repository, man)
	}
	if err != nil {
		return err
	}
	if len(art.Tag) > 0 && tagMoved(r, remoteRepo, art) {
		log.Debugf("the tag %v:%v doesn't point to %v anymore, push the manifest list by digest", art.Repository, art.Tag, art.Digest)
		art.Tag = ""
	}
	return m.push(ctx, art, newMan)
}

// tagMoved checks whether the tag of art points to another manifest in the remote registry
func tagMoved(r RemoteInterface, remoteRepo string, art lib.ArtifactInfo) bool {
	if r == nil {
		return false
	}
	exist, desc, err := r.ManifestExist(remoteRepo, art.Tag)
	if err != nil {
		log.Debugf("failed to check the tag %v:%v in the remote registry, error: %v", remoteRepo, art.Tag, err)
		return false
	}
	if !exist || desc == nil {
		return true
	}
	return len(desc.Digest) > 0 && string(desc.Digest) != art.Digest
}

// push pushes the trimmed manifest list by the tag of art, or by its own digest when art has no tag
func (m *ManifestListCache) push(ctx context.Context, art lib.ArtifactInfo, newMan distribution.Manifest) error {
	if len(newMan.References()) == 0 {
		return libErrors.New("manifest list doesn't contain any pushed manifest")
	}
	_, pl, err := newMan.Payload()
	if err != nil {
		log.Errorf("failed to get payload, error %v", err)
		return err
	}
	log.Debugf("The manifest list payload: %v", string(pl))
	newDig := string(digest.FromBytes(pl))
	m.cacheTrimmedDigest(ctx, art.Repository, art.Digest, newDig)
	reference := art.Tag
	if len(reference) == 0 {
		reference = newDig
	}
	// Because the manifest list maybe updated, need to recheck if it is exist in local
	a, err := m.local.GetManifest(ctx, lib.ArtifactInfo{Repository: art.Repository, Tag: reference})
	if err != nil {
		return err
	}
	if a == nil || a.Digest != newDig {
		if err := m.local.PushManifest(art.Repository, reference, newMan); err != nil {
			log.Errorf("failed to push manifest list, error: %v", err)
			return err
		}
		log.Debugf("push manifest list successfully, repository: %v, reference: %v, digest: %v", art.Repository, reference, newDig)
	}
	log.Debug("update artifact pull time to avoid it is removed by GC before the manifest list is pushed to local")
	artForPullTime := lib.ArtifactInfo{Repository: art.Repository, Tag: art.Tag, Digest: newDig}
	if err := m.local.UpdatePullTime(ctx, artForPullTime); err != nil {
		log.Errorf("failed to update pull time for artifact %v:%v, error: %v", art.Repository, newDig, err)
	}
	return nil
}

// ManifestCache default Manifest handler
type ManifestCache struct {
	local localInterface
}

// CacheContent ...
func (m *ManifestCache) CacheContent(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, r RemoteInterface, _ string) {
	var waitBlobs []distribution.Descriptor
	for n := range maxManifestWait {
		time.Sleep(sleepIntervalSec * time.Second)
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

	err := m.push(art, man)
	if err != nil {
		log.Errorf("error occurred on manifest push to local: %v", err)
	}
}

func (m *ManifestCache) push(art lib.ArtifactInfo, man distribution.Manifest) error {
	errs := []error{}
	if len(art.Digest) > 0 {
		err := m.local.PushManifest(art.Repository, art.Digest, man)
		if err != nil {
			log.Errorf("failed to push manifest referencing digest, tag: %v, digest: %v, error %v", art.Tag, art.Digest, err)
			errs = append(errs, err)
		}
	}
	if len(art.Tag) > 0 {
		err := m.local.PushManifest(art.Repository, art.Tag, man)
		if err != nil {
			log.Errorf("failed to push manifest referencing tag, tag: %v, digest: %v, error %v", art.Tag, art.Digest, err)
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
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
