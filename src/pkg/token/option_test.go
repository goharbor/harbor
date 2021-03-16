package token

import (
	"testing"

	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	assert.NotNil(t, defaultOpt)
	assert.Equal(t, defaultOpt.SignMethod, jwt.GetSigningMethod("RS256"))
	assert.Equal(t, defaultOpt.Issuer, "harbor-token-defaultIssuer")
}

func TestGetKey(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	key, err := defaultOpt.GetKey()
	assert.Nil(t, err)
	assert.NotNil(t, key)
}
