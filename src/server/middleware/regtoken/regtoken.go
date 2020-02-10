package regtoken

import (
	"errors"
	"github.com/docker/distribution/registry/auth"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	pkg_token "github.com/goharbor/harbor/src/pkg/token"
	"github.com/goharbor/harbor/src/pkg/token/claims/registry"
	serror "github.com/goharbor/harbor/src/server/error"
	"github.com/goharbor/harbor/src/server/middleware"
	"net/http"
	"strings"
)

// Middleware parses the docker pull bearer token and check whether it's a scanner pull.
func Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			err := parseToken(req)
			if err != nil {
				serror.SendError(rw, err)
				return
			}
			next.ServeHTTP(rw, req)
		})
	}
}

func parseToken(req *http.Request) error {
	mf, ok := middleware.ManifestInfoFromContext(req.Context())
	if !ok {
		return errors.New("cannot get the manifest information from request context")
	}

	parts := strings.Split(req.Header.Get("Authorization"), " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil
	}

	rawToken := parts[1]
	opt := pkg_token.DefaultTokenOptions()
	regTK, err := pkg_token.Parse(opt, rawToken, &registry.Claim{})
	if err != nil {
		log.Errorf("failed to decode reg token: %v, the error is skipped and round the request to native registry.", err)
		return nil
	}

	accessItems := []auth.Access{}
	accessItems = append(accessItems, auth.Access{
		Resource: auth.Resource{
			Type: rbac.ResourceRepository.String(),
			Name: mf.Repository,
		},
		Action: rbac.ActionScannerPull.String(),
	})

	accessSet := regTK.Claims.(*registry.Claim).GetAccess()
	for _, access := range accessItems {
		if accessSet.Contains(access) {
			*req = *(req.WithContext(middleware.NewScannerPullContext(req.Context(), true)))
		}
	}

	return nil
}
