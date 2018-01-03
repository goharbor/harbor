package keydbstore

import (
	"crypto"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/Sirupsen/logrus"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
)

type activatingPrivateKey struct {
	data.PrivateKey
	activationFunc func(keyID string) error
}

func (a activatingPrivateKey) Sign(rand io.Reader, digest []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	keyID := a.PrivateKey.ID()
	sig, err := a.PrivateKey.Sign(rand, digest, opts)
	if err == nil {
		if activationErr := a.activationFunc(keyID); activationErr != nil {
			logrus.Errorf("Key %s was just used to sign hash %s, error when trying to mark key as active: %s",
				keyID, digest, activationErr.Error())
		}
	}
	return sig, err
}

// helper function to generate private keys for the signer databases - does not implement RSA since that is not
// supported by the signer
func generatePrivateKey(algorithm string) (data.PrivateKey, error) {
	var privKey data.PrivateKey
	var err error
	switch algorithm {
	case data.ECDSAKey:
		privKey, err = utils.GenerateECDSAKey(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate EC key: %v", err)
		}
	case data.ED25519Key:
		privKey, err = utils.GenerateED25519Key(rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("failed to generate ED25519 key: %v", err)
		}
	default:
		return nil, fmt.Errorf("private key type not supported for key generation: %s", algorithm)
	}
	return privKey, nil
}
