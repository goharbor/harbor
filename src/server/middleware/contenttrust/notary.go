package contenttrust

import (
	"net/http"

	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/signature"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/goharbor/harbor/src/server/middleware/util"
)

var (
	// isArtifactSigned use the sign manager to check the signature, it could handle pull by tag or digtest
	// if pull by digest, any tag of the artifact is signed, will return true.
	isArtifactSigned = func(req *http.Request, art lib.ArtifactInfo) (bool, error) {
		checker, err := signature.GetManager().GetCheckerByRepo(req.Context(), art.Repository)
		if err != nil {
			return false, err
		}
		if len(art.Tag) > 0 {
			return checker.IsTagSigned(art.Tag, art.Digest), nil
		}
		return checker.IsArtifactSigned(art.Digest), nil
	}
)

// Notary handle docker pull content trust check
func Notary() func(http.Handler) http.Handler {
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
		if pro.ContentTrustEnabled() {
			art, err := artifact.Ctl.GetByReference(ctx, af.Repository, af.Reference, nil)
			if err != nil {
				return err
			}
			if len(af.Digest) == 0 {
				af.Digest = art.Digest
			}
			ok, err := util.SkipPolicyChecking(r, pro.ProjectID, art.ID)
			if err != nil {
				return err
			}
			if ok {
				logger.Debugf("artifact %s@%s is pulling by the scanner/cosign, skip the checking", af.Repository, af.Digest)
				return nil
			}

			match, err := isArtifactSigned(r, af)
			if err != nil {
				return err
			}
			if !match {
				pkgE := errors.New(nil).WithCode(errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed in Notary.")
				return pkgE
			}
		}
		return nil
	})
}
