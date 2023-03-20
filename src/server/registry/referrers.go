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

	// Check if the reference is a valid digest
	if _, err := digest.Parse(reference); err != nil {
		lib_http.SendError(w, errors.Wrapf(err, "unsupported digest %s", reference).WithCode(errors.BadRequestCode))
		return
	}

	result := &ocispec.Index{}

	// Get the artifact by reference
	art, err := r.artifactManager.GetByDigest(ctx, repository, reference)
	if err != nil {
		if errors.IsNotFoundErr(err) {
			// If artifact not found, return empty index
			newListReferrersOK().WithPayload(result).WriteResponse(w)
			return
		}
		lib_http.SendError(w, err)
		return
	}

	// Query accessories with matching subject artifact digest and artifactType
	query := q.New(q.KeyWords{"SubjectArtifactDigest": art.Digest})
	if at != "" {
		query = q.New(q.KeyWords{"SubjectArtifactDigest": art.Digest, "Type": at})
	}
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
	var mfs []ocispec.Descriptor
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
		mfs = append(mfs, mf)
	}

	// Populate index manifest
	result.SchemaVersion = ReferrersSchemaVersion
	result.MediaType = ReferrersMediaType
	result.Manifests = mfs

	// Write response with index manifest and headers
	baseAPI := &handler.BaseAPI{}
	newListReferrersOK().
		WithXTotalCount(total).
		WithLink(baseAPI.Links(ctx, req.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(result).WriteResponse(w)
}

type listReferrersOK struct {
	/*Link refers to the previous page and next page

	 */
	Link string `json:"Link"`
	/*The total count of accessories

	 */
	XTotalCount int64 `json:"X-Total-Count"`

	/*
	  In: Body
	*/
	Payload *ocispec.Index `json:"body,omitempty"`
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

// WithXTotalCount adds the xTotalCount to the list accessories o k response
func (o *listReferrersOK) WithXTotalCount(xTotalCount int64) *listReferrersOK {
	o.XTotalCount = xTotalCount
	return o
}

// WithPayload adds the payload to the list accessories o k response
func (o *listReferrersOK) WithPayload(payload *ocispec.Index) *listReferrersOK {
	o.Payload = payload
	return o
}

// WriteResponse to the client
func (o *listReferrersOK) WriteResponse(rw http.ResponseWriter) {
	rw.Header().Set("Content-Type", "application/json; charset=utf-8")

	link := o.Link
	if link != "" {
		rw.Header().Set("Link", link)
	}
	xTotalCount := swag.FormatInt64(o.XTotalCount)
	if xTotalCount != "" {
		rw.Header().Set("X-Total-Count", xTotalCount)
	}

	rw.WriteHeader(200)
	payload := o.Payload
	if payload == nil {
		// return empty index
		payload = &ocispec.Index{}
	}

	enc := json.NewEncoder(rw)
	if err := enc.Encode(payload); err != nil {
		lib_http.SendError(rw, err)
		return
	}
}
