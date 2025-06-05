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
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/docker/distribution"
	"github.com/opencontainers/go-digest"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/blob"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	model_tag "github.com/goharbor/harbor/src/pkg/tag/model/tag"
)

const (
	// wait more time than manifest (maxManifestWait) because manifest list depends on manifest ready
	maxManifestListWait = 20
	maxManifestWait     = 10
	sleepIntervalSec    = 20
	// keep manifest list in cache for one week
	manifestListCacheInterval = 7 * 24 * 60 * 60 * time.Second
)

var (
	// Ctl is a global proxy controller instance
	ctl  Controller
	once sync.Once
)

// Controller defines the operations related with pull through proxy
type Controller interface {
	// UseLocalBlob check if the blob should use local copy
	UseLocalBlob(ctx context.Context, art lib.ArtifactInfo) bool
	// UseLocalManifest check manifest should use local copy
	UseLocalManifest(ctx context.Context, art lib.ArtifactInfo, remote RemoteInterface) (bool, *ManifestList, error)
	// ProxyBlob proxy the blob request to the remote server, p is the proxy project
	// art is the ArtifactInfo which includes the digest of the blob
	ProxyBlob(ctx context.Context, p *proModels.Project, art lib.ArtifactInfo) (int64, io.ReadCloser, error)
	// ProxyManifest proxy the manifest request to the remote server, p is the proxy project,
	// art is the ArtifactInfo which includes the tag or digest of the manifest
	ProxyManifest(ctx context.Context, art lib.ArtifactInfo, remote RemoteInterface) (distribution.Manifest, error)
	// HeadManifest send manifest head request to the remote server
	HeadManifest(ctx context.Context, art lib.ArtifactInfo, remote RemoteInterface) (bool, *distribution.Descriptor, error)
	// EnsureTag ensure tag for digest
	EnsureTag(ctx context.Context, art lib.ArtifactInfo, tagName string) error
}

type controller struct {
	blobCtl         blob.Controller
	artifactCtl     artifact.Controller
	local           localInterface
	cache           cache.Cache
	handlerRegistry map[string]ManifestCacheHandler
}

// ControllerInstance -- Get the proxy controller instance
func ControllerInstance() Controller {
	// Lazy load the controller
	// Because LocalHelper is not ready unless core startup completely
	once.Do(func() {
		l := newLocalHelper()
		ctl = &controller{
			blobCtl:         blob.Ctl,
			artifactCtl:     artifact.Ctl,
			local:           newLocalHelper(),
			cache:           cache.Default(),
			handlerRegistry: NewCacheHandlerRegistry(l),
		}
	})

	return ctl
}

func (c *controller) EnsureTag(ctx context.Context, art lib.ArtifactInfo, tagName string) error {
	// search the digest in cache and query with trimmed digest
	var trimmedDigest string
	err := c.cache.Fetch(ctx, TrimmedManifestlist+art.Digest, &trimmedDigest)
	if errors.Is(err, cache.ErrNotFound) { // nolint:revive
		// skip to update digest, continue
	} else if err != nil {
		// for other error, return
		return err
	} else {
		// found in redis, update the digest
		art.Digest = trimmedDigest
		log.Debugf("Found trimmed digest: %v", trimmedDigest)
	}
	a, err := c.local.GetManifest(ctx, art)
	if err != nil {
		return err
	}
	if a == nil {
		return fmt.Errorf("the artifact is not ready yet, failed to tag it to %v", tagName)
	}
	tagID, err := tag.Ctl.Ensure(ctx, a.RepositoryID, a.Artifact.ID, tagName)
	if err != nil {
		return err
	}
	// update the pull time of tag for the first time cache
	return tag.Ctl.Update(ctx, &tag.Tag{
		Tag: model_tag.Tag{
			ID:       tagID,
			PullTime: time.Now(),
		},
	}, "PullTime")
}

func (c *controller) UseLocalBlob(ctx context.Context, art lib.ArtifactInfo) bool {
	if len(art.Digest) == 0 {
		return false
	}
	exist, err := c.local.BlobExist(ctx, art)
	if err != nil {
		return false
	}
	return exist
}

// ManifestList ...
type ManifestList struct {
	Content     []byte
	Digest      string
	ContentType string
}

// UseLocalManifest check if these manifest could be found in local registry,
// the return error should be nil when it is not found in local and need to delegate to remote registry
// the return error should be NotFoundError when it is not found in remote registry
// the error will be captured by framework and return 404 to client
func (c *controller) UseLocalManifest(ctx context.Context, art lib.ArtifactInfo, remote RemoteInterface) (bool, *ManifestList, error) {
	a, err := c.local.GetManifest(ctx, art)
	if err != nil {
		return false, nil, err
	}
	// Pull by digest when artifact exist in local
	if a != nil && len(art.Digest) > 0 {
		return true, nil, nil
	}

	remoteRepo := getRemoteRepo(art)
	exist, desc, err := remote.ManifestExist(remoteRepo, getReference(art)) // HEAD
	if err != nil {
		if errors.IsRateLimitError(err) && a != nil { // if rate limit, use local if it exists, otherwise return error
			return true, nil, nil
		}
		return false, nil, err
	}
	if !exist || desc == nil {
		return false, nil, errors.NotFoundError(fmt.Errorf("repo %v, tag %v not found", art.Repository, art.Tag))
	}

	var content []byte
	var contentType string
	if c.cache == nil {
		return a != nil && string(desc.Digest) == a.Digest, nil, nil // digest matches
	}
	// Pass digest to the cache key, digest is more stable than tag, because tag could be updated
	if len(art.Digest) == 0 {
		art.Digest = string(desc.Digest)
	}
	err = c.cache.Fetch(ctx, manifestListKey(art.Repository, art), &content)
	if err != nil {
		if errors.Is(err, cache.ErrNotFound) {
			log.Debugf("Digest is not found in manifest list cache, key=cache:%v", manifestListKey(art.Repository, art))
		} else {
			log.Errorf("Failed to get manifest list from cache, error: %v", err)
		}
		return a != nil && string(desc.Digest) == a.Digest, nil, nil
	}
	err = c.cache.Fetch(ctx, manifestListContentTypeKey(art.Repository, art), &contentType)
	if err != nil {
		log.Debugf("failed to get the manifest list content type, not use local. error:%v", err)
		return false, nil, nil
	}
	log.Debugf("Get the manifest list with key=cache:%v", manifestListKey(art.Repository, art))
	return true, &ManifestList{content, string(desc.Digest), contentType}, nil
}

func manifestListKey(repo string, art lib.ArtifactInfo) string {
	// actual redis key format is cache:manifestlist:<repo name>:<tag> or cache:manifestlist:<repo name>:sha256:xxxx
	return "manifestlist:" + repo + ":" + getReference(art)
}

func manifestListContentTypeKey(rep string, art lib.ArtifactInfo) string {
	return manifestListKey(rep, art) + ":contenttype"
}

func (c *controller) ProxyManifest(ctx context.Context, art lib.ArtifactInfo, remote RemoteInterface) (distribution.Manifest, error) {
	var man distribution.Manifest
	remoteRepo := getRemoteRepo(art)
	ref := getReference(art)
	man, dig, err := remote.Manifest(remoteRepo, ref)
	if err != nil {
		return man, err
	}
	ct, _, err := man.Payload()
	if err != nil {
		return man, err
	}

	// Push manifest in background
	go func(operator string) {
		bCtx := orm.Copy(ctx)
		a, err := c.local.GetManifest(bCtx, art)
		if err != nil {
			log.Errorf("failed to get manifest, error %v", err)
		}
		// Push manifest to local when pull with digest, or artifact not found, or digest mismatch
		if len(art.Tag) == 0 || a == nil || a.Digest != dig {
			artInfo := art
			if len(artInfo.Digest) == 0 {
				artInfo.Digest = dig
			}
			c.waitAndPushManifest(bCtx, remoteRepo, man, artInfo, ct, remote)
		}

		// Query artifact after push
		if a == nil {
			a, err = c.local.GetManifest(bCtx, art)
			if err != nil {
				log.Errorf("failed to get manifest, error %v", err)
			}
		}
		if a != nil {
			SendPullEvent(bCtx, a, art.Tag, operator)
		}
	}(operator.FromContext(ctx))

	return man, nil
}

func (c *controller) HeadManifest(_ context.Context, art lib.ArtifactInfo, remote RemoteInterface) (bool, *distribution.Descriptor, error) {
	remoteRepo := getRemoteRepo(art)
	ref := getReference(art)
	return remote.ManifestExist(remoteRepo, ref)
}

func (c *controller) ProxyBlob(ctx context.Context, p *proModels.Project, art lib.ArtifactInfo) (int64, io.ReadCloser, error) {
	remoteRepo := getRemoteRepo(art)
	log.Debugf("The blob doesn't exist, proxy the request to the target server, url:%v", remoteRepo)
	rHelper, err := NewRemoteHelper(ctx, p.RegistryID, WithSpeed(p.ProxyCacheSpeed()))
	if err != nil {
		return 0, nil, err
	}

	size, bReader, err := rHelper.BlobReader(remoteRepo, art.Digest)
	if err != nil {
		log.Errorf("failed to pull blob, error %v", err)
		return 0, nil, err
	}
	desc := distribution.Descriptor{Size: size, Digest: digest.Digest(art.Digest)}
	go func() {
		err := c.putBlobToLocal(remoteRepo, art.Repository, desc, rHelper)
		if err != nil {
			log.Errorf("error while putting blob to local repo, %v", err)
		}
	}()
	return size, bReader, nil
}

func (c *controller) putBlobToLocal(remoteRepo string, localRepo string, desc distribution.Descriptor, r RemoteInterface) error {
	log.Debugf("Put blob to local registry!, sourceRepo:%v, localRepo:%v, digest: %v", remoteRepo, localRepo, desc.Digest)
	_, bReader, err := r.BlobReader(remoteRepo, string(desc.Digest))
	if err != nil {
		log.Errorf("failed to create blob reader, error %v", err)
		return err
	}
	defer bReader.Close()
	err = c.local.PushBlob(localRepo, desc, bReader)
	return err
}

func (c *controller) waitAndPushManifest(ctx context.Context, remoteRepo string, man distribution.Manifest, art lib.ArtifactInfo, contType string, r RemoteInterface) {
	h, ok := c.handlerRegistry[contType]
	if !ok {
		h, ok = c.handlerRegistry[defaultHandler]
		if !ok {
			return
		}
	}
	h.CacheContent(ctx, remoteRepo, man, art, r, contType)
}

// getRemoteRepo get the remote repository name, used in proxy cache
func getRemoteRepo(art lib.ArtifactInfo) string {
	return strings.TrimPrefix(art.Repository, art.ProjectName+"/")
}

func getReference(art lib.ArtifactInfo) string {
	if len(art.Digest) > 0 {
		return art.Digest
	}
	return art.Tag
}
