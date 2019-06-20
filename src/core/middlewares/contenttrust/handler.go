package contenttrust

import (
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/common/utils/notary"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	"net/http"
	"strings"
)

var NotaryEndpoint = ""

type contentTrustHandler struct {
	next http.Handler
}

func New(next http.Handler) http.Handler {
	return &contentTrustHandler{
		next: next,
	}
}

func (cth contentTrustHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	imgRaw := req.Context().Value(util.ImageInfoCtxKey)
	if imgRaw == nil || !config.WithNotary() {
		cth.next.ServeHTTP(rw, req)
		return
	}
	img, _ := req.Context().Value(util.ImageInfoCtxKey).(util.ImageInfo)
	if img.Digest == "" {
		cth.next.ServeHTTP(rw, req)
		return
	}
	if !util.GetPolicyChecker().ContentTrustEnabled(img.ProjectName) {
		cth.next.ServeHTTP(rw, req)
		return
	}
	match, err := matchNotaryDigest(img)
	if err != nil {
		http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", "Failed in communication with Notary please check the log"), http.StatusInternalServerError)
		return
	}
	if !match {
		log.Debugf("digest mismatch, failing the response.")
		http.Error(rw, util.MarshalError("PROJECT_POLICY_VIOLATION", "The image is not signed in Notary."), http.StatusPreconditionFailed)
		return
	}
	cth.next.ServeHTTP(rw, req)
}

func matchNotaryDigest(img util.ImageInfo) (bool, error) {
	if NotaryEndpoint == "" {
		NotaryEndpoint = config.InternalNotaryEndpoint()
	}
	targets, err := notary.GetInternalTargets(NotaryEndpoint, util.TokenUsername, img.Repository)
	if err != nil {
		return false, err
	}
	for _, t := range targets {
		if isDigest(img.Reference) {
			d, err := notary.DigestFromTarget(t)
			if err != nil {
				return false, err
			}
			if img.Digest == d {
				return true, nil
			}
		} else {
			if t.Tag == img.Reference {
				log.Debugf("found reference: %s in notary, try to match digest.", img.Reference)
				d, err := notary.DigestFromTarget(t)
				if err != nil {
					return false, err
				}
				if img.Digest == d {
					return true, nil
				}
			}
		}
	}
	log.Debugf("image: %#v, not found in notary", img)
	return false, nil
}

// A sha256 is a string with 64 characters.
func isDigest(ref string) bool {
	return strings.HasPrefix(ref, "sha256:") && len(ref) == 71
}
