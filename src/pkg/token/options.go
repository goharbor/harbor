package token

import (
	"crypto/rsa"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"io/ioutil"
	"time"
)

const (
	defaultTTL          = 60 * time.Minute
	defaultIssuer       = "harbor-token-defaultIssuer"
	defaultSignedMethod = "RS256"
)

// Options ...
type Options struct {
	SignMethod jwt.SigningMethod
	PublicKey  []byte
	PrivateKey []byte
	TTL        time.Duration
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
	privateKeyFile := config.TokenPrivateKeyPath()
	privateKey, err := ioutil.ReadFile(privateKeyFile)
	if err != nil {
		log.Errorf(fmt.Sprintf("failed to read private key %v", err))
		return nil
	}
	opt := &Options{
		SignMethod: jwt.GetSigningMethod(defaultSignedMethod),
		PrivateKey: privateKey,
		Issuer:     defaultIssuer,
		TTL:        defaultTTL,
	}
	return opt
}
