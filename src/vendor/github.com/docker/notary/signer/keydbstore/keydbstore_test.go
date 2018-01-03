package keydbstore

import (
	"crypto/rand"
	"errors"
	"fmt"
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

func constRetriever(string, string, bool, int) (string, bool, error) {
	return "constantPass", false, nil
}

var validAliases = []string{"validAlias1", "validAlias2"}
var validAliasesAndPasswds = map[string]string{
	"validAlias1": "passphrase_1",
	"validAlias2": "passphrase_2",
}

func multiAliasRetriever(_, alias string, _ bool, _ int) (string, bool, error) {
	if passwd, ok := validAliasesAndPasswds[alias]; ok {
		return passwd, false, nil
	}
	return "", false, errors.New("password alias not found")
}

type keyRotator interface {
	signed.CryptoService
	RotateKeyPassphrase(keyID, newPassphraseAlias string) error
}

// A key can only be added to the DB once.  Returns a list of expected keys, and which keys are expected to exist.
func testKeyCanOnlyBeAddedOnce(t *testing.T, dbStore signed.CryptoService) []data.PrivateKey {
	expectedKeys := make([]data.PrivateKey, 2)
	for i := 0; i < len(expectedKeys); i++ {
		testKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err)
		expectedKeys[i] = testKey
	}

	// Test writing new key in database alone, not cache
	err := dbStore.AddKey(data.CanonicalTimestampRole, "gun", expectedKeys[0])
	require.NoError(t, err)
	requireGetKeySuccess(t, dbStore, data.CanonicalTimestampRole.String(), expectedKeys[0])

	// Test writing the same key in the database. Should fail.
	err = dbStore.AddKey(data.CanonicalTimestampRole, "gun", expectedKeys[0])
	require.Error(t, err, "failed to add private key to database:")

	// Test writing new key succeeds
	err = dbStore.AddKey(data.CanonicalTimestampRole, "gun", expectedKeys[1])
	require.NoError(t, err)

	return expectedKeys
}

// a key can be deleted - returns a list of expected keys
func testCreateDelete(t *testing.T, dbStore signed.CryptoService) []data.PrivateKey {
	testKeys := make([]data.PrivateKey, 2)
	for i := 0; i < len(testKeys); i++ {
		testKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err)
		testKeys[i] = testKey

		// Add them to the DB
		err = dbStore.AddKey(data.CanonicalTimestampRole, "gun", testKey)
		require.NoError(t, err)
		requireGetKeySuccess(t, dbStore, data.CanonicalTimestampRole.String(), testKey)
	}

	// Deleting the key should succeed and only remove the key that was deleted
	require.NoError(t, dbStore.RemoveKey(testKeys[0].ID()))
	requireGetKeyFailure(t, dbStore, testKeys[0].ID())
	requireGetKeySuccess(t, dbStore, data.CanonicalTimestampRole.String(), testKeys[1])

	// Deleting the key again should succeed even though it's not in the DB
	require.NoError(t, dbStore.RemoveKey(testKeys[0].ID()))
	requireGetKeyFailure(t, dbStore, testKeys[0].ID())

	return testKeys[1:]
}

// key rotation is successful provided the other alias is valid.
// Returns the key that was rotated and one that was not rotated
func testKeyRotation(t *testing.T, dbStore keyRotator, newValidAlias string) (data.PrivateKey, data.PrivateKey) {
	testKeys := make([]data.PrivateKey, 2)
	for i := 0; i < len(testKeys); i++ {
		testKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err)
		testKeys[i] = testKey

		// Add them to the DB
		err = dbStore.AddKey(data.CanonicalTimestampRole, "gun", testKey)
		require.NoError(t, err)
	}

	// Try rotating the key to a valid alias
	err := dbStore.RotateKeyPassphrase(testKeys[0].ID(), newValidAlias)
	require.NoError(t, err)

	// Try rotating the key to an invalid alias
	err = dbStore.RotateKeyPassphrase(testKeys[0].ID(), "invalidAlias")
	require.Error(t, err, "there should be no password for invalidAlias so rotation should fail")

	return testKeys[0], testKeys[1]
}

type badReader struct{}

func (b badReader) Read([]byte) (n int, err error) {
	return 0, fmt.Errorf("Nope, not going to read")
}

// Signing with a key marks it as active if the signing is successful.  Marking as active is successful no matter what,
// but should only activate a key that exists in the DB.
// Returns the key that was used and one that was not
func testSigningWithKeyMarksAsActive(t *testing.T, dbStore signed.CryptoService) (data.PrivateKey, data.PrivateKey) {
	testKeys := make([]data.PrivateKey, 3)
	for i := 0; i < len(testKeys); i++ {
		testKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err)

		// Add them to the DB
		err = dbStore.AddKey(data.CanonicalTimestampRole, "gun", testKey)
		require.NoError(t, err)
		requireGetKeySuccess(t, dbStore, data.CanonicalTimestampRole.String(), testKey)

		// store the gotten key, because that key is special
		gottenKey, _, err := dbStore.GetPrivateKey(testKey.ID())
		require.NoError(t, err)
		testKeys[i] = gottenKey
	}

	// sign successfully with the first key - this key will become active
	msg := []byte("successful")
	sig, err := testKeys[0].Sign(rand.Reader, msg, nil)
	require.NoError(t, err)
	require.NoError(t, signed.Verifiers[data.ECDSASignature].Verify(
		data.PublicKeyFromPrivate(testKeys[0]), sig, msg))

	// sign unsuccessfully with the second key - this key should remain inactive
	sig, err = testKeys[1].Sign(badReader{}, []byte("unsuccessful"), nil)
	require.Error(t, err)
	require.Equal(t, "Nope, not going to read", err.Error())
	require.Nil(t, sig)

	// delete the third key from the DB - sign should still succeed, even though
	// this key cannot be marked as active anymore due to it not existing
	// (this probably won't return an error)
	require.NoError(t, dbStore.RemoveKey(testKeys[2].ID()))
	requireGetKeyFailure(t, dbStore, testKeys[2].ID())
	msg = []byte("successful, not active")
	sig, err = testKeys[2].Sign(rand.Reader, msg, nil)
	require.NoError(t, err)
	require.NoError(t, signed.Verifiers[data.ECDSASignature].Verify(
		data.PublicKeyFromPrivate(testKeys[2]), sig, msg))

	return testKeys[0], testKeys[1] // testKeys[2] should no longer exist in the DB
}

func testCreateKey(t *testing.T, dbStore signed.CryptoService) (data.PrivateKey, data.PrivateKey, data.PrivateKey) {
	// Create a test key, and check that it is successfully added to the database
	role := data.CanonicalSnapshotRole
	var gun data.GUN = "gun"

	// First create an ECDSA key
	createdECDSAKey, err := dbStore.Create(role, gun, data.ECDSAKey)
	require.NoError(t, err)
	require.NotNil(t, createdECDSAKey)
	require.Equal(t, data.ECDSAKey, createdECDSAKey.Algorithm())

	// Retrieve the key from the database by ID, and check that it is correct
	requireGetPubKeySuccess(t, dbStore, role.String(), createdECDSAKey)

	// Calling Create with the same parameters will return the same key because it is inactive
	createdSameECDSAKey, err := dbStore.Create(role, gun, data.ECDSAKey)
	require.NoError(t, err)
	require.Equal(t, createdECDSAKey.Algorithm(), createdSameECDSAKey.Algorithm())
	require.Equal(t, createdECDSAKey.Public(), createdSameECDSAKey.Public())
	require.Equal(t, createdECDSAKey.ID(), createdSameECDSAKey.ID())

	// Calling Create with the same role and gun but a different algorithm will create a new key
	createdED25519Key, err := dbStore.Create(role, gun, data.ED25519Key)
	require.NoError(t, err)
	require.NotEqual(t, createdECDSAKey.Algorithm(), createdED25519Key.Algorithm())
	require.NotEqual(t, createdECDSAKey.Public(), createdED25519Key.Public())
	require.NotEqual(t, createdECDSAKey.ID(), createdED25519Key.ID())

	// Retrieve the key from the database by ID, and check that it is correct
	requireGetPubKeySuccess(t, dbStore, role.String(), createdED25519Key)

	// Sign with the ED25519 key from the DB to mark it as active
	activeED25519Key, _, err := dbStore.GetPrivateKey(createdED25519Key.ID())
	require.NoError(t, err)
	_, err = activeED25519Key.Sign(rand.Reader, []byte("msg"), nil)
	require.NoError(t, err)

	// Calling Create for the same role, gun and ED25519 algorithm will now create a new key
	createdNewED25519Key, err := dbStore.Create(role, gun, data.ED25519Key)
	require.NoError(t, err)
	require.Equal(t, activeED25519Key.Algorithm(), createdNewED25519Key.Algorithm())
	require.NotEqual(t, activeED25519Key.Public(), createdNewED25519Key.Public())
	require.NotEqual(t, activeED25519Key.ID(), createdNewED25519Key.ID())

	// Get the inactive ED25519 key from the database explicitly to return
	inactiveED25519Key, _, err := dbStore.GetPrivateKey(createdNewED25519Key.ID())
	require.NoError(t, err)

	// Get the inactive ECDSA key from the database explicitly to return
	inactiveECDSAKey, _, err := dbStore.GetPrivateKey(createdSameECDSAKey.ID())
	require.NoError(t, err)

	// Calling Create with an invalid algorithm gives an error
	_, err = dbStore.Create(role, gun, "invalid")
	require.Error(t, err)

	return activeED25519Key, inactiveED25519Key, inactiveECDSAKey
}

func testUnimplementedInterfaceMethods(t *testing.T, dbStore signed.CryptoService) {
	// add one key to the db
	testKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)
	err = dbStore.AddKey(data.CanonicalTimestampRole, "gun", testKey)
	require.NoError(t, err)
	requireGetKeySuccess(t, dbStore, data.CanonicalTimestampRole.String(), testKey)

	// these are unimplemented/unused, and return nil
	require.Nil(t, dbStore.ListAllKeys())
	require.Nil(t, dbStore.ListKeys(data.CanonicalTimestampRole))
}
