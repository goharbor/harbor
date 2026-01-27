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

package repoproxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib"
	libCache "github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	httpLib "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/redis"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/proxy/connection"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/server/middleware"
)

const (
	contentLength                  = "Content-Length"
	contentType                    = "Content-Type"
	dockerContentDigest            = "Docker-Content-Digest"
	etag                           = "Etag"
	ensureTagInterval              = 10 * time.Second
	ensureTagMaxRetry              = 60
	upstreamRegistryLimitOnProject = "UPSTREAM_REGISTRY_LIMIT_ON_PROJECT" // if UPSTREAM_REGISTRY_LIMIT_ON_PROJECT is true, the upstream registry connection is based on project level, by default it is artifact level
	referrerCacheTTL               = 7 * 24 * time.Hour
	link                           = "Link"
	xTotalCount                    = "X-Total-Count"
)

var proxyHeaderNames = []string{
	contentType,
	link,
	xTotalCount,
}

var tooManyRequestsError = errors.New("too many requests to upstream registry").WithCode(errors.RateLimitCode)

// BlobGetMiddleware handle get blob request
func BlobGetMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if err := handleBlob(w, r, next); err != nil {
			httpLib.SendError(w, err)
		}
	})
}

func handleBlob(w http.ResponseWriter, r *http.Request, next http.Handler) error {
	ctx := r.Context()
	art, p, proxyCtl, err := preCheck(ctx, true)
	if err != nil {
		return err
	}

	// go will panic if the network connection is ever interrupted, despite recovering successfully
	// it will return an error type of http.ErrAbortHandler if it recovers. Middleware is expected to suppress this.
	// golang/go#28239
	defer func() {
		r := recover()
		err, ok := r.(error)

		switch {
		case r == nil:
			return

		case ok && errors.Is(err, http.ErrAbortHandler):
			log.Debugf("Suppressed HttpAbortHandler: %s", err)
			return // suppress

		default:
			panic(r)
		}
	}()

	// Handle dockerhub request without library prefix
	isDefault, name, err := defaultLibrary(ctx, p.RegistryID, art)
	if err != nil {
		return err
	}
	if isDefault {
		http.Redirect(w, r, defaultBlobURL(p.Name, name, art.Digest), http.StatusMovedPermanently)
		return nil
	}

	if !canProxy(r.Context(), p) || proxyCtl.UseLocalBlob(ctx, art) {
		next.ServeHTTP(w, r)
		return nil
	}

	if p.MaxUpstreamConnection() > 0 {
		client, err := redis.GetHarborClient()
		if err != nil {
			return errors.NewErrs(err)
		}
		key := upstreamRegistryConnectionKey(art)
		log.Debugf("handle blob, upstream registry connection limit key: %s", key)
		if !connection.Limiter.Acquire(ctx, client, key, p.MaxUpstreamConnection()) {
			log.Infof("current connection exceed max connections to upstream registry")
			// send http code 429 to client
			return tooManyRequestsError
		}
		defer connection.Limiter.Release(context.Background(), client, key) // use background context in defer to avoid been canceled
	}

	size, reader, err := proxyCtl.ProxyBlob(ctx, p, art)
	if err != nil {
		return err
	}
	defer reader.Close()
	// Use io.CopyN to avoid out of memory when pulling big blob
	written, err := io.CopyN(w, reader, size)
	if err != nil {
		return err
	}
	if written != size {
		return errors.Errorf("the size mismatch, actual:%d, expected: %d", written, size)
	}
	setHeaders(w, size, "", art.Digest)
	return nil
}

func preCheck(ctx context.Context, withProjectMetadata bool) (art lib.ArtifactInfo, p *proModels.Project, ctl proxy.Controller, err error) {
	none := lib.ArtifactInfo{}
	art = lib.GetArtifactInfo(ctx)
	if art == none {
		return none, nil, nil, errors.New("artifactinfo is not found").WithCode(errors.NotFoundCode)
	}
	ctl = proxy.ControllerInstance()
	p, err = project.Ctl.GetByName(ctx, art.ProjectName, project.Metadata(withProjectMetadata))
	return
}

// ManifestMiddleware middleware handle request for get or head manifest
func ManifestMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		if err := handleManifest(w, r, next); err != nil {
			if errors.IsNotFoundErr(err) {
				httpLib.SendError(w, err)
				return
			}
			if errors.IsRateLimitError(err) {
				httpLib.SendError(w, tooManyRequestsError)
				return
			}
			log.Errorf("failed to proxy manifest, fallback to local, request uri: %v, error: %v", r.RequestURI, err)
			next.ServeHTTP(w, r)
		}
	})
}

func defaultLibrary(ctx context.Context, registryID int64, a lib.ArtifactInfo) (bool, string, error) {
	if registryID <= 0 {
		return false, "", nil
	}
	reg, err := registry.Ctl.Get(ctx, registryID)
	if err != nil {
		return false, "", err
	}
	if reg.Type != model.RegistryTypeDockerHub {
		return false, "", nil
	}
	name := strings.TrimPrefix(a.Repository, a.ProjectName+"/")
	if strings.Contains(name, "/") {
		return false, "", nil
	}
	return true, name, nil
}

// defaultManifestURL return the real url for request with default project
func defaultManifestURL(projectName string, name string, a lib.ArtifactInfo) string {
	return fmt.Sprintf("/v2/%s/library/%s/manifests/%s", projectName, name, a.Reference)
}

// defaultBlobURL return the real url for request with default project
func defaultBlobURL(projectName string, name string, digest string) string {
	return fmt.Sprintf("/v2/%s/library/%s/blobs/%s", projectName, name, digest)
}

// upstreamRegistryConnectionKey get upstream registry connection key
func upstreamRegistryConnectionKey(art lib.ArtifactInfo) string {
	limitOnProject := os.Getenv(upstreamRegistryLimitOnProject)
	if strings.EqualFold("true", limitOnProject) {
		return fmt.Sprintf("{upstream_registry_connection}:%s", art.ProjectName)
	}
	return fmt.Sprintf("{upstream_registry_connection}:%s:%s", art.Repository, art.Digest)
}

func handleManifest(w http.ResponseWriter, r *http.Request, next http.Handler) error {
	ctx := r.Context()
	art, p, proxyCtl, err := preCheck(ctx, true)
	if err != nil {
		return err
	}

	// Handle dockerhub request without library prefix
	defaultProj, name, err := defaultLibrary(ctx, p.RegistryID, art)
	if err != nil {
		return err
	}
	if defaultProj {
		http.Redirect(w, r, defaultManifestURL(p.Name, name, art), http.StatusMovedPermanently)
		return nil
	}

	if !canProxy(r.Context(), p) {
		next.ServeHTTP(w, r)
		return nil
	}
	remote, err := proxy.NewRemoteHelper(r.Context(), p.RegistryID, proxy.WithSpeed(p.ProxyCacheSpeed()))
	if err != nil {
		return err
	}
	useLocal, man, err := proxyCtl.UseLocalManifest(ctx, art, remote)
	if err != nil {
		return err
	}
	if useLocal {
		if man != nil {
			w.Header().Set(contentLength, fmt.Sprintf("%v", len(man.Content)))
			w.Header().Set(contentType, man.ContentType)
			w.Header().Set(dockerContentDigest, man.Digest)
			w.Header().Set(etag, man.Digest)
			if r.Method == http.MethodGet {
				_, err = w.Write(man.Content)
				if err != nil {
					return err
				}
			}
			return nil
		}
		next.ServeHTTP(w, r)
		return nil
	}
	if p.MaxUpstreamConnection() > 0 {
		client, err := redis.GetHarborClient()
		if err != nil {
			return errors.NewErrs(err)
		}
		key := upstreamRegistryConnectionKey(art)
		log.Debugf("handle manifest key %v", key)
		if !connection.Limiter.Acquire(ctx, client, key, p.MaxUpstreamConnection()) {
			log.Infof("current connection exceed max connections to upstream registry")
			// send http code 429 to client
			return tooManyRequestsError
		}
		defer connection.Limiter.Release(context.Background(), client, key) // use background context in defer to avoid been canceled
	}

	log.Debugf("the tag is %v, digest is %v", art.Tag, art.Digest)
	if r.Method == http.MethodHead {
		err = proxyManifestHead(ctx, w, proxyCtl, p, art, remote)
	} else if r.Method == http.MethodGet {
		log.Warningf("Artifact: %v:%v, digest:%v is not found in proxy cache, fetch it from remote repo", art.Repository, art.Tag, art.Digest)
		err = proxyManifestGet(ctx, w, proxyCtl, p, art, remote)
	}
	if err != nil {
		if errors.IsNotFoundErr(err) || errors.IsRateLimitError(err) {
			return err
		}
		log.Warningf("Proxy to remote failed, fallback to local repo, error: %v", err)
		next.ServeHTTP(w, r)
	}
	return nil
}

func proxyManifestGet(ctx context.Context, w http.ResponseWriter, ctl proxy.Controller, _ *proModels.Project, art lib.ArtifactInfo, remote proxy.RemoteInterface) error {
	man, err := ctl.ProxyManifest(ctx, art, remote)
	if err != nil {
		return err
	}
	ct, payload, err := man.Payload()
	if err != nil {
		return err
	}
	setHeaders(w, int64(len(payload)), ct, art.Digest)
	if _, err = w.Write(payload); err != nil {
		return err
	}
	return nil
}

func canProxy(ctx context.Context, p *proModels.Project) bool {
	if p.RegistryID < 1 {
		return false
	}
	reg, err := registry.Ctl.Get(ctx, p.RegistryID)
	if err != nil {
		log.Errorf("failed to get registry, error:%v", err)
		return false
	}
	if reg.Status != model.Healthy {
		log.Errorf("current registry is unhealthy, regID:%v, Name:%v, Status: %v", reg.ID, reg.Name, reg.Status)
	}
	return reg.Status == model.Healthy
}

func setHeaders(w http.ResponseWriter, size int64, mediaType string, dig string) {
	h := w.Header()
	h.Set(contentLength, fmt.Sprintf("%v", size))
	if len(mediaType) > 0 {
		h.Set(contentType, mediaType)
	}
	h.Set(dockerContentDigest, dig)
	h.Set(etag, dig)
}

// isProxySession check if current security context is proxy session
func isProxySession(ctx context.Context, projectName string) bool {
	sc, ok := security.FromContext(ctx)
	if !ok {
		log.Error("Failed to get security context")
		return false
	}
	username := sc.GetUsername()
	if username == proxycachesecret.ProxyCacheService {
		return true
	}
	// it should include the auto generate SBOM session, so that it could generate SBOM accessory in proxy cache project
	robotPrefix := config.RobotPrefix(ctx)
	scannerPrefix := config.ScannerRobotPrefix(ctx)
	prefix := fmt.Sprintf("%s%s+%s", robotPrefix, projectName, scannerPrefix)
	return strings.HasPrefix(username, prefix)
}

// DisableBlobAndManifestUploadMiddleware disable push artifact to a proxy project with a non-proxy session
func DisableBlobAndManifestUploadMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		ctx := r.Context()
		art := lib.GetArtifactInfo(ctx)
		p, err := project.Ctl.GetByName(ctx, art.ProjectName)
		if err != nil {
			httpLib.SendError(w, err)
			return
		}
		if p.IsProxy() && !isProxySession(ctx, art.ProjectName) {
			httpLib.SendError(w,
				errors.DeniedError(
					errors.Errorf("can not push artifact to a proxy project: %v", p.Name)))
			return
		}
		next.ServeHTTP(w, r)
	})
}

func proxyManifestHead(ctx context.Context, w http.ResponseWriter, ctl proxy.Controller, _ *proModels.Project, art lib.ArtifactInfo, remote proxy.RemoteInterface) error {
	exist, desc, err := ctl.HeadManifest(ctx, art, remote)
	if err != nil {
		return err
	}
	if !exist || desc == nil {
		return errors.NotFoundError(fmt.Errorf("the tag %v:%v is not found", art.Repository, art.Tag))
	}
	go func(art lib.ArtifactInfo) {
		// After docker 20.10 or containerd, the client heads the tag first,
		// Then GET the image by digest, in order to associate the tag with the digest
		// Ensure tag after head request, make sure tags in proxy cache keep update
		bCtx := orm.Context()
		for range ensureTagMaxRetry {
			time.Sleep(ensureTagInterval)
			bArt := lib.ArtifactInfo{ProjectName: art.ProjectName, Repository: art.Repository, Digest: string(desc.Digest)}
			err := ctl.EnsureTag(bCtx, bArt, art.Tag)
			if err == nil {
				return
			}
			log.Debugf("Failed to ensure tag %+v , error %v", art, err)
		}
	}(art)

	w.Header().Set(contentType, desc.MediaType)
	w.Header().Set(contentLength, fmt.Sprintf("%v", desc.Size))
	w.Header().Set(dockerContentDigest, string(desc.Digest))
	w.Header().Set(etag, string(desc.Digest))
	return nil
}

func ProxyReferrerMiddleware() func(http.Handler) http.Handler {
	return middleware.New(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
		ctx := r.Context()
		art := lib.GetArtifactInfo(ctx)
		p, err := project.Ctl.GetByName(ctx, art.ProjectName)
		if err != nil {
			httpLib.SendError(w, err)
			return
		}
		if !p.IsProxy() {
			next.ServeHTTP(w, r)
			return
		}

		if !p.ProxyReferrerAPI() {
			log.Debug("the proxy referrer API is not enabled for current project, fallback to local registry")
			next.ServeHTTP(w, r)
			return
		}

		log.Debug("current project is a harbor proxy cache project, will proxy the referrer API to the upstream")
		remote, err := proxy.NewRemoteHelper(r.Context(), p.RegistryID, proxy.WithSpeed(p.ProxyCacheSpeed()))
		if err != nil {
			log.Errorf("failed to proxy the referrer API to upstream, error %v, fallback to local registry", err)
			next.ServeHTTP(w, r)
			return
		}
		err = proxyReferrerGet(r, w, art, remote, p.RegistryID)
		// for any error, fallback to local registry
		if err != nil {
			log.Errorf("failed to proxy the referrer API to upstream, error %v, fallback to local registry", err)
			next.ServeHTTP(w, r)
		}
	})
}

// referrerCache defines the cache structure for referrer index
type referrerCache struct {
	Content []byte              `json:"content"`
	Header  map[string][]string `json:"header"`
}

// WriteProxyHeaders writes proxy-allowed headers from headerMap to response writer
func WriteProxyHeaders(w http.ResponseWriter, headerMap map[string][]string) {
	for k, v := range headerMap {
		if slices.Contains(proxyHeaderNames, k) {
			for _, hv := range v {
				w.Header().Add(k, hv)
			}
		}
	}
}

// referrerCacheKey generates cache key for referrer index
func referrerCacheKey(requestURI string) string {
	return fmt.Sprintf("{referrer_cache}:%s", requestURI)
}

func proxyReferrerGet(r *http.Request, w http.ResponseWriter, art lib.ArtifactInfo, remote proxy.RemoteInterface, registryID int64) error {
	key := referrerCacheKey(r.RequestURI)
	c := libCache.Default()
	if c == nil {
		return fmt.Errorf("the global cache is not initialized")
	}
	ctx := r.Context()
	// get registry status
	reg, err := registry.Ctl.Get(ctx, registryID)
	if err != nil {
		return err
	}
	// if registry is unhealthy, try to get from cache
	if reg.Status != model.Healthy {
		cachedContent := referrerCache{}
		err = c.Fetch(ctx, key, &cachedContent)
		if err != nil {
			log.Debugf("failed to get referrer index from cache, key: %s, error: %v", key, err)
			return fmt.Errorf("failed to get referrer index from cache")
		}
		if len(cachedContent.Content) == 0 {
			log.Debugf("no cached referrer index found in cache, key: %s", key)
			return fmt.Errorf("the upstream registry is unhealthy or no cached referrer index found")
		}

		log.Debugf("referrer index found in cache, key: %s", key)
		WriteProxyHeaders(w, cachedContent.Header)
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, bytes.NewReader(cachedContent.Content)) // return cached content if registry is unhealthy
		if err != nil {
			// just log the error, as the response has been written
			log.Errorf("failed to copy response body from cache, %v", err)
		}
		return nil
	}

	index, headerMap, err := remote.ListReferrers(proxy.GetRemoteRepo(art), art.Reference, r.URL.RawQuery)
	// if upstream return 404, return not found error
	if errors.IsNotFoundErr(err) {
		httpLib.SendError(w, err)
		return nil
	}

	if err != nil {
		return err
	}
	log.Debugf("current headers from upstream registry: %v", headerMap)
	WriteProxyHeaders(w, headerMap)
	w.WriteHeader(http.StatusOK)
	b, err := json.Marshal(index)
	if err != nil {
		return err
	}
	cacheContent := referrerCache{
		Content: b,
		Header:  headerMap,
	}
	// write to response
	// cache the index into the global cache
	log.Debugf("save referrer index to cache, key: %s", key)
	if err := c.Save(ctx, key, cacheContent, referrerCacheTTL); err != nil {
		log.Warningf("failed to save referrer index to cache, key: %s, error: %v", key, err)
	}

	_, err = io.Copy(w, bytes.NewReader(b))
	if err != nil {
		// just log the error, as the response has been written
		log.Errorf("failed to copy response body, %v", err)
	}
	return nil
}
