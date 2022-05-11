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

package registry

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/event/operator"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/server/router"
	"github.com/opencontainers/go-digest"
)

// https://github.com/distribution/distribution/blob/c202b9b0d7b79a67337dec8e1f1bafb1c7095315/registry/handlers/manifests.go#L280
func etagMatch(r *http.Request, etag string) bool {
	for _, headerVal := range r.Header["If-None-Match"] {
		if headerVal == etag || headerVal == fmt.Sprintf(`"%s"`, etag) { // allow quoted or unquoted
			return true
		}
	}
	return false
}

// make sure the artifact exist before proxying the request to the backend registry
func getManifest(w http.ResponseWriter, req *http.Request) {
	repository := router.Param(req.Context(), ":splat")
	reference := router.Param(req.Context(), ":reference")
	art, err := artifact.Ctl.GetByReference(req.Context(), repository, reference, nil)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	// the reference is tag, replace it with digest
	if _, err = digest.Parse(reference); err != nil {
		req = req.Clone(req.Context())
		req.URL.Path = strings.TrimSuffix(req.URL.Path, reference) + art.Digest
		req.URL.RawPath = req.URL.EscapedPath()
	}

	// if etag match, we can earlier return and no need to proxy to distribution
	// as we have stored digest in database.
	// https://github.com/distribution/distribution/blob/c202b9b0d7b79a67337dec8e1f1bafb1c7095315/registry/handlers/manifests.go#L135
	if etagMatch(req, art.Digest) {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	buffer := lib.NewResponseBuffer(w)
	// whether get manifest from cache
	fromCache := false
	// whether need to write-back cache
	wbCache := false
	// check cache
	if config.CacheEnabled() {
		manifest, err := pkg.ManifestMgr.Get(req.Context(), art.Digest)
		if err == nil {
			fromCache = true
			// write header
			buffer.Header().Set("Content-Length", fmt.Sprintf("%d", len(manifest)))
			buffer.Header().Set("Content-Type", art.ManifestMediaType)
			buffer.Header().Set("Docker-Distribution-Api-Version", "registry/2.0")
			buffer.Header().Set("Docker-Content-Digest", art.Digest)
			buffer.Header().Set("Etag", fmt.Sprintf(`"%s"`, art.Digest))
			buffer.WriteHeader(http.StatusOK)
			// write data from cache, no need to write body if is head request
			if req.Method == http.MethodGet {
				buffer.Write(manifest)
			}
		} else {
			log.Warningf("failed to get manifest from cache, error: %v", err)
			// only write cache when request is GET because HEAD request resp
			// body is empty.
			if req.Method == http.MethodGet {
				wbCache = true
			}
		}
	}
	// proxy to registry if not from cache
	if !fromCache {
		proxy.ServeHTTP(buffer, req)
	}
	// flush data
	if _, err = buffer.Flush(); err != nil {
		log.Errorf("failed to flush: %v", err)
		return
	}
	// return if not success
	if !buffer.Success() {
		return
	}
	// write back manifest to cache if needed
	if wbCache {
		if err = pkg.ManifestMgr.Save(req.Context(), art.Digest, buffer.Buffer()); err != nil {
			log.Warningf("failed to save manifest %s to cache, error: %v", art.Digest, err)
		}
	}

	// fire event, ignore the HEAD request and pulling request from replication service
	if req.Method == http.MethodHead || req.UserAgent() == registry.UserAgent {
		return
	}

	e := &metadata.PullArtifactEventMetadata{
		Artifact: &art.Artifact,
		Operator: operator.FromContext(req.Context()),
	}
	// the reference is tag
	if _, err = digest.Parse(reference); err != nil {
		e.Tag = reference
	}
	notification.AddEvent(req.Context(), e)
}

// just delete the artifact from database
func deleteManifest(w http.ResponseWriter, req *http.Request) {
	repository := router.Param(req.Context(), ":splat")
	reference := router.Param(req.Context(), ":reference")
	// v2 doesn't support delete by tag
	// add parse digest here is to return ErrDigestInvalidFormat before GetByReference throws an NOT_FOUND(404)
	// Do not add the logic into GetByReference as it's a shared method for PUT/GET/DELETE/Internal call,
	// and NOT_FOUND satisfy PUT/GET/Internal call.
	if _, err := digest.Parse(reference); err != nil {
		lib_http.SendError(w, errors.Wrapf(err, "unsupported digest %s", reference).WithCode(errors.UNSUPPORTED))
		return
	}
	art, err := artifact.Ctl.GetByReference(req.Context(), repository, reference, nil)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	if err = artifact.Ctl.Delete(req.Context(), art.ID); err != nil {
		lib_http.SendError(w, err)
		return
	}
	w.WriteHeader(http.StatusAccepted)

	// clean cache if enabled
	if config.CacheEnabled() {
		if err = pkg.ManifestMgr.Delete(req.Context(), art.Digest); err != nil {
			log.Errorf("failed to delete manifest cache: %v", err)
		}
	}
}

func putManifest(w http.ResponseWriter, req *http.Request) {
	repo := router.Param(req.Context(), ":splat")
	reference := router.Param(req.Context(), ":reference")

	// make sure the repository exist before pushing the manifest
	_, _, err := repository.Ctl.Ensure(req.Context(), repo)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	buffer := lib.NewResponseBuffer(w)
	// proxy the req to the backend docker registry
	proxy.ServeHTTP(buffer, req)
	if !buffer.Success() {
		if _, err := buffer.Flush(); err != nil {
			log.Errorf("failed to flush: %v", err)
		}
		return
	}

	// When got the response from the backend docker registry, the manifest and
	// tag are both ready, so we don't need to handle the issue anymore:
	// https://github.com/docker/distribution/issues/2625

	var tags []string
	dgt := reference
	// the reference is tag, get the digest from the response header
	if _, err = digest.Parse(reference); err != nil {
		dgt = buffer.Header().Get("Docker-Content-Digest")
		tags = append(tags, reference)
	}

	_, _, err = artifact.Ctl.Ensure(req.Context(), repo, dgt, &artifact.ArtOption{
		Tags: tags,
	})
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	// flush the origin response from the docker registry to the underlying response writer
	if _, err := buffer.Flush(); err != nil {
		log.Errorf("failed to flush: %v", err)
	}
}
