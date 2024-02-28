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
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"

	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/log"
)

const (
	defaultIssuer       = "harbor-token-defaultIssuer"
	defaultSignedMethod = "RS256"
)

// Options ...
type Options struct {
	SignMethod jwt.SigningMethod
	PublicKey  []byte
	PrivateKey []byte
	Issuer     string
}

// GetKey ...
func (o *Options) GetKey() (interface{}, error) {
	var err error
	var privateKey *rsa.PrivateKey
	var publicKey *rsa.PublicKey

	switch o.SignMethod.(type) {
	case *jwt.SigningMethodRSA, *jwt.SigningMethodRSAPSS:
		if len(o.PrivateKey) > 0 {
			privateKey, err = jwt.ParseRSAPrivateKeyFromPEM(o.PrivateKey)
			if err != nil {
				return nil, err
			}
		}
		if len(o.PublicKey) > 0 {
			publicKey, err = jwt.ParseRSAPublicKeyFromPEM(o.PublicKey)
			if err != nil {
				return nil, err
			}
		}
		if privateKey == nil {
			if publicKey != nil {
				return publicKey, nil
			}
			return nil, fmt.Errorf("key is provided")
		}
		if publicKey != nil && publicKey.E != privateKey.E && publicKey.N.Cmp(privateKey.N) != 0 {
			return nil, fmt.Errorf("the public key and private key are not match")
		}
		return privateKey, nil
	default:
		return nil, fmt.Errorf(fmt.Sprintf("unsupported sign method, %s", o.SignMethod))
	}
}

// DefaultTokenOptions ...
func DefaultTokenOptions() *Options {
	opt, _ := NewOptions(defaultSignedMethod, defaultIssuer, config.TokenPrivateKeyPath())
	return opt
}

// NewOptions create Options based on input parms
func NewOptions(sm, iss, keyPath string) (*Options, error) {
	pk, err := os.ReadFile(keyPath)
	if err != nil {
		log.Errorf(fmt.Sprintf("failed to read private key %v", err))
		return nil, err
	}
	return &Options{
		PrivateKey: pk,
		SignMethod: jwt.GetSigningMethod(sm),
		Issuer:     iss,
	}, nil
}
