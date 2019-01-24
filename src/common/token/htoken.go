package token

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/utils/log"
	"time"
)

// HToken ...
type HToken struct {
	jwt.Token
}

// NewWithClaims ...
func NewWithClaims(claims *RobotClaims) *HToken {
	rClaims := &RobotClaims{
		TokenID:   claims.TokenID,
		ProjectID: claims.ProjectID,
		Policy:    claims.Policy,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(DefaultOptions.TTL).Unix(),
			Issuer:    DefaultOptions.Issuer,
		},
	}
	return &HToken{
		Token: *jwt.NewWithClaims(DefaultOptions.SignMethod, rClaims),
	}
}

// SignedString get the SignedString.
func (htk *HToken) SignedString() (string, error) {
	key, err := DefaultOptions.GetKey()
	if err != nil {
		return "", nil
	}
	raw, err := htk.Token.SignedString(key)
	if err != nil {
		log.Debugf(fmt.Sprintf("failed to issue token %v", err))
		return "", err
	}
	return raw, err
}

// ParseWithClaims ...
func ParseWithClaims(rawToken string, claims jwt.Claims) (*HToken, error) {
	key, err := DefaultOptions.GetKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != DefaultOptions.SignMethod.Alg() {
			return nil, errors.New("invalid signing method")
		}
		switch k := key.(type) {
		case *rsa.PrivateKey:
			return &k.PublicKey, nil
		case *ecdsa.PrivateKey:
			return &k.PublicKey, nil
		default:
			return key, nil
		}
	})
	if err != nil {
		log.Errorf(fmt.Sprintf("parse token error, %v", err))
		return nil, err
	}
	if !token.Valid {
		log.Errorf(fmt.Sprintf("invalid jwt token, %v", token))
		return nil, err
	}
	return &HToken{
		Token: *token,
	}, nil
}
