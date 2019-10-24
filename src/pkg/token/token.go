package token

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// Token is a jwt token for harbor robot account,
type Token struct {
	jwt.Token
	Opt   *Options
	Claim jwt.Claims
}

// New ...
func New(opt *Options, claims jwt.Claims) (*Token, error) {
	err := claims.Valid()
	if err != nil {
		return nil, err
	}
	return &Token{
		Token: *jwt.NewWithClaims(opt.SignMethod, claims),
		Opt:   opt,
		Claim: claims,
	}, nil
}

// Raw get the Raw string of token
func (tk *Token) Raw() (string, error) {
	key, err := tk.Opt.GetKey()
	if err != nil {
		return "", nil
	}
	raw, err := tk.Token.SignedString(key)
	if err != nil {
		log.Debugf(fmt.Sprintf("failed to issue token %v", err))
		return "", err
	}
	return raw, err
}

// Parse ...
func Parse(opt *Options, rawToken string, claims jwt.Claims) (*Token, error) {
	key, err := opt.GetKey()
	if err != nil {
		return nil, err
	}
	token, err := jwt.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != opt.SignMethod.Alg() {
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
	return &Token{
		Token: *token,
	}, nil
}
