package regtoken

import (
	"github.com/docker/distribution/registry/auth"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/middlewares/util"
	pkg_token "github.com/goharbor/harbor/src/pkg/token"
	"github.com/goharbor/harbor/src/pkg/token/claims/registry"
	"net/http"
	"strings"
)

// regTokenHandler is responsible for decoding the registry token in the docker pull request header,
// as harbor adds customized claims action into registry auth token, the middlerware is for decode it and write it into
// request context, then for other middlerwares in chain to use it to bypass request validation.
type regTokenHandler struct {
	next http.Handler
}

// New ...
func New(next http.Handler) http.Handler {
	return &regTokenHandler{
		next: next,
	}
}

// ServeHTTP ...
func (r *regTokenHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	img, _ := util.ImageInfoFromContext(req.Context())
	if img.Digest == "" {
		r.next.ServeHTTP(rw, req)
		return
	}

	parts := strings.Split(req.Header.Get("Authorization"), " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		r.next.ServeHTTP(rw, req)
		return
	}
	rawToken := parts[1]
	opt := pkg_token.DefaultTokenOptions()
	regTK, err := pkg_token.Parse(opt, rawToken, &registry.Claim{})
	if err != nil {
		log.Errorf("failed to decode reg token: %v, the error is skipped and round the request to native registry.", err)
		r.next.ServeHTTP(rw, req)
		return
	}

	accessItems := []auth.Access{}
	accessItems = append(accessItems, auth.Access{
		Resource: auth.Resource{
			Type: rbac.ResourceRepository.String(),
			Name: img.Repository,
		},
		Action: rbac.ActionScannerPull.String(),
	})

	accessSet := regTK.Claims.(*registry.Claim).GetAccess()
	for _, access := range accessItems {
		if accessSet.Contains(access) {
			*req = *(req.WithContext(util.NewBypassPolicyCheckContext(req.Context(), true)))
		}
	}
	r.next.ServeHTTP(rw, req)
}
