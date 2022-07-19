package nydus

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

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
)

var (
	// nydus boostrap layer annotation
	nydusBoostrapAnnotation = "containerd.io/snapshot/nydus-bootstrap"

	// source artifact digest annotation
	sourceDigestAnnotation = "io.goharbor.artifact.v1alpha1.acceleration.source.digest"
)

// NydusAcceleratorMiddleware middleware to record the linkeage of artifact and its accessory
/*
/v2/library/hello-world/manifests/sha256:f54a58bc1aac5ea1a25d796ae155dc228b3f0e11d046ae276b39c4bf2f13d8c4
{
    "schemaVersion": 2,
    "config": {
      "mediaType": "application/vnd.oci.image.config.v1+json",
      "digest": "sha256:f7d0778a3c468a5203e95a9efd4d67ecef0d2a04866bb3320f0d5d637812aaee",
      "size": 466
    },
    "layers": [
      {
        "mediaType": "application/vnd.oci.image.layer.nydus.blob.v1",
        "digest": "sha256:fd9923a8e2bdc53747dbba3311be876a1deff4658785830e6030c5a8287acf74 ",
        "size": 3011,
        "annotations": {
          "containerd.io/snapshot/nydus-blob": "true"
        }
      },
      {
        "mediaType": "application/vnd.oci.image.layer.v1.tar+gzip",
        "digest": "sha256:d49bf6d7db9dac935b99d4c2c846b0d280f550aae62012f888d5a6e3ca59a589",
        "size": 459,
        "annotations": {
            "containerd.io/snapshot/nydus-blob-ids": "[\"fd9923a8e2bdc53747dbba3311be876a1deff4658785830e6030c5a8287acf74\"]",
            "containerd.io/snapshot/nydus-bootstrap": "true",
            "containerd.io/snapshot/nydus-rafs-version": "5"
        }
      }
    ],
    "annotations": {
        "io.goharbor.artifact.v1alpha1.acceleration.driver.name":"nydus",
        "io.goharbor.artifact.v1alpha1.acceleration.driver.version":"5",
        "io.goharbor.artifact.v1alpha1.acceleration.source.digest":"sha256:f54a58bc1aac5ea1a25d796ae155dc228b3f0e11d046ae276b39c4bf2f13d8c4"
    }
}

*/
func AcceleratorMiddleware() func(http.Handler) http.Handler {
	return middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		log.Debug("Start NydusAccelerator Middleware")
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
			annotationMap := descriptor.Annotations
			if _, ok := annotationMap[nydusBoostrapAnnotation]; ok {
				isNydus = true
				break
			}
		}
		log.Debug("isNydus: ", isNydus)

		_, payload, err := manifest.Payload()
		if err != nil {
			return err
		}
		mf := &v1.Manifest{}
		if err := json.Unmarshal(payload, mf); err != nil {
			return err
		}

		if isNydus {
			subjectArt, err := artifact.Ctl.GetByReference(ctx, info.Repository, mf.Annotations[sourceDigestAnnotation], nil)
			if err != nil {
				logger.Errorf("failed to get subject artifact: %s, error: %v", info.Tag, err)
				return err
			}
			art, err := artifact.Ctl.GetByReference(ctx, info.Repository, desc.Digest.String(), nil)
			if err != nil {
				logger.Errorf("failed to get nydus accel accelerator: %s, error: %v", desc.Digest.String(), err)
				return err
			}

			if err := orm.WithTransaction(func(ctx context.Context) error {
				id, err := accessory.Mgr.Create(ctx, model.AccessoryData{
					ArtifactID:    art.ID,
					SubArtifactID: subjectArt.ID,
					Size:          desc.Size,
					Digest:        desc.Digest.String(),
					Type:          model.TypeNydusAccelerator,
				})
				log.Debug("accessory id:", id)
				return err
			})(orm.SetTransactionOpNameToContext(ctx, "tx-create-nydus-accessory")); err != nil {
				if !errors.IsConflictErr(err) {
					logger.Errorf("failed to create nydus accelerator artifact: %s, error: %v", desc.Digest.String(), err)
					return err
				}
			}
		}

		return nil
	})
}
