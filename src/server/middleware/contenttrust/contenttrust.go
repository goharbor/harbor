package contenttrust

import (
	internal_errors "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/signature"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
	"net/http/httptest"
)

// NotaryEndpoint ...
var NotaryEndpoint = ""

// Middleware handle docker pull content trust check
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			doContentTrustCheck, mf := validate(req)
			if !doContentTrustCheck {
				next.ServeHTTP(rw, req)
				return
			}
			rec := httptest.NewRecorder()
			next.ServeHTTP(rec, req)
			if rec.Result().StatusCode == http.StatusOK {
				match, err := isArtifactSigned(req, mf)
				if err != nil {
					serror.SendError(rw, err)
					return
				}
				if !match {
					pkgE := internal_errors.New(nil).WithCode(internal_errors.PROJECTPOLICYVIOLATION).WithMessage("The image is not signed in Notary.")
					serror.SendError(rw, pkgE)
					return
				}
			}
			middleware.CopyResp(rec, rw)
		})
	}
}

func validate(req *http.Request) (bool, middleware.ArtifactInfo) {
	none := middleware.ArtifactInfo{}
	if err := middleware.EnsureArtifactDigest(req.Context()); err != nil {
		return false, none
	}
	af, ok := middleware.ArtifactInfoFromContext(req.Context())
	if !ok {
		return false, none
	}
	if !middleware.GetPolicyChecker().ContentTrustEnabled(af.ProjectName) {
		return false, af
	}
	return true, af
}

// isArtifactSigned use the sign manager to check the signature, it could handle pull by tag or digtest
// if pull by digest, any tag of the artifact is signed, will return true.
func isArtifactSigned(req *http.Request, art middleware.ArtifactInfo) (bool, error) {
	checker, err := signature.GetManager().GetCheckerByRepo(req.Context(), art.Repository)
	if err != nil {
		return false, err
	}
	return checker.IsArtifactSigned(art.Digest), nil
}
