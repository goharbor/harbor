package token

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"time"
)

// HToken htoken is a jwt token for harbor robot account,
// which contains the robot ID, project ID and the access permission for the project.
// It used for authn/authz for robot account in Harbor.
type HToken struct {
	jwt.Token
}

// New ...
func New(tokenID, projectID, expiresAt int64, access []*rbac.Policy) (*HToken, error) {
	rClaims := &RobotClaims{
		TokenID:   tokenID,
		ProjectID: projectID,
		Access:    access,
		StandardClaims: jwt.StandardClaims{
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: expiresAt,
			Issuer:    DefaultOptions().Issuer,
		},
	}
	err := rClaims.Valid()
	if err != nil {
		return nil, err
	}
	return &HToken{
		Token: *jwt.NewWithClaims(DefaultOptions().SignMethod, rClaims),
	}, nil
}

// Raw get the Raw string of token
func (htk *HToken) Raw() (string, error) {
	key, err := DefaultOptions().GetKey()
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
	key, err := DefaultOptions().GetKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != DefaultOptions().SignMethod.Alg() {
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
		return nil, errors.New("invalid jwt token")
	}
	return &HToken{
		Token: *token,
	}, nil
}
