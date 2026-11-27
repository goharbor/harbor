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
	"slices"
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

const (
	defaultHandler = "default"
	// tagCheckTimeout bounds the checks of the tags of a manifest list upstream, which hold the lock of the repository
	tagCheckTimeout = 30 * time.Second
	// maxListLength is the number of manifest lists recorded as waiting for one manifest,
	// and of tags recorded for one manifest list
	maxListLength = 10
)

// NewCacheHandlerRegistry ...
func NewCacheHandlerRegistry(local localInterface) map[string]ManifestCacheHandler {
	manListHandler := &ManifestListCache{
		local:  local,
		cache:  libCache.Default(),
		locker: newRedisLocker(),
	}
	manHandler := &ManifestCache{local: local, lists: manListHandler}
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
	cache  libCache.Cache
	local  localInterface
	locker locker
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
	if err := m.waitAndPush(ctx, remoteRepo, art, man, r, maxManifestListWait); err != nil {
		log.Errorf("error when push manifest list to local :%v", err)
	}
}

// trimmedDigestKey is the key of the digest of the manifest list pushed to repo for the manifest list dig
func trimmedDigestKey(repo, dig string) string {
	return TrimmedManifestlist + repo + ":" + dig
}

// manifestListParentsKey is the key of the digests of the manifest lists in repo that wait for the manifest dig to be cached
func manifestListParentsKey(repo, dig string) string {
	return "manifestlistparents:" + repo + ":" + dig
}

// manifestListTagsKey is the key of the tags the manifest list dig was pulled by in repo
func manifestListTagsKey(repo, dig string) string {
	return "manifestlisttags:" + repo + ":" + dig
}

// manifestListLockKey is the key of the lock to push the manifest lists of repo
func manifestListLockKey(repo string) string {
	return "{proxy_manifest_list_lock}:" + repo
}

// cacheTrimmedDigest caches the digest newDig of the manifest list pushed to repo for the manifest list dig,
// for controller.EnsureTag and controller.manifestListTagged
func (m *ManifestListCache) cacheTrimmedDigest(ctx context.Context, repo, dig, newDig string) {
	if m.cache == nil {
		return
	}
	key := trimmedDigestKey(repo, dig)
	if err := m.cache.Save(ctx, key, newDig, manifestListCacheInterval); err != nil {
		log.Warningf("failed to cache the trimmed manifest, err %v", err)
		return
	}
	log.Debugf("Saved key:%v, value:%v", key, newDig)
}

// waitForManifests records the manifest list art as one that waits for each manifest of man to be cached,
// and the tag of art as one to push it by. It runs under the lock of the repository,
// so concurrent publications don't overwrite each other's records.
func (m *ManifestListCache) waitForManifests(ctx context.Context, art lib.ArtifactInfo, man distribution.Manifest) {
	if m.cache == nil {
		return
	}
	for _, ref := range man.References() {
		if err := m.addToList(ctx, manifestListParentsKey(art.Repository, string(ref.Digest)), art.Digest); err != nil {
			log.Warningf("failed to record the manifest list %v@%v of the manifest %v, error: %v", art.Repository, art.Digest, ref.Digest, err)
		}
	}
	if len(art.Tag) > 0 {
		if err := m.addToList(ctx, manifestListTagsKey(art.Repository, art.Digest), art.Tag); err != nil {
			log.Warningf("failed to record the tag %v of the manifest list %v@%v, error: %v", art.Tag, art.Repository, art.Digest, err)
		}
	}
}

// addToList adds value as the last item of the list cached with key, and keeps its last maxListLength items
func (m *ManifestListCache) addToList(ctx context.Context, key, value string) error {
	var list []string
	// the Redis cache wraps the errors of failed reads in ErrNotFound: only a missing key returns it as is
	if err := m.cache.Fetch(ctx, key, &list); err != nil && err != libCache.ErrNotFound {
		return err
	}
	list = append(slices.DeleteFunc(list, func(v string) bool { return v == value }), value)
	if len(list) > maxListLength {
		list = list[len(list)-maxListLength:]
	}
	return m.cache.Save(ctx, key, list, manifestListCacheInterval)
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

// pushCached pushes the manifest list art cached with contentType and content, without waiting for its manifests
func (m *ManifestListCache) pushCached(ctx context.Context, remoteRepo string, art lib.ArtifactInfo, content []byte, contentType string, r RemoteInterface) error {
	man, _, err := distribution.UnmarshalManifest(contentType, content)
	if err != nil {
		return err
	}
	return m.waitAndPush(ctx, remoteRepo, art, man, r, 0)
}

// pushParents pushes again the manifest lists that wait for the manifest dig, now that it is cached in repo
func (m *ManifestListCache) pushParents(ctx context.Context, remoteRepo, repo, dig string, r RemoteInterface) error {
	if m.cache == nil {
		return nil
	}
	var parents []string
	if err := m.cache.Fetch(ctx, manifestListParentsKey(repo, dig), &parents); err != nil {
		if err == libCache.ErrNotFound {
			return nil
		}
		return err
	}
	var errs []error
	for _, parent := range parents {
		if err := m.pushParent(ctx, remoteRepo, lib.ArtifactInfo{Repository: repo, Digest: parent}, r); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// pushParent pushes again the cached manifest list art. It doesn't skip when this core instance already publishes
// art: that publication may have checked its manifests before the last one was cached.
func (m *ManifestListCache) pushParent(ctx context.Context, remoteRepo string, art lib.ArtifactInfo, r RemoteInterface) error {
	var content []byte
	var contentType string
	if err := m.cache.Fetch(ctx, manifestListKey(art.Repository, art), &content); err != nil {
		log.Debugf("the manifest list %v@%v isn't cached anymore, error: %v", art.Repository, art.Digest, err)
		return nil
	}
	if err := m.cache.Fetch(ctx, manifestListContentTypeKey(art.Repository, art), &contentType); err != nil {
		log.Debugf("the content type of the manifest list %v@%v isn't cached anymore, error: %v", art.Repository, art.Digest, err)
		return nil
	}
	man, _, err := distribution.UnmarshalManifest(contentType, content)
	if err != nil {
		return err
	}
	return m.publish(ctx, remoteRepo, art, man, r)
}

// waitAndPush waits up to maxWait times for the manifests of man to be cached, then pushes the manifest list
func (m *ManifestListCache) waitAndPush(ctx context.Context, remoteRepo string, art lib.ArtifactInfo, man distribution.Manifest, r RemoteInterface, maxWait int) error {
	// skip when this core instance already waits for the same manifest list
	key := manifestListKey(art.Repository, art) + ":" + art.Tag
	if !inflightChecker.addRequest(key) {
		return nil
	}
	defer inflightChecker.removeRequest(key)
	// For manifest list, it might include some different manifest
	// it checks maxWait times every sleepIntervalSec, if all depend manifests are ready then push it
	// if time exceed, then push a updated manifest list which contains existing manifest
	for range maxWait {
		newMan, err := m.updateManifestList(ctx, art.Repository, man)
		if err != nil {
			return err
		}
		if len(newMan.References()) == len(man.References()) {
			break
		}
		log.Debugf("waiting for the manifest ready, repo %v, reference:%v", art.Repository, getReference(art))
		time.Sleep(sleepIntervalSec * time.Second)
	}
	return m.publish(ctx, remoteRepo, art, man, r)
}

// publish pushes the manifests of man cached locally as a manifest list. Core instances publish the manifest lists
// of a repository one at a time, so the last push reflects the latest local and upstream state.
func (m *ManifestListCache) publish(ctx context.Context, remoteRepo string, art lib.ArtifactInfo, man distribution.Manifest, r RemoteInterface) error {
	// the operations below stop when the lock is lost
	ctx, unlock, err := m.locker.lock(ctx, manifestListLockKey(art.Repository))
	if err != nil {
		return err
	}
	defer unlock()
	// record it before checking which manifests are cached, so caching any other manifest publishes it again
	m.waitForManifests(ctx, art, man)
	newMan, err := m.updateManifestList(ctx, art.Repository, man)
	if err != nil {
		return err
	}
	if len(newMan.References()) == 0 {
		log.Debugf("none of the manifests of the manifest list %v@%v is cached yet", art.Repository, art.Digest)
		return nil
	}
	return m.push(ctx, remoteRepo, art, newMan, r)
}

// tagPointsTo checks whether the tag points to the manifest dig in the remote registry.
// It returns false when it can't be confirmed, so a manifest list never replaces a newer one.
// It gives up when ctx is done, as the remote registry client can't be canceled.
func tagPointsTo(ctx context.Context, r RemoteInterface, remoteRepo, tag, dig string) bool {
	if r == nil || ctx.Err() != nil {
		return false
	}
	type result struct {
		exist bool
		desc  *distribution.Descriptor
		err   error
	}
	ch := make(chan result, 1)
	go func() {
		exist, desc, err := r.ManifestExist(remoteRepo, tag)
		ch <- result{exist, desc, err}
	}()
	select {
	case res := <-ch:
		if res.err != nil {
			log.Debugf("failed to check the tag %v:%v in the remote registry, error: %v", remoteRepo, tag, res.err)
			return false
		}
		return res.exist && res.desc != nil && string(res.desc.Digest) == dig
	case <-ctx.Done():
		log.Debugf("gave up checking the tag %v:%v in the remote registry, error: %v", remoteRepo, tag, ctx.Err())
		return false
	}
}

// tags returns the tags to push the manifest list art.Digest by: the tag of art, the tags it was pulled by,
// and the tags of the copy of it pushed before, unless that is newDig
func (m *ManifestListCache) tags(ctx context.Context, art lib.ArtifactInfo, newDig string) ([]string, error) {
	var tags []string
	add := func(names ...string) {
		for _, name := range names {
			if len(name) > 0 && !slices.Contains(tags, name) {
				tags = append(tags, name)
			}
		}
	}
	add(art.Tag)
	if m.cache == nil {
		return tags, nil
	}
	var pulledBy []string
	if err := m.cache.Fetch(ctx, manifestListTagsKey(art.Repository, art.Digest), &pulledBy); err == nil {
		add(pulledBy...)
	}
	var prev string
	if m.cache.Fetch(ctx, trimmedDigestKey(art.Repository, art.Digest), &prev) != nil || prev == newDig {
		return tags, nil
	}
	prevTags, err := m.local.GetTags(ctx, art.Repository, prev)
	if err != nil {
		return nil, err
	}
	add(prevTags...)
	return tags, nil
}

// push pushes newMan, the manifests of the manifest list art.Digest cached locally, by the tags of tags() that still
// point to art.Digest upstream. Without such tag, it pushes newMan by its own digest.
func (m *ManifestListCache) push(ctx context.Context, remoteRepo string, art lib.ArtifactInfo, newMan distribution.Manifest, r RemoteInterface) error {
	_, pl, err := newMan.Payload()
	if err != nil {
		log.Errorf("failed to get payload, error %v", err)
		return err
	}
	log.Debugf("The manifest list payload: %v", string(pl))
	newDig := string(digest.FromBytes(pl))
	tags, err := m.tags(ctx, art, newDig)
	if err != nil {
		return err
	}
	// the checks of the tags upstream hold the lock of the repository, bound them
	checkCtx, cancel := context.WithTimeout(ctx, tagCheckTimeout)
	defer cancel()
	var pushed []string
	var errs []error
	for _, tag := range tags {
		a, err := m.local.GetManifest(ctx, lib.ArtifactInfo{Repository: art.Repository, Tag: tag})
		if err != nil {
			errs = append(errs, err)
			continue
		}
		if a == nil || a.Digest != newDig {
			if !tagPointsTo(checkCtx, r, remoteRepo, tag, art.Digest) {
				log.Debugf("the tag %v:%v isn't confirmed to point to %v upstream, skip it", art.Repository, tag, art.Digest)
				continue
			}
			if err := m.pushManifest(ctx, art.Repository, tag, newDig, newMan); err != nil {
				errs = append(errs, err)
				continue
			}
		}
		pushed = append(pushed, tag)
	}
	if len(pushed) == 0 {
		a, err := m.local.GetManifest(ctx, lib.ArtifactInfo{Repository: art.Repository, Digest: newDig})
		if err == nil && a == nil {
			err = m.pushManifest(ctx, art.Repository, newDig, newDig, newMan)
		}
		if err != nil {
			return errors.Join(append(errs, err)...)
		}
	}
	// newDig exists locally now, even when some tags failed: record it, so EnsureTag doesn't move tags back
	m.cacheTrimmedDigest(ctx, art.Repository, art.Digest, newDig)
	log.Debug("update artifact pull time to avoid it is removed by GC before the manifest list is pushed to local")
	artForPullTime := lib.ArtifactInfo{Repository: art.Repository, Digest: newDig}
	if slices.Contains(pushed, art.Tag) {
		artForPullTime.Tag = art.Tag
	}
	if err := m.local.UpdatePullTime(ctx, artForPullTime); err != nil {
		log.Errorf("failed to update pull time for artifact %v:%v, error: %v", art.Repository, newDig, err)
	}
	return errors.Join(errs...)
}

func (m *ManifestListCache) pushManifest(ctx context.Context, repo, reference, dig string, man distribution.Manifest) error {
	// don't push when the lock was lost, e.g. while checking the tag upstream
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := m.local.PushManifest(repo, reference, man); err != nil {
		log.Errorf("failed to push manifest list, error: %v", err)
		return err
	}
	log.Debugf("push manifest list successfully, repository: %v, reference: %v, digest: %v", repo, reference, dig)
	return nil
}

// ManifestCache default Manifest handler
type ManifestCache struct {
	local localInterface
	// lists pushes again the manifest lists that wait for the cached manifests
	lists *ManifestListCache
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

	if err := m.push(art, man); err != nil {
		log.Errorf("error occurred on manifest push to local: %v", err)
		return
	}
	if m.lists != nil && len(art.Digest) > 0 {
		if err := m.lists.pushParents(ctx, remoteRepo, art.Repository, art.Digest, r); err != nil {
			log.Errorf("failed to push the manifest lists waiting for %v@%v, error: %v", art.Repository, art.Digest, err)
		}
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
