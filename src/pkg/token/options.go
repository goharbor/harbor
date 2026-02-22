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
	"crypto/elliptic"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
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
func (o *Options) GetKey() (any, error) {
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
	case *jwt.SigningMethodECDSA:
		var privateKey *ecdsa.PrivateKey
		var publicKey *ecdsa.PublicKey
		var err error

		if len(o.PrivateKey) > 0 {
			privateKey, err = jwt.ParseECPrivateKeyFromPEM(o.PrivateKey)
			if err != nil {
				return nil, err
			}
		}
		if len(o.PublicKey) > 0 {
			publicKey, err = jwt.ParseECPublicKeyFromPEM(o.PublicKey)
			if err != nil {
				return nil, err
			}
		}
		if privateKey == nil {
			if publicKey != nil {
				return publicKey, nil
			}
			return nil, fmt.Errorf("no key provided")
		}
		if publicKey != nil && (publicKey.X.Cmp(privateKey.X) != 0 || publicKey.Y.Cmp(privateKey.Y) != 0 || publicKey.Curve != privateKey.Curve) {
			return nil, fmt.Errorf("the public key and private key are not match")
		}
		return privateKey, nil
	default:
		return nil, fmt.Errorf("unsupported sign method, %s", o.SignMethod)
	}
}

// DefaultTokenOptions ...
func DefaultTokenOptions() *Options {
	opt, _ := NewOptions(defaultSignedMethod, defaultIssuer, config.TokenPrivateKeyPath())
	return opt
}

// NewOptions create Options based on input parms
func NewOptions(_, iss, keyPath string) (*Options, error) {
	pkBytes, err := os.ReadFile(keyPath)
	if err != nil {
		log.Errorf("failed to read private key %v", err)
		return nil, err
	}
	block, _ := pem.Decode(pkBytes)
	if block == nil {
		return nil, fmt.Errorf("failed to decode PEM")
	}
	var privateKey any
	switch block.Type {
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
	case "EC PRIVATE KEY":
		privateKey, err = x509.ParseECPrivateKey(block.Bytes)
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}
	var signMethod jwt.SigningMethod
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		signMethod = jwt.SigningMethodRS256
	case *ecdsa.PrivateKey:
		switch k.Curve {
		case elliptic.P256():
			signMethod = jwt.SigningMethodES256
		case elliptic.P384():
			signMethod = jwt.SigningMethodES384
		case elliptic.P521():
			signMethod = jwt.SigningMethodES512
		default:
			return nil, fmt.Errorf("unsupported ECDSA curve: %s", k.Curve.Params().Name)
		}
	default:
		return nil, fmt.Errorf("unsupported private key type: %T", privateKey)
	}
	return &Options{
		PrivateKey: pkBytes,
		SignMethod: signMethod,
		Issuer:     iss,
	}, nil
}
