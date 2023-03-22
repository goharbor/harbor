package subject

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/pkg/accessory"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/server/middleware"
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

		if mf.Subject != nil {
			subjectArt, err := artifact.Ctl.GetByReference(ctx, info.Repository, mf.Subject.Digest.String(), nil)
			if err != nil {
				if !errors.IsNotFoundErr(err) {
					logger.Errorf("failed to get subject artifact: %s, error: %v", mf.Subject.Digest, err)
					return err
				}
				log.Debug("the subject of the signature doesn't exist.")
			}
			art, err := artifact.Ctl.GetByReference(ctx, info.Repository, info.Reference, nil)
			if err != nil {
				logger.Errorf("failed to get artifact with subject field: %s, error: %v", info.Reference, err)
				return err
			}
			accData := model.AccessoryData{
				ArtifactID:        art.ID,
				SubArtifactDigest: mf.Subject.Digest.String(),
				Size:              art.Size,
				Digest:            art.Digest,
				Type:              model.TypeSubject,
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
		}

		return nil
	})
}
