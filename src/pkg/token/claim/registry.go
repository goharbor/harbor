package claim

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/docker/distribution/registry/auth/token"
)

// Registry implements the interface of jwt.Claims
type Registry struct {
	jwt.StandardClaims
	PolicyCheck bool                     `json:"policy_check"`
	Access      []*token.ResourceActions `json:"access"`
}

// Valid valid the standard claims
func (rc *Registry) Valid() error {
	stdErr := rc.StandardClaims.Valid()
	if stdErr != nil {
		return stdErr
	}
	return nil
}
