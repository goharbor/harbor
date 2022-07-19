package contenttrust

import (
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/accessory/model"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/util"
	"net/http"
)

// Cosign handle docker pull content trust check
func Cosign() func(http.Handler) http.Handler {
	return middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()

		logger := log.G(ctx)

		none := lib.ArtifactInfo{}
		af := lib.GetArtifactInfo(ctx)
		if af == none {
			return errors.New("artifactinfo middleware required before this middleware").WithCode(errors.NotFoundCode)
		}
		pro, err := project.Ctl.GetByName(ctx, af.ProjectName)
		if err != nil {
			return err
		}

		// If cosign policy enabled, it has to at least have one cosign signature.
		if pro.ContentTrustCosignEnabled() {
			art, err := artifact.Ctl.GetByReference(ctx, af.Repository, af.Reference, &artifact.Option{
				WithAccessory: true,
			})
			if err != nil {
				return err
			}

			ok, err := util.SkipPolicyChecking(r, pro.ProjectID, art.ID)
			if err != nil {
				return err
			}
			if ok {
				logger.Debugf("artifact %s@%s is pulling by the scanner/cosign, skip the checking", af.Repository, af.Digest)
				return nil
			}

			if len(art.Accessories) == 0 {
				pkgE := errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed in Cosign.")
				return pkgE
			}

			var hasCosignSignature bool
			for _, acc := range art.Accessories {
				if acc.GetData().Type == model.TypeCosignSignature {
					hasCosignSignature = true
					break
				}
			}
			if !hasCosignSignature {
				pkgE := errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed in Cosign.")
				return pkgE
			}
		}

		return nil
	})
}
