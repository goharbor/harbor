package contenttrust

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/notary"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	internal_errors "github.com/goharbor/harbor/src/internal/error"
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
				match, err := matchNotaryDigest(mf)
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

func validate(req *http.Request) (bool, *middleware.ManifestInfo) {
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		return false, nil
	}
	_, err := mf.ManifestExists(req.Context())
	if err != nil {
		return false, mf
	}
	if scannerPull, ok := middleware.ScannerPullFromContext(req.Context()); ok && scannerPull {
		return false, mf
	}
	if !middleware.GetPolicyChecker().ContentTrustEnabled(mf.ProjectName) {
		return false, mf
	}
	return true, mf
}

func matchNotaryDigest(mf *middleware.ManifestInfo) (bool, error) {
	if NotaryEndpoint == "" {
		NotaryEndpoint = config.InternalNotaryEndpoint()
	}
	targets, err := notary.GetInternalTargets(NotaryEndpoint, util.TokenUsername, mf.Repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if mf.Digest != "" {
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			if mf.Digest == d {
				return true, nil
			}
		} else {
			if t.Tag == mf.Tag {
				log.Debugf("found reference: %s in notary, try to match digest.", mf.Tag)
				d, err := notary.DigestFromTarget(t)
				if err != nil {
					return false, err
				}
				if mf.Digest == d {
					return true, nil
				}
			}
		}
	}
	log.Debugf("image: %#v, not found in notary", mf)
	return false, nil
}
