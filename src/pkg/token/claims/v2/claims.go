package v2

import (
	"crypto/subtle"
	"fmt"

	"github.com/docker/distribution/registry/auth/token"
	"github.com/golang-jwt/jwt/v4"
)

func init() {
	jwt.MarshalSingleStringAsArray = false
}

const (
	// Issuer is the only valid issuer for jwt token sent to /v2/xxxx
	Issuer = "harbor-token-issuer"
)

// Claims represents the token claims that encapsulated in a JWT token for registry/notary resources
type Claims struct {
	jwt.RegisteredClaims
	Access []*token.ResourceActions `json:"access"`
}

// Valid checks if the issuer is harbor
func (c *Claims) Valid() error {
	if err := c.RegisteredClaims.Valid(); err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(c.Issuer), []byte(Issuer)) == 0 {
		return fmt.Errorf("invalid token issuer: %s", c.Issuer)
	}
	return nil
}
