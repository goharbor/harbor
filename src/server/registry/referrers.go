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

	"github.com/goharbor/harbor/src/lib/errors"
	lib_http "github.com/goharbor/harbor/src/lib/http"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/artifact"
	"github.com/goharbor/harbor/src/server/router"
	"github.com/goharbor/harbor/src/server/v2.0/handler"
)

const ReferrersSchemaVersion = 2
const ReferrersMediaType = "application/vnd.oci.image.index.v1+json"

func newReferrersHandler() http.Handler {
	return &referrersHandler{
		artifactManager:  artifact.NewManager(),
		accessoryManager: accessory.NewManager(),
	}
}

type referrersHandler struct {
	artifactManager  artifact.Manager
	accessoryManager accessory.Manager
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
		accArt, err := r.artifactManager.GetByDigest(ctx, repository, acc.GetData().Digest)
		if err != nil {
			lib_http.SendError(w, err)
			return
		}
		mf := ocispec.Descriptor{
			MediaType:    accArt.ManifestMediaType,
			Size:         accArt.Size,
			Digest:       digest.Digest(accArt.Digest),
			Annotations:  accArt.Annotations,
			ArtifactType: accArt.MediaType,
		}
		// filter by the artifactType since the artifactType is actually the config media type of the artifact.
		if at != "" {
			if accArt.MediaType == at {
				mfs = append(mfs, mf)
			}
		} else {
			mfs = append(mfs, mf)
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
	Payload interface{} `json:"body,omitempty"`
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
func (o *listReferrersOK) WithPayload(payload interface{}) *listReferrersOK {
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

	rw.WriteHeader(200)
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
