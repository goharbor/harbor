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

package subject

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/server/middleware"
)

var (
	// the media type of notation signature layer
	mediaTypeNotationLayer = "application/vnd.cncf.notary.signature"

	// cosign media type  in config layer, which would support in oci-spec1.1
	mediaTypeCosignConfig = "application/vnd.dev.cosign.artifact.sig.v1+json"
	// cosign media type in artifact type (New Format)
	mediaTypeCosignArtifactType = "application/vnd.dev.sigstore.bundle.v0.3+json"

	// annotation of nydus image
	layerAnnotationNydusBootstrap = "containerd.io/snapshot/nydus-bootstrap"

	// media type of harbor sbom
	mediaTypeHarborSBOM = "application/vnd.goharbor.harbor.sbom.v1"
)

/*
	{
	  "schemaVersion": 2,
	  "mediaType": "application/vnd.oci.image.manifest.v1+json",
	  "config": {
	    "mediaType": "application/vnd.oci.image.config.v1+json",
	    "size": 7023,
	    "digest": "sha256:b5b2b2c507a0944348e0303114d8d93aaaa081732b86451d9bce1f432a537bc7"
	  },
	  "layers": [
	    {
	      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
	      "size": 32654,
	      "digest": "sha256:9834876dcfb05cb167a5c24953eba58c4ac89b1adf57f28f2f9d09af107ee8f0"
	    },
	    {
	      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
	      "size": 16724,
	      "digest": "sha256:3c3a4604a545cdc127456d94e421cd355bca5b528f4a9c1905b15da2eb4a4c6b"
	    },
	    {
	      "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
	      "size": 73109,
	      "digest": "sha256:ec4b8955958665577945c89419d1af06b5f7636b4ac3da7f12184802ad867736"
	    }
	  ],
	  "subject": {
	    "mediaType": "application/vnd.oci.image.manifest.v1+json",
	    "size": 7682,
	    "digest": "sha256:5b0bcabd1ed22e9fb1310cf6c2dec7cdef19f0ad69efa1f392e94a4333501270"
	  },
	  "annotations": {
	    "com.example.key1": "value1",
	    "com.example.key2": "value2"
	  }
	}
*/
func Middleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		ctx := r.Context()
		logger := log.G(ctx).WithFields(log.Fields{"middleware": "subject"})

		none := lib.ArtifactInfo{}
		info := lib.GetArtifactInfo(ctx)
		if info == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			return err
		}

		mf := &ocispec.Manifest{}
		if err := json.Unmarshal(body, mf); err != nil {
			logger.Errorf("unmarshal manifest failed, error: %v", err)
			return err
		}

		/*
			when an images is pushed, it could be
				1. single image (do nothing)
				2. an accesory image
				3. a subject image
				4. both as an accessory and a subject image
			and a subject image or accessory image could be pushed in either order
		*/

		if mf.Subject != nil {
			subjectArt, err := artifact.Ctl.GetByReference(ctx, info.Repository, mf.Subject.Digest.String(), nil)
			if err != nil {
				if !errors.IsNotFoundErr(err) {
					logger.Errorf("failed to get subject artifact: %s, error: %v", mf.Subject.Digest, err)
					return err
				}
				log.Debug("the subject artifact doesn't exist.")
			}
			art, err := artifact.Ctl.GetByReference(ctx, info.Repository, info.Reference, nil)
			if err != nil {
				logger.Errorf("failed to get artifact with subject field: %s, error: %v", info.Reference, err)
				return err
			}
			accData := model.AccessoryData{
				ArtifactID:        art.ID,
				SubArtifactRepo:   info.Repository,
				SubArtifactDigest: mf.Subject.Digest.String(),
				Size:              art.Size,
				Digest:            art.Digest,
			}
			accData.Type = model.TypeSubject
			// since oci-spec 1.1, image type may from artifactType if presents, otherwise would be Config.MediaType
			fromType := mf.Config.MediaType
			if mf.ArtifactType != "" {
				fromType = mf.ArtifactType
			}
			switch fromType {
			case ocispec.MediaTypeImageConfig, schema2.MediaTypeImageConfig:
				if isNydusImage(mf) {
					accData.Type = model.TypeNydusAccelerator
				}
			case mediaTypeNotationLayer:
				accData.Type = model.TypeNotationSignature
			case mediaTypeCosignConfig, mediaTypeCosignArtifactType:
				accData.Type = model.TypeCosignSignature
			case mediaTypeHarborSBOM:
				accData.Type = model.TypeHarborSBOM
			}
			if subjectArt != nil {
				accData.SubArtifactID = subjectArt.ID
			}
			if err := orm.WithTransaction(func(ctx context.Context) error {
				_, err := accessory.Mgr.Create(ctx, accData)
				return err
			})(orm.SetTransactionOpNameToContext(ctx, "tx-create-subject-accessory")); err != nil {
				if !errors.IsConflictErr(err) {
					logger.Errorf("failed to create subject accessory artifact: %s, error: %v", art.Digest, err)
					return err
				}
			}

			// when subject artifact is pushed after accessory artifact, current subject artifact do not exist.
			// so we use reference manifest subject digest instead of subjectArt.Digest
			w.Header().Set("OCI-Subject", mf.Subject.Digest.String())
		}

		// check if images is a Subject artifact
		digest := digest.FromBytes(body)
		accs, err := accessory.Mgr.List(ctx, q.New(q.KeyWords{"SubjectArtifactDigest": digest, "SubjectArtifactRepo": info.Repository}))
		if err != nil {
			logger.Errorf("failed to list accessory artifact: %s, error: %v", digest, err)
			return err
		}
		if len(accs) > 0 {
			// In certain cases, the OCI client may push the subject artifact and accessory in either order.
			// Therefore, it is necessary to handle situations where the client pushes the accessory ahead of the subject artifact.
			art, err := artifact.Ctl.GetByReference(ctx, info.Repository, digest.String(), nil)
			if err != nil {
				logger.Errorf("failed to list artifact: %s, error: %v", digest, err)
				return err
			}
			if art != nil {
				for _, acc := range accs {
					accData := model.AccessoryData{
						ID:            acc.GetData().ID,
						SubArtifactID: art.ID,
					}
					if err := accessory.Mgr.Update(ctx, accData); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

// isNydusImage checks if the image is a nydus image.
func isNydusImage(manifest *ocispec.Manifest) bool {
	layers := manifest.Layers
	if len(layers) != 0 {
		desc := layers[len(layers)-1]
		if desc.Annotations == nil {
			return false
		}
		_, hasAnno := desc.Annotations[layerAnnotationNydusBootstrap]
		return hasAnno
	}
	return false
}
