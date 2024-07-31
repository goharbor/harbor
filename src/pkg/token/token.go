// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package token

import (
	"crypto/ecdsa"
	"crypto/rsa"
	"errors"
	"fmt"

	"github.com/golang-jwt/jwt/v5"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/lib/log"
)

// Token is a jwt token for harbor robot account,
type Token struct {
	jwt.Token
	Opt   *Options
	Claim jwt.Claims
}

// New ...
func New(opt *Options, claims jwt.Claims) (*Token, error) {
	var v = jwt.NewValidator(jwt.WithLeeway(common.JwtLeeway))
	if err := v.Validate(claims); err != nil {
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
	var parser = jwt.NewParser(jwt.WithLeeway(common.JwtLeeway), jwt.WithValidMethods([]string{opt.SignMethod.Alg()}))
	token, err := parser.ParseWithClaims(rawToken, claims, func(token *jwt.Token) (interface{}, error) {
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
