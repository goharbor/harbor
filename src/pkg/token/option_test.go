package token

import (
	"testing"

	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewOptions(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	if defaultOpt == nil {
		assert.NotNil(t, defaultOpt)
		return
	}
	assert.Equal(t, defaultOpt.SignMethod, jwt.GetSigningMethod("RS256"))
	assert.Equal(t, defaultOpt.Issuer, "harbor-token-defaultIssuer")
}

func TestGetKey(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	if defaultOpt == nil {
		assert.NotNil(t, defaultOpt)
		return
	}
	key, err := defaultOpt.GetKey()
	assert.Nil(t, err)
	assert.NotNil(t, key)
}
