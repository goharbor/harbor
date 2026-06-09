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

package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-openapi/swag"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/lib/cache"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/pkg/cached/manifest/redis"
	"github.com/goharbor/harbor/src/pkg/registry"
	"github.com/goharbor/harbor/src/server/router"
)

const referrersSchemaVersion = 2
const referrersMediaType = "application/vnd.oci.image.index.v1+json"

// NewReferrersAPIHandler creates a new handler for the OCI referrers API
// exposed via the Harbor REST API (/api/v2.0/.../referrers).
func NewReferrersAPIHandler() http.Handler {
	return &referrersAPIHandler{
		BaseAPI:          &BaseAPI{},
		artifactManager:  artifact.NewManager(),
		accessoryManager: accessory.NewManager(),
		registryClient:   registry.Cli,
		maniCacheManager: redis.NewManager(),
	}
}

type referrersAPIHandler struct {
	*BaseAPI
	artifactManager  artifact.Manager
	accessoryManager accessory.Manager
	registryClient   registry.Client
	maniCacheManager redis.CachedManager
}

func (r *referrersAPIHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	projectName := router.Param(ctx, ":project_name")
	repoName := router.Param(ctx, ":repo_name")
	reference := router.Param(ctx, ":reference")

	// Check project-level permission
	if err := r.RequireProjectAccess(ctx, projectName, rbac.ActionList, rbac.ResourceAccessory); err != nil {
		lib_http.SendError(w, err)
		return
	}

	at := req.URL.Query().Get("artifactType")
	var filter string
	if at != "" {
		filter = "artifactType"
	}

	// Validate the reference is a valid digest
	if _, err := digest.Parse(reference); err != nil {
		lib_http.SendError(w, errors.Wrapf(err, "unsupported digest %s", reference).WithCode(errors.BadRequestCode))
		return
	}

	repository := fmt.Sprintf("%s/%s", projectName, repoName)

	// Query accessories by digest and repo (OCI referrer approach)
	query := q.New(q.KeyWords{"SubjectArtifactDigest": reference, "SubjectArtifactRepo": repository})
	total, err := r.accessoryManager.Count(ctx, query)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}
	accs, err := r.accessoryManager.List(ctx, query)
	if err != nil {
		lib_http.SendError(w, err)
		return
	}

	// Build OCI index manifest from accessories
	mfs := make([]ocispec.Descriptor, 0)
	for _, acc := range accs {
		accArtDigest := acc.GetData().Digest
		accArt, err := r.artifactManager.GetByDigest(ctx, repository, accArtDigest)
		if err != nil {
			lib_http.SendError(w, err)
			return
		}

		fromCache := false
		writeCache := false
		var maniContent []byte

		if config.CacheEnabled() {
			maniContent, err = r.maniCacheManager.Get(req.Context(), accArtDigest)
			if err == nil {
				fromCache = true
			} else {
				log.Debugf("failed to get manifest %s from cache, will fallback to registry, error: %v", accArtDigest, err)
				if errors.As(err, &cache.ErrNotFound) {
					writeCache = true
				}
			}
		}
		if !fromCache {
			mani, _, err := r.registryClient.PullManifest(accArt.RepositoryName, accArtDigest)
			if err != nil {
				lib_http.SendError(w, err)
				return
			}
			_, maniContent, err = mani.Payload()
			if err != nil {
				lib_http.SendError(w, err)
				return
			}
			if writeCache {
				err = r.maniCacheManager.Save(req.Context(), accArtDigest, maniContent)
				if err != nil {
					log.Warningf("failed to save manifest %s to cache, error: %v", accArtDigest, err)
				}
			}
		}

		desc := ocispec.Descriptor{
			MediaType:    accArt.ManifestMediaType,
			Size:         int64(len(maniContent)),
			Digest:       digest.Digest(accArt.Digest),
			Annotations:  accArt.Annotations,
			ArtifactType: accArt.ArtifactType,
		}
		if at != "" {
			if accArt.ArtifactType == at {
				mfs = append(mfs, desc)
			}
		} else {
			mfs = append(mfs, desc)
		}
	}

	// Build and return the OCI index
	result := &ocispec.Index{}
	result.SchemaVersion = referrersSchemaVersion
	result.MediaType = referrersMediaType
	result.Manifests = mfs

	// Write response headers
	w.Header().Set("Content-Type", referrersMediaType)
	if filter != "" {
		w.Header().Set("OCI-Filters-Applied", filter)
	}
	xTotalCount := swag.FormatInt64(total)
	if xTotalCount != "" {
		w.Header().Set("X-Total-Count", xTotalCount)
	}
	w.WriteHeader(http.StatusOK)

	enc := json.NewEncoder(w)
	if err := enc.Encode(result); err != nil {
		lib_http.SendError(w, err)
		return
	}
}
