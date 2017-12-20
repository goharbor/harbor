package interfaces

import (
	"crypto/rand"
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

// These are tests that can be used to test a cryptoservice

// EmptyCryptoServiceInterfaceBehaviorTests tests expected behavior for
// an empty signed.CryptoService:
// 1.  Getting the public key of a key that doesn't exist should fail
// 2.  Listing an empty cryptoservice returns no keys
// 3.  Removing a non-existent key succeeds (no-op)
func EmptyCryptoServiceInterfaceBehaviorTests(t *testing.T, empty signed.CryptoService) {
	for _, role := range append(data.BaseRoles, "targets/delegation", "invalid") {
		keys := empty.ListKeys(role)
		require.Len(t, keys, 0)
	}
	keys := empty.ListAllKeys()
	require.Len(t, keys, 0)

	require.NoError(t, empty.RemoveKey("nonexistent"))

	require.Nil(t, empty.GetKey("nonexistent"))

	k, role, err := empty.GetPrivateKey("nonexistent")
	require.Error(t, err)
	require.Nil(t, k)
	require.EqualValues(t, "", role)
}

// CreateGetKeyCryptoServiceInterfaceBehaviorTests tests expected behavior for
// creating keys in a signed.CryptoService and other read operations on the
// crypto service after keys are present
// 1.  Creating a key succeeds and returns a non-nil public key
// 2.  Getting the key should return the same key, without error
// 3.  Removing the key succeeds
func CreateGetKeyCryptoServiceInterfaceBehaviorTests(t *testing.T, cs signed.CryptoService, algo string) {

	expectedRolesToKeys := make(map[string]string)
	for i := 0; i < 2; i++ {
		role := data.BaseRoles[i+1]
		createdPubKey, err := cs.Create(role, "docker.io/notary", algo)
		require.NoError(t, err)
		require.NotNil(t, createdPubKey)
		expectedRolesToKeys[role.String()] = createdPubKey.ID()
	}

	testGetKey(t, cs, expectedRolesToKeys, algo)
}

// CreateListKeyCryptoServiceInterfaceBehaviorTests tests expected behavior for
// creating keys in a signed.CryptoService and listing keys after keys are
// present
// 1.  Creating a key succeeds and returns a non-nil public key
// 2.  Listing returns the correct number of keys and right roles
func CreateListKeyCryptoServiceInterfaceBehaviorTests(t *testing.T, cs signed.CryptoService, algo string) {
	expectedRolesToKeys := make(map[string]string)
	for i := 0; i < 2; i++ {
		role := data.BaseRoles[i+1]
		createdPubKey, err := cs.Create(role, "docker.io/notary", algo)
		require.NoError(t, err)
		require.NotNil(t, createdPubKey)
		expectedRolesToKeys[role.String()] = createdPubKey.ID()
	}

	testListKeys(t, cs, expectedRolesToKeys)
}

// AddGetKeyCryptoServiceInterfaceBehaviorTests tests expected behavior for
// adding keys in a signed.CryptoService and other read operations on the
// crypto service after keys are present
// 1.  Adding a key succeeds
// 2.  Getting the key should return the same key, without error
// 3.  Removing the key succeeds
func AddGetKeyCryptoServiceInterfaceBehaviorTests(t *testing.T, cs signed.CryptoService, algo string) {
	expectedRolesToKeys := make(map[string]string)
	for i := 0; i < 2; i++ {
		var (
			addedPrivKey data.PrivateKey
			err          error
		)
		role := data.BaseRoles[i+1]
		switch algo {
		case data.RSAKey:
			addedPrivKey, err = utils.GenerateRSAKey(rand.Reader, 2048)
		case data.ECDSAKey:
			addedPrivKey, err = utils.GenerateECDSAKey(rand.Reader)
		case data.ED25519Key:
			addedPrivKey, err = utils.GenerateED25519Key(rand.Reader)
		default:
			require.FailNow(t, "invalid algorithm %s", algo)
		}
		require.NoError(t, err)
		require.NotNil(t, addedPrivKey)
		require.NoError(t, cs.AddKey(role, "docker.io/notary", addedPrivKey))
		expectedRolesToKeys[role.String()] = addedPrivKey.ID()
	}

	testGetKey(t, cs, expectedRolesToKeys, algo)
}

// AddListKeyCryptoServiceInterfaceBehaviorTests tests expected behavior for
// adding keys in a signed.CryptoService and other read operations on the
// crypto service after keys are present
// 1.  Adding a key succeeds
// 2.  Listing returns the correct number of keys and right roles
func AddListKeyCryptoServiceInterfaceBehaviorTests(t *testing.T, cs signed.CryptoService, algo string) {
	expectedRolesToKeys := make(map[string]string)
	for i := 0; i < 2; i++ {
		var (
			addedPrivKey data.PrivateKey
			err          error
		)
		role := data.BaseRoles[i+1]
		switch algo {
		case data.RSAKey:
			addedPrivKey, err = utils.GenerateRSAKey(rand.Reader, 2048)
		case data.ECDSAKey:
			addedPrivKey, err = utils.GenerateECDSAKey(rand.Reader)
		case data.ED25519Key:
			addedPrivKey, err = utils.GenerateED25519Key(rand.Reader)
		default:
			require.FailNow(t, "invalid algorithm %s", algo)
		}
		require.NoError(t, err)
		require.NotNil(t, addedPrivKey)
		require.NoError(t, cs.AddKey(role, "docker.io/notary", addedPrivKey))
		expectedRolesToKeys[role.String()] = addedPrivKey.ID()
	}

	testListKeys(t, cs, expectedRolesToKeys)
}

func testGetKey(t *testing.T, cs signed.CryptoService, expectedRolesToKeys map[string]string, algo string) {
	for role, keyID := range expectedRolesToKeys {
		pubKey := cs.GetKey(keyID)
		require.NotNil(t, pubKey)
		require.Equal(t, keyID, pubKey.ID())
		require.Equal(t, algo, pubKey.Algorithm())

		privKey, gotRole, err := cs.GetPrivateKey(keyID)
		require.NoError(t, err)
		require.NotNil(t, privKey)
		require.Equal(t, keyID, privKey.ID())
		require.Equal(t, algo, privKey.Algorithm())
		require.EqualValues(t, role, gotRole)

		require.NoError(t, cs.RemoveKey(keyID))
		require.Nil(t, cs.GetKey(keyID))
	}
}

func testListKeys(t *testing.T, cs signed.CryptoService, expectedRolesToKeys map[string]string) {
	for _, role := range append(data.BaseRoles, "targets/delegation", "invalid") {
		keys := cs.ListKeys(role)

		if keyID, ok := expectedRolesToKeys[role.String()]; ok {
			require.Len(t, keys, 1)
			require.Equal(t, keyID, keys[0])
		} else {
			require.Len(t, keys, 0)
		}
	}

	keys := cs.ListAllKeys()
	require.Len(t, keys, len(expectedRolesToKeys))
	for role, keyID := range expectedRolesToKeys {
		require.Equal(t, data.RoleName(role), keys[keyID])
	}
}
