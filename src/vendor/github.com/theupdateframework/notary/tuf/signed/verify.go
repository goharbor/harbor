package signed

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/sirupsen/logrus"
	"github.com/theupdateframework/notary/tuf/data"
	"github.com/theupdateframework/notary/tuf/utils"
)

// Various basic signing errors
var (
	ErrNoSignatures = errors.New("tuf: data has no signatures")
	ErrInvalid      = errors.New("tuf: signature verification failed")
	ErrWrongType    = errors.New("tuf: meta file has wrong type")
)

// IsExpired checks if the given time passed before the present time
func IsExpired(t time.Time) bool {
	return t.Before(time.Now())
}

// VerifyExpiry returns ErrExpired if the metadata is expired
func VerifyExpiry(s *data.SignedCommon, role data.RoleName) error {
	if IsExpired(s.Expires) {
		logrus.Errorf("Metadata for %s expired", role)
		return ErrExpired{Role: role, Expired: s.Expires.Format("Mon Jan 2 15:04:05 MST 2006")}
	}
	return nil
}

// VerifyVersion returns ErrLowVersion if the metadata version is lower than the min version
func VerifyVersion(s *data.SignedCommon, minVersion int) error {
	if s.Version < minVersion {
		return ErrLowVersion{Actual: s.Version, Current: minVersion}
	}
	return nil
}

// VerifySignatures checks the we have sufficient valid signatures for the given role
func VerifySignatures(s *data.Signed, roleData data.BaseRole) error {
	if len(s.Signatures) == 0 {
		return ErrNoSignatures
	}

	if roleData.Threshold < 1 {
		return ErrRoleThreshold{}
	}
	logrus.Debugf("%s role has key IDs: %s", roleData.Name, strings.Join(roleData.ListKeyIDs(), ","))

	// remarshal the signed part so we can verify the signature, since the signature has
	// to be of a canonically marshalled signed object
	var decoded map[string]interface{}
	if err := json.Unmarshal(*s.Signed, &decoded); err != nil {
		return err
	}
	msg, err := json.MarshalCanonical(decoded)
	if err != nil {
		return err
	}

	valid := make(map[string]struct{})
	for i := range s.Signatures {
		sig := &(s.Signatures[i])
		logrus.Debug("verifying signature for key ID: ", sig.KeyID)
		key, ok := roleData.Keys[sig.KeyID]
		if !ok {
			logrus.Debugf("continuing b/c keyid lookup was nil: %s\n", sig.KeyID)
			continue
		}
		// Check that the signature key ID actually matches the content ID of the key
		if key.ID() != sig.KeyID {
			return ErrInvalidKeyID{}
		}
		if err := VerifySignature(msg, sig, key); err != nil {
			logrus.Debugf("continuing b/c %s", err.Error())
			continue
		}
		valid[sig.KeyID] = struct{}{}
	}
	if len(valid) < roleData.Threshold {
		return ErrRoleThreshold{
			Msg: fmt.Sprintf("valid signatures did not meet threshold for %s", roleData.Name),
		}
	}

	return nil
}

// VerifySignature checks a single signature and public key against a payload
// If the signature is verified, the signature's is valid field will actually
// be mutated to be equal to the boolean true
func VerifySignature(msg []byte, sig *data.Signature, pk data.PublicKey) error {
	// method lookup is consistent due to Unmarshal JSON doing lower case for us.
	method := sig.Method
	verifier, ok := Verifiers[method]
	if !ok {
		return fmt.Errorf("signing method is not supported: %s", sig.Method)
	}

	if err := verifier.Verify(pk, sig.Signature, msg); err != nil {
		return fmt.Errorf("signature was invalid")
	}
	sig.IsValid = true
	return nil
}

// VerifyPublicKeyMatchesPrivateKey checks if the private key and the public keys forms valid key pairs.
// Supports both x509 certificate PublicKeys and non-certificate PublicKeys
func VerifyPublicKeyMatchesPrivateKey(privKey data.PrivateKey, pubKey data.PublicKey) error {
	pubKeyID, err := utils.CanonicalKeyID(pubKey)
	if err != nil {
		return fmt.Errorf("could not verify key pair: %v", err)
	}
	if privKey == nil || pubKeyID != privKey.ID() {
		return fmt.Errorf("private key is nil or does not match public key")
	}
	return nil
}
