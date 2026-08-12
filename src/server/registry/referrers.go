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
	"encoding/json"
	"net/http"

	"github.com/go-openapi/swag"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

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
	"github.com/goharbor/harbor/src/server/v2.0/handler"
)

const ReferrersSchemaVersion = 2
const ReferrersMediaType = "application/vnd.oci.image.index.v1+json"

func newReferrersHandler() http.Handler {
	return &referrersHandler{
		artifactManager:  artifact.NewManager(),
		accessoryManager: accessory.NewManager(),
		registryClient:   registry.Cli,
		maniCacheManager: redis.NewManager(),
	}
}

type referrersHandler struct {
	artifactManager  artifact.Manager
	accessoryManager accessory.Manager
	registryClient   registry.Client
	maniCacheManager redis.CachedManager
}

func (r *referrersHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()

	repository := router.Param(ctx, ":splat")
	reference := router.Param(ctx, ":reference")
	at := req.URL.Query().Get("artifactType")
	var filter string
	if at != "" {
		filter = "artifactType"
	}

	// Check if the reference is a valid digest
	if _, err := digest.Parse(reference); err != nil {
		lib_http.SendError(w, errors.Wrapf(err, "unsupported digest %s", reference).WithCode(errors.BadRequestCode))
		return
	}

	// Query accessories with matching subject artifact digest
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
	// Build index manifest from accessories
	mfs := make([]ocispec.Descriptor, 0)
	for _, acc := range accs {
		accArtDigest := acc.GetData().Digest
		accArt, err := r.artifactManager.GetByDigest(ctx, repository, accArtDigest)
		if err != nil {
			lib_http.SendError(w, err)
			return
		}
		// whether get manifest from cache
		fromCache := false
		// whether need write manifest to cache
		writeCache := false
		var maniContent []byte

		// pull manifest, will try to pull from cache first
		// and write to cache when pull manifest from registry at first time
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
			// write manifest to cache when first time pulling
			if writeCache {
				err = r.maniCacheManager.Save(req.Context(), accArtDigest, maniContent)
				if err != nil {
					log.Warningf("failed to save accArt manifest %s to cache, error: %v", accArtDigest, err)
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
		// filter use accArt.ArtifactType as artifactType
		if at != "" {
			if accArt.ArtifactType == at {
				mfs = append(mfs, desc)
			}
		} else {
			mfs = append(mfs, desc)
		}
	}

	// Populate index manifest
	result := &ocispec.Index{}
	result.SchemaVersion = ReferrersSchemaVersion
	result.MediaType = ReferrersMediaType
	result.Manifests = mfs

	// Write response with index manifest and headers
	baseAPI := &handler.BaseAPI{}
	newListReferrersOK().
		WithXTotalCount(total).
		WithFilter(filter).
		WithLink(baseAPI.Links(ctx, req.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(result).WriteResponse(w)
}

type listReferrersOK struct {
	/*Link refers to the previous page and next page

	 */
	Link string `json:"Link"`
	/*Filter refers to the filter used to fetch the referrers

	 */
	Filter string `json:"Filter"`
	/*The total count of accessories

	 */
	XTotalCount int64 `json:"X-Total-Count"`

	/*
	  In: Body
	*/
	Payload any `json:"body,omitempty"`
}

// newListReferrersOK creates newlistReferrersOK with default headers values
func newListReferrersOK() *listReferrersOK {
	return &listReferrersOK{}
}

// WithLink adds the link to the get referrers o k response
func (o *listReferrersOK) WithLink(link string) *listReferrersOK {
	o.Link = link
	return o
}

// WithFilter adds the filter to the get referrers
func (o *listReferrersOK) WithFilter(filter string) *listReferrersOK {
	o.Filter = filter
	return o
}

// WithXTotalCount adds the xTotalCount to the list accessories o k response
func (o *listReferrersOK) WithXTotalCount(xTotalCount int64) *listReferrersOK {
	o.XTotalCount = xTotalCount
	return o
}

// WithPayload adds the payload to the list accessories o k response
func (o *listReferrersOK) WithPayload(payload any) *listReferrersOK {
	o.Payload = payload
	return o
}

// WriteResponse to the client
func (o *listReferrersOK) WriteResponse(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/vnd.oci.image.index.v1+json")

	link := o.Link
	if link != "" {
		rw.Header().Set("Link", link)
	}
	filter := o.Filter
	if filter != "" {
		rw.Header().Set("OCI-Filters-Applied", filter)
	}
	xTotalCount := swag.FormatInt64(o.XTotalCount)
	if xTotalCount != "" {
		rw.Header().Set("X-Total-Count", xTotalCount)
	}

	rw.WriteHeader(http.StatusOK)
	payload := o.Payload
	if payload == nil {
		// return empty index
		payload = struct{}{}
	}

	enc := json.NewEncoder(rw)
	if err := enc.Encode(payload); err != nil {
		lib_http.SendError(rw, err)
		return
	}
}
