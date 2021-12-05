package token

import (
	"crypto/rsa"
	"fmt"
	"github.com/goharbor/harbor/src/lib/config"
	"io/ioutil"

	"github.com/goharbor/harbor/src/lib/log"
	"github.com/golang-jwt/jwt/v4"
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
	pk, err := ioutil.ReadFile(keyPath)
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
