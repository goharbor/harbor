package nydus

import (
	"context"
	"encoding/json"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"io/ioutil"
	"net/http"
)

var (
	// the media type of consign signature layer
	mediaTypeNydusLayer = "application/vnd.oci.image.layer.nydus.blob.v1"
)

// NydusMiddleware middleware to record the linkage of artifact and its accessory
func NydusMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		ctx := r.Context()
		logger := log.G(ctx).WithFields(log.Fields{"middleware": "nydus"})

		none := lib.ArtifactInfo{}
		info := lib.GetArtifactInfo(ctx)
		if info == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}
		if info.Tag == "" {
			return nil
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		manifest, desc, err := distribution.UnmarshalManifest(contentType, body)
		if err != nil {
			logger.Errorf("unmarshal manifest failed, error: %v", err)
			return err
		}

		var isNydus bool
		for _, descriptor := range manifest.References() {
			if descriptor.MediaType == mediaTypeNydusLayer {
				isNydus = true
				break
			}
		}

		_, content, err := manifest.Payload()
		if err != nil {
			return err
		}

		// get manifest
		mani := &v1.Manifest{}
		if err := json.Unmarshal(content, mani); err != nil {
			return err
		}

		if isNydus {
			subjectArt, err := artifact.Ctl.GetByReference(ctx, info.Repository, mani.Annotations["io.goharbor.artifact.v1alpha1.acceleration.source.digest"], nil)
			if err != nil {
				logger.Errorf("failed to get subject artifact: %s, error: %v", info.Tag, err)
				return err
			}
			art, err := artifact.Ctl.GetByReference(ctx, info.Repository, desc.Digest.String(), nil)
			if err != nil {
				logger.Errorf("failed to get cosign signature artifact: %s, error: %v", desc.Digest.String(), err)
				return err
			}

			if err := orm.WithTransaction(func(ctx context.Context) error {
				_, err := accessory.Mgr.Create(ctx, model.AccessoryData{
					ArtifactID:    art.ID,
					SubArtifactID: subjectArt.ID,
					Size:          desc.Size,
					Digest:        desc.Digest.String(),
					Type:          model.TypeAccelNydus,
				})
				return err
			})(orm.SetTransactionOpNameToContext(ctx, "tx-create-nydus-accessory")); err != nil {
				if !errors.IsConflictErr(err) {
					logger.Errorf("failed to create cosign signature artifact: %s, error: %v", desc.Digest.String(), err)
					return err
				}
			}
		}

		return nil
	})
}
