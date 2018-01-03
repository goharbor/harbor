package signed

import (
	"crypto"
	"crypto/rand"
	"encoding/pem"
	"io"
	"testing"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

const (
	testKeyPEM1 = "-----BEGIN PUBLIC KEY-----\nMIIBojANBgkqhkiG9w0BAQEFAAOCAY8AMIIBigKCAYEAnKuXZeefa2LmgxaL5NsM\nzKOHNe+x/nL6ik+lDBCTV6OdcwAhHQS+PONGhrChIUVR6Vth3hUCrreLzPO73Oo5\nVSCuRJ53UronENl6lsa5mFKP8StYLvIDITNvkoT3j52BJIjyNUK9UKY9As2TNqDf\nBEPIRp28ev/NViwGOEkBu2UAbwCIdnDXm8JQErCZA0Ydm7PKGgjLbFsFGrVzqXHK\n6pdzJXlhr9yap3UpgQ/iO9JtoEYB2EXsnSrPc9JRjR30bNHHtnVql3fvinXrAEwq\n3xmN4p+R4VGzfdQN+8Kl/IPjqWB535twhFYEG/B7Ze8IwbygBjK3co/KnOPqMUrM\nBI8ztvPiogz+MvXb8WvarZ6TMTh8ifZI96r7zzqyzjR1hJulEy3IsMGvz8XS2J0X\n7sXoaqszEtXdq5ef5zKVxkiyIQZcbPgmpHLq4MgfdryuVVc/RPASoRIXG4lKaTJj\n1ANMFPxDQpHudCLxwCzjCb+sVa20HBRPTnzo8LSZkI6jAgMBAAE=\n-----END PUBLIC KEY-----"
)

type FailingPrivateKeyErr struct {
}

func (err FailingPrivateKeyErr) Error() string {
	return "FailingPrivateKey.Sign failed"
}

// A data.PrivateKey which fails signing with a recognizable error.
type FailingPrivateKey struct {
	data.PrivateKey
}

func (fpk FailingPrivateKey) Sign(rand io.Reader, msg []byte, opts crypto.SignerOpts) (signature []byte, err error) {
	return nil, FailingPrivateKeyErr{}
}

// A CryptoService which does not contain any keys.
type FailingCryptoService struct {
}

func (mts *FailingCryptoService) Create(_ data.RoleName, _ data.GUN, _ string) (data.PublicKey, error) {
	return nil, nil
}

func (mts *FailingCryptoService) ListKeys(role data.RoleName) []string {
	return []string{}
}

func (mts *FailingCryptoService) AddKey(role data.RoleName, gun data.GUN, key data.PrivateKey) error {
	return nil
}

func (mts *FailingCryptoService) ListAllKeys() map[string]data.RoleName {
	return map[string]data.RoleName{}
}

func (mts *FailingCryptoService) GetKey(keyID string) data.PublicKey {
	return nil
}

func (mts *FailingCryptoService) GetPrivateKey(keyID string) (data.PrivateKey, data.RoleName, error) {
	return nil, "", trustmanager.ErrKeyNotFound{KeyID: keyID}
}

func (mts *FailingCryptoService) RemoveKey(keyID string) error {
	return nil
}

// A CryptoService which only only allows using one key
type MockCryptoService struct {
	testKey data.PrivateKey
}

func (mts *MockCryptoService) Create(_ data.RoleName, _ data.GUN, _ string) (data.PublicKey, error) {
	return mts.testKey, nil
}

func (mts *MockCryptoService) AddKey(role data.RoleName, gun data.GUN, key data.PrivateKey) error {
	return nil
}

func (mts *MockCryptoService) GetKey(keyID string) data.PublicKey {
	if keyID == mts.testKey.ID() {
		return data.PublicKeyFromPrivate(mts.testKey)
	}
	return nil
}

func (mts *MockCryptoService) ListKeys(role data.RoleName) []string {
	return []string{mts.testKey.ID()}
}

func (mts *MockCryptoService) ListAllKeys() map[string]data.RoleName {
	return map[string]data.RoleName{
		mts.testKey.ID(): data.CanonicalRootRole,
		mts.testKey.ID(): data.CanonicalTargetsRole,
		mts.testKey.ID(): data.CanonicalSnapshotRole,
		mts.testKey.ID(): data.CanonicalTimestampRole,
	}
}

func (mts *MockCryptoService) GetPrivateKey(keyID string) (data.PrivateKey, data.RoleName, error) {
	if keyID == mts.testKey.ID() {
		return mts.testKey, "testRole", nil
	}
	return nil, "", trustmanager.ErrKeyNotFound{KeyID: keyID}
}

func (mts *MockCryptoService) RemoveKey(keyID string) error {
	return nil
}

// Test signing and ensure the expected signature is added
func TestBasicSign(t *testing.T) {
	cs := NewEd25519()
	key, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)
	testData := data.Signed{
		Signed: &json.RawMessage{},
	}

	err = Sign(cs, &testData, []data.PublicKey{key}, 1, nil)
	require.NoError(t, err)

	if len(testData.Signatures) != 1 {
		t.Fatalf("Incorrect number of signatures: %d", len(testData.Signatures))
	}

	if testData.Signatures[0].KeyID != key.ID() {
		t.Fatalf("Wrong signature ID returned: %s", testData.Signatures[0].KeyID)
	}
}

// Signing with the same key multiple times should not produce multiple sigs
// with the same key ID
func TestReSign(t *testing.T) {
	cs := NewEd25519()
	key, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)
	testData := data.Signed{
		Signed: &json.RawMessage{},
	}

	Sign(cs, &testData, []data.PublicKey{key}, 1, nil)
	Sign(cs, &testData, []data.PublicKey{key}, 1, nil)

	if len(testData.Signatures) != 1 {
		t.Fatalf("Incorrect number of signatures: %d", len(testData.Signatures))
	}

	if testData.Signatures[0].KeyID != key.ID() {
		t.Fatalf("Wrong signature ID returned: %s", testData.Signatures[0].KeyID)
	}

}

// Should not remove signatures for valid keys that were not resigned with
func TestMultiSign(t *testing.T) {
	cs := NewEd25519()
	testData := data.Signed{
		Signed: &json.RawMessage{},
	}

	key1, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	require.NoError(t, Sign(cs, &testData, []data.PublicKey{key1}, 1, nil))

	// reinitializing cs means it won't know about key1. We want
	// to attempt to sign passing both key1 and key2, while expecting
	// that the signature for key1 is left intact and the signature
	// for key2 is added
	cs = NewEd25519()
	key2, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	err = Sign(
		cs,
		&testData,
		[]data.PublicKey{key2},
		1,
		[]data.PublicKey{key1},
	)
	require.NoError(t, err)

	require.Len(t, testData.Signatures, 2)
	require.Equal(t, key2.ID(), testData.Signatures[0].KeyID)
	require.Equal(t, key1.ID(), testData.Signatures[1].KeyID)
}

func TestSignReturnsNoSigs(t *testing.T) {
	failingCryptoService := &FailingCryptoService{}
	testData := data.Signed{
		Signed: &json.RawMessage{},
	}

	testKey, _ := pem.Decode([]byte(testKeyPEM1))
	key := data.NewPublicKey(data.RSAKey, testKey.Bytes)
	err := Sign(failingCryptoService, &testData, []data.PublicKey{key}, 1, nil)

	require.Error(t, err)
	require.IsType(t, ErrInsufficientSignatures{}, err)
	if len(testData.Signatures) != 0 {
		t.Fatalf("Incorrect number of signatures, expected 0: %d", len(testData.Signatures))
	}
}

func TestSignWithX509(t *testing.T) {
	// generate a key becase we need a cert
	privKey, err := utils.GenerateRSAKey(rand.Reader, 1024)
	require.NoError(t, err)

	// make a RSA x509 key
	cert, err := cryptoservice.GenerateCertificate(privKey, "test", time.Now(), time.Now().AddDate(10, 0, 0))
	require.NoError(t, err)

	tufRSAx509Key := utils.CertToKey(cert)
	require.NoError(t, err)

	// test signing against a service that only recognizes a RSAKey (not
	// RSAx509 key)
	mockCryptoService := &MockCryptoService{privKey}
	testData := data.Signed{
		Signed: &json.RawMessage{},
	}

	err = Sign(mockCryptoService, &testData, []data.PublicKey{tufRSAx509Key}, 1, nil)
	require.NoError(t, err)

	require.Len(t, testData.Signatures, 1)
	require.Equal(t, tufRSAx509Key.ID(), testData.Signatures[0].KeyID)
}

func TestSignRemovesValidSigByInvalidKey(t *testing.T) {
	cs := NewEd25519()
	testData := data.Signed{
		Signed: &json.RawMessage{},
	}

	key1, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	key2, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	require.NoError(t, Sign(cs, &testData, []data.PublicKey{key1, key2}, 1, nil))
	require.Len(t, testData.Signatures, 2)
	var signatureKeys []string
	for _, sig := range testData.Signatures {
		signatureKeys = append(signatureKeys, sig.KeyID)
	}
	require.Contains(t, signatureKeys, key1.ID())
	require.Contains(t, signatureKeys, key2.ID())

	key3, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	// should remove key1 sig even though it's valid. It no longer appears
	// in the list of signing keys or valid signing keys for the role
	require.NoError(t, Sign(cs, &testData, []data.PublicKey{key3}, 1, []data.PublicKey{key2}))
	require.Len(t, testData.Signatures, 2)
	signatureKeys = nil
	for _, sig := range testData.Signatures {
		signatureKeys = append(signatureKeys, sig.KeyID)
	}
	require.Contains(t, signatureKeys, key2.ID())
	require.Contains(t, signatureKeys, key3.ID())
}

func TestSignRemovesInvalidSig(t *testing.T) {
	cs := NewEd25519()
	testData := data.Signed{
		Signed: &json.RawMessage{},
	}

	key1, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	require.NoError(t, Sign(cs, &testData, []data.PublicKey{key1}, 1, nil))
	require.Len(t, testData.Signatures, 1)
	require.Equal(t, key1.ID(), testData.Signatures[0].KeyID)

	// we need cs to "forget" key1 so we can't sign with it
	cs = NewEd25519()
	key2, err := cs.Create(data.CanonicalRootRole, "", data.ED25519Key)
	require.NoError(t, err)

	// modify test data to invalidate key1 sig
	raw := json.RawMessage([]byte{0xff})
	testData.Signed = &raw
	// should remove key1 sig because it's out of date
	Sign(cs, &testData, []data.PublicKey{key1, key2}, 1, nil)

	require.Len(t, testData.Signatures, 1)
	require.Equal(t, key2.ID(), testData.Signatures[0].KeyID)
}

func TestSignMinSignatures(t *testing.T) {
	csA := NewEd25519()
	keyA1, err := csA.Create("keyA", "", data.ED25519Key)
	require.NoError(t, err)
	keyA2, err := csA.Create("keyA", "", data.ED25519Key)
	require.NoError(t, err)
	// csB is only used to create public keys which are unavailable from csA.
	csB := NewEd25519()
	keyB, err := csB.Create("keyB", "", data.ED25519Key)
	require.NoError(t, err)

	allKeys := []data.PublicKey{keyA1, keyA2, keyB}

	// 2 available keys, threshold 1: 2 signatures created nevertheless
	testData := data.Signed{Signed: &json.RawMessage{}}
	err = Sign(csA, &testData, allKeys, 1, nil)
	require.NoError(t, err)
	require.Len(t, testData.Signatures, 2)

	// 2 available keys, threshold 2
	testData = data.Signed{Signed: &json.RawMessage{}}
	err = Sign(csA, &testData, allKeys, 2, nil)
	require.NoError(t, err)
	require.Len(t, testData.Signatures, 2)

	// 2 available keys, threshold 3
	testData = data.Signed{Signed: &json.RawMessage{}}
	err = Sign(csA, &testData, allKeys, 3, nil)
	require.Error(t, err)
	if err2, ok := err.(ErrInsufficientSignatures); ok {
		require.Equal(t, err2.FoundKeys, 2)
		require.Equal(t, err2.NeededKeys, 3)
	} else {
		// We know this will fail if !ok
		require.IsType(t, ErrInsufficientSignatures{}, err)
	}
}

func TestSignFailingKeys(t *testing.T) {
	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	cs := &MockCryptoService{FailingPrivateKey{privKey}}

	testData := data.Signed{Signed: &json.RawMessage{}}
	err = Sign(cs, &testData, []data.PublicKey{privKey}, 1, nil)
	require.Error(t, err)
	require.IsType(t, FailingPrivateKeyErr{}, err)
}

// make sure we produce readable error messages
func TestErrInsufficientSignaturesMessaging(t *testing.T) {
	require.Contains(t,
		ErrInsufficientSignatures{NeededKeys: 2, MissingKeyIDs: []string{"ID1", "ID2"}}.Error(),
		"need 2 keys from 2 possible keys (ID1, ID2)")
	require.Contains(t,
		ErrInsufficientSignatures{FoundKeys: 1, NeededKeys: 2, MissingKeyIDs: []string{"ID1", "ID2"}}.Error(),
		"found 1 of 2 needed keys - 2 other possible keys (ID1, ID2)")
	require.Contains(t,
		ErrInsufficientSignatures{FoundKeys: 1, NeededKeys: 2, MissingKeyIDs: []string{}}.Error(),
		"found 1 of 2 needed keys - 0 other possible keys")
}
