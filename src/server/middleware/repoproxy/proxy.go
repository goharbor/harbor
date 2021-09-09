//  Copyright Project Harbor Authors
//
//  Licensed under the Apache License, Version 2.0 (the "License");
//  you may not use this file except in compliance with the License.
//  You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
//  Unless required by applicable law or agreed to in writing, software
//  distributed under the License is distributed on an "AS IS" BASIS,
//  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//  See the License for the specific language governing permissions and
//  limitations under the License.

package repoproxy

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/proxycachesecret"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/proxy"
	"github.com/goharbor/harbor/src/controller/registry"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	httpLib "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	proModels "github.com/goharbor/harbor/src/pkg/project/models"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	"github.com/goharbor/harbor/src/server/middleware"
)

const (
	contentLength       = "Content-Length"
	contentType         = "Content-Type"
	dockerContentDigest = "Docker-Content-Digest"
	etag                = "Etag"
	ensureTagInterval   = 10 * time.Second
	ensureTagMaxRetry   = 60
)

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
	art, p, proxyCtl, err := preCheck(ctx)
	if err != nil {
		return err
	}
	if !canProxy(r.Context(), p) || proxyCtl.UseLocalBlob(ctx, art) {
		next.ServeHTTP(w, r)
		return nil
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
		return errors.Errorf("The size mismatch, actual:%d, expected: %d", written, size)
	}
	setHeaders(w, size, "", art.Digest)
	return nil
}

func preCheck(ctx context.Context) (art lib.ArtifactInfo, p *proModels.Project, ctl proxy.Controller, err error) {
	none := lib.ArtifactInfo{}
	art = lib.GetArtifactInfo(ctx)
	if art == none {
		return none, nil, nil, errors.New("artifactinfo is not found").WithCode(errors.NotFoundCode)
	}
	ctl = proxy.ControllerInstance()
	p, err = project.Ctl.GetByName(ctx, art.ProjectName, project.Metadata(false))
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
			log.Errorf("failed to proxy manifest, fallback to local, request uri: %v, error: %v", r.RequestURI, err)
			next.ServeHTTP(w, r)
		}
	})
}

func handleManifest(w http.ResponseWriter, r *http.Request, next http.Handler) error {
	ctx := r.Context()
	art, p, proxyCtl, err := preCheck(ctx)
	if err != nil {
		return err
	}
	if !canProxy(r.Context(), p) {
		next.ServeHTTP(w, r)
		return nil
	}
	remote, err := proxy.NewRemoteHelper(r.Context(), p.RegistryID)
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
				w.Write(man.Content)
			}
			return nil
		}
		next.ServeHTTP(w, r)
		return nil
	}

	log.Warningf("Artifact: %v:%v, digest:%v is not found in proxy cache, fetch it from remote repo", art.Repository, art.Tag, art.Digest)

	log.Debugf("the tag is %v, digest is %v", art.Tag, art.Digest)
	if r.Method == http.MethodHead {
		err = proxyManifestHead(ctx, w, proxyCtl, p, art, remote)
	} else if r.Method == http.MethodGet {
		err = proxyManifestGet(ctx, w, proxyCtl, p, art, remote)
	}
	if err != nil {
		if errors.IsNotFoundErr(err) {
			return err
		}
		log.Warningf("Proxy to remote failed, fallback to local repo, error: %v", err)
		next.ServeHTTP(w, r)
	}
	return nil
}

func proxyManifestGet(ctx context.Context, w http.ResponseWriter, ctl proxy.Controller, p *proModels.Project, art lib.ArtifactInfo, remote proxy.RemoteInterface) error {
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
func isProxySession(ctx context.Context) bool {
	sc, ok := security.FromContext(ctx)
	if !ok {
		log.Error("Failed to get security context")
		return false
	}
	if sc.GetUsername() == proxycachesecret.ProxyCacheService {
		return true
	}
	return false
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
		if p.IsProxy() && !isProxySession(ctx) {
			httpLib.SendError(w,
				errors.DeniedError(
					errors.Errorf("can not push artifact to a proxy project: %v", p.Name)))
			return
		}
		next.ServeHTTP(w, r)
		return
	})
}

func proxyManifestHead(ctx context.Context, w http.ResponseWriter, ctl proxy.Controller, p *proModels.Project, art lib.ArtifactInfo, remote proxy.RemoteInterface) error {
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
		for i := 0; i < ensureTagMaxRetry; i++ {
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
