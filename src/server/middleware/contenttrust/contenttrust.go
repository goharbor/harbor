package contenttrust

import (
	"fmt"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/pkg/signature"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
)

var (
	// isArtifactSigned use the sign manager to check the signature, it could handle pull by tag or digtest
	// if pull by digest, any tag of the artifact is signed, will return true.
	isArtifactSigned = func(req *http.Request, art lib.ArtifactInfo) (bool, error) {
		checker, err := signature.GetManager().GetCheckerByRepo(req.Context(), art.Repository)
		if err != nil {
			return false, err
		}
		return checker.IsArtifactSigned(art.Digest), nil
	}
)

// Middleware handle docker pull content trust check
func Middleware() func(http.Handler) http.Handler {
	return middleware.BeforeRequest(func(r *http.Request) error {
		ctx := r.Context()
		none := lib.ArtifactInfo{}
		af := lib.GetArtifactInfo(ctx)
		if af == none {
			return fmt.Errorf("artifactinfo middleware required before this middleware")
		}
		if len(af.Digest) == 0 {
			art, err := artifact.Ctl.GetByReference(ctx, af.Repository, af.Reference, nil)
			if err != nil {
				return err
			}
			af.Digest = art.Digest
		}
		pro, err := project.Ctl.GetByName(ctx, af.ProjectName)
		if err != nil {
			return err
		}
		securityCtx, ok := security.FromContext(ctx)
		// only authenticated robot account with scanner pull access can bypass.
		if ok && securityCtx.IsAuthenticated() &&
			(securityCtx.Name() == "robot" || securityCtx.Name() == "v2token") &&
			securityCtx.Can(rbac.ActionScannerPull, rbac.NewProjectNamespace(pro.ProjectID).Resource(rbac.ResourceRepository)) {
			// the artifact is pulling by the scanner, skip the checking
			logger.Debugf("artifact %s@%s is pulling by the scanner, skip the checking", af.Repository, af.Digest)
			return nil
		}

		if pro.ContentTrustEnabled() {
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
