package token

import (
	"github.com/dgrijalva/jwt-go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewOptions(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	assert.NotNil(t, defaultOpt)
	assert.Equal(t, defaultOpt.SignMethod, jwt.GetSigningMethod("RS256"))
	assert.Equal(t, defaultOpt.Issuer, "harbor-token-issuer")
	assert.Equal(t, defaultOpt.TTL, 60*time.Minute)
}

func TestGetKey(t *testing.T) {
	defaultOpt := DefaultTokenOptions()
	key, err := defaultOpt.GetKey()
	assert.Nil(t, err)
	assert.NotNil(t, key)
}
