package security

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	registry_token "github.com/docker/distribution/registry/auth/token"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/v2token"
	svc_token "github.com/goharbor/harbor/src/core/service/token"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/token"
)

type v2TokenClaims struct {
	jwt.StandardClaims
	Access []*registry_token.ResourceActions `json:"access"`
}

func (vtc *v2TokenClaims) Valid() error {
	if err := vtc.StandardClaims.Valid(); err != nil {
		return err
	}
	if !vtc.VerifyAudience(svc_token.Registry, true) {
		return fmt.Errorf("invalid token audience: %s", vtc.Audience)
	}
	if !vtc.VerifyIssuer(svc_token.Issuer, true) {
		return fmt.Errorf("invalid token issuer: %s", vtc.Issuer)
	}
	return nil
}

type v2Token struct{}

func (vt *v2Token) Generate(req *http.Request) security.Context {
	logger := log.G(req.Context())
	if !strings.HasPrefix(req.URL.Path, "/v2") {
		return nil
	}
	tokenStr := bearerToken(req)
	if len(tokenStr) == 0 {
		return nil
	}

	opt := token.DefaultTokenOptions()
	cl := &v2TokenClaims{}
	t, err := token.Parse(opt, tokenStr, cl)
	if err != nil {
		logger.Warningf("failed to decode bearer token: %v", err)
		return nil
	}
	if err := t.Claims.Valid(); err != nil {
		logger.Warningf("failed to decode bearer token: %v", err)
		return nil
	}
	claims, ok := t.Claims.(*v2TokenClaims)
	if !ok {
		logger.Warningf("invalid token claims.")
		return nil
	}
	return v2token.New(req.Context(), claims.Subject, claims.Access)
}
