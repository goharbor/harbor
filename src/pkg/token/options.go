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
			return nil, fmt.Errorf("no key provided")
		}
		if publicKey != nil && (publicKey.E != privateKey.E || publicKey.N.Cmp(privateKey.N) != 0) {
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
		if publicKey != nil && !publicKey.Equal(&privateKey.PublicKey) {
			return nil, fmt.Errorf("the public key and private key are not match")
		}
		return privateKey, nil
	default:
		return nil, fmt.Errorf("unsupported sign method, %v", o.SignMethod)
	}
}

// DefaultTokenOptions ...
func DefaultTokenOptions() *Options {
	opt, _ := NewOptions(defaultSignedMethod, defaultIssuer, config.TokenPrivateKeyPath())
	return opt
}

// NewOptions creates Options based on the input parameters.
// The first parameter is deprecated and ignored; the signing method is
// automatically determined from the key type.
func NewOptions(_, iss, keyPath string) (*Options, error) {
	pkBytes, err := os.ReadFile(keyPath)
	if err != nil {
		log.Errorf("failed to read private key %v", err)
		return nil, err
	}
	var (
		block      *pem.Block
		rest       = pkBytes
		privateKey any
	)
	for {
		block, rest = pem.Decode(rest)
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM")
		}
		switch block.Type {
		case "RSA PRIVATE KEY":
			privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		case "PRIVATE KEY":
			privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		case "EC PRIVATE KEY":
			privateKey, err = x509.ParseECPrivateKey(block.Bytes)
		default:
			// Skip unsupported PEM block types (e.g., EC PARAMETERS) and
			// continue scanning remaining blocks, if any.
			if len(rest) > 0 {
				continue
			}
			return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
		}
		// Reached a supported private key block (parsing result checked below).
		break
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}
	var signMethod jwt.SigningMethod
	switch k := privateKey.(type) {
	case *rsa.PrivateKey:
		signMethod = jwt.SigningMethodRS256
	case *ecdsa.PrivateKey:
		switch k.Curve.Params().Name {
		case "P-256":
			signMethod = jwt.SigningMethodES256
		case "P-384":
			signMethod = jwt.SigningMethodES384
		case "P-521":
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
