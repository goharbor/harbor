package cryptoservice

import (
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/testutils/interfaces"
	"github.com/docker/notary/tuf/utils"
)

var algoToSigType = map[string]data.SigAlgorithm{
	data.ECDSAKey:   data.ECDSASignature,
	data.ED25519Key: data.EDDSASignature,
	data.RSAKey:     data.RSAPSSSignature,
}

var passphraseRetriever = func(string, string, bool, int) (string, bool, error) { return "", false, nil }

type CryptoServiceTester struct {
	role    data.RoleName
	keyAlgo string
	gun     data.GUN
}

func (c CryptoServiceTester) cryptoServiceFactory() *CryptoService {
	return NewCryptoService(trustmanager.NewKeyMemoryStore(passphraseRetriever))
}

// asserts that created key exists
func (c CryptoServiceTester) TestCreateAndGetKey(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()

	// Test Create
	tufKey, err := cryptoService.Create(c.role, c.gun, c.keyAlgo)
	require.NoError(t, err, c.errorMsg("error creating key"))

	// Test GetKey
	retrievedKey := cryptoService.GetKey(tufKey.ID())
	require.NotNil(t, retrievedKey,
		c.errorMsg("Could not find key ID %s", tufKey.ID()))
	require.Equal(t, tufKey.Public(), retrievedKey.Public(),
		c.errorMsg("retrieved public key didn't match"))

	// Test GetPrivateKey
	retrievedKey, alias, err := cryptoService.GetPrivateKey(tufKey.ID())
	require.NoError(t, err)
	require.Equal(t, tufKey.ID(), retrievedKey.ID(),
		c.errorMsg("retrieved private key didn't have the right ID"))
	require.Equal(t, c.role, alias)
}

// If there are multiple keystores, ensure that a key is only added to one -
// the first in the list of keyStores (which is in order of preference)
func (c CryptoServiceTester) TestCreateAndGetWhenMultipleKeystores(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()
	cryptoService.keyStores = append(cryptoService.keyStores,
		trustmanager.NewKeyMemoryStore(passphraseRetriever))

	// Test Create
	tufKey, err := cryptoService.Create(c.role, c.gun, c.keyAlgo)
	require.NoError(t, err, c.errorMsg("error creating key"))

	// Only the first keystore should have the key
	keyPath := tufKey.ID()
	_, _, err = cryptoService.keyStores[0].GetKey(keyPath)
	require.NoError(t, err, c.errorMsg(
		"First keystore does not have the key %s", keyPath))
	_, _, err = cryptoService.keyStores[1].GetKey(keyPath)
	require.Error(t, err, c.errorMsg(
		"Second keystore has the key %s", keyPath))

	// GetKey works across multiple keystores
	retrievedKey := cryptoService.GetKey(tufKey.ID())
	require.NotNil(t, retrievedKey,
		c.errorMsg("Could not find key ID %s", tufKey.ID()))
}

// asserts that getting key fails for a non-existent key
func (c CryptoServiceTester) TestGetNonexistentKey(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()

	require.Nil(t, cryptoService.GetKey("boguskeyid"),
		c.errorMsg("non-nil result for bogus keyid"))

	_, _, err := cryptoService.GetPrivateKey("boguskeyid")
	require.Error(t, err)
	// The underlying error has been correctly propagated.
	_, ok := err.(trustmanager.ErrKeyNotFound)
	require.True(t, ok)
}

// asserts that signing with a created key creates a valid signature
func (c CryptoServiceTester) TestSignWithKey(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()
	content := []byte("this is a secret")

	tufKey, err := cryptoService.Create(c.role, c.gun, c.keyAlgo)
	require.NoError(t, err, c.errorMsg("error creating key"))

	// Test Sign
	privKey, role, err := cryptoService.GetPrivateKey(tufKey.ID())
	require.NoError(t, err, c.errorMsg("failed to get private key"))
	require.Equal(t, c.role, role)

	signature, err := privKey.Sign(rand.Reader, content, nil)
	require.NoError(t, err, c.errorMsg("signing failed"))

	verifier, ok := signed.Verifiers[algoToSigType[c.keyAlgo]]
	require.True(t, ok, c.errorMsg("Unknown verifier for algorithm"))

	err = verifier.Verify(tufKey, signature, content)
	require.NoError(t, err,
		c.errorMsg("verification failed for %s key type", c.keyAlgo))
}

// asserts that signing, if there are no matching keys, produces no signatures
func (c CryptoServiceTester) TestSignNoMatchingKeys(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, c.errorMsg("error creating key"))

	// Test Sign
	_, _, err = cryptoService.GetPrivateKey(privKey.ID())
	require.Error(t, err, c.errorMsg("Should not have found private key"))
}

// Test GetPrivateKey succeeds when multiple keystores have the same key
func (c CryptoServiceTester) TestGetPrivateKeyMultipleKeystores(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()
	cryptoService.keyStores = append(cryptoService.keyStores,
		trustmanager.NewKeyMemoryStore(passphraseRetriever))

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, c.errorMsg("error creating key"))

	for _, store := range cryptoService.keyStores {
		err := store.AddKey(trustmanager.KeyInfo{Role: c.role, Gun: c.gun}, privKey)
		require.NoError(t, err)
	}

	foundKey, role, err := cryptoService.GetPrivateKey(privKey.ID())
	require.NoError(t, err, c.errorMsg("failed to get private key"))
	require.Equal(t, c.role, role)
	require.Equal(t, privKey.ID(), foundKey.ID())
}

func giveUpPassphraseRetriever(_, _ string, _ bool, _ int) (string, bool, error) {
	return "", true, nil
}

// Test that ErrPasswordInvalid is correctly propagated
func (c CryptoServiceTester) TestGetPrivateKeyPasswordInvalid(t *testing.T) {
	tempBaseDir, err := ioutil.TempDir("", "cs-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	defer os.RemoveAll(tempBaseDir)

	// Do not use c.cryptoServiceFactory(), we need a KeyFileStore.
	retriever := passphrase.ConstantRetriever("password")
	store, err := trustmanager.NewKeyFileStore(tempBaseDir, retriever)
	require.NoError(t, err)
	cryptoService := NewCryptoService(store)
	pubKey, err := cryptoService.Create(c.role, c.gun, c.keyAlgo)
	require.NoError(t, err, "error generating key: %s", err)

	// cryptoService's FileKeyStore caches the unlocked private key, so to test
	// private key unlocking we need a new instance.
	store, err = trustmanager.NewKeyFileStore(tempBaseDir, giveUpPassphraseRetriever)
	require.NoError(t, err)
	cryptoService = NewCryptoService(store)

	_, _, err = cryptoService.GetPrivateKey(pubKey.ID())
	require.EqualError(t, err, trustmanager.ErrPasswordInvalid{}.Error())
}

// Test that ErrAtttemptsExceeded is correctly propagated
func (c CryptoServiceTester) TestGetPrivateKeyAttemptsExceeded(t *testing.T) {
	tempBaseDir, err := ioutil.TempDir("", "cs-test-")
	require.NoError(t, err, "failed to create a temporary directory: %s", err)
	defer os.RemoveAll(tempBaseDir)

	// Do not use c.cryptoServiceFactory(), we need a KeyFileStore.
	retriever := passphrase.ConstantRetriever("password")
	store, err := trustmanager.NewKeyFileStore(tempBaseDir, retriever)
	require.NoError(t, err)
	cryptoService := NewCryptoService(store)
	pubKey, err := cryptoService.Create(c.role, c.gun, c.keyAlgo)
	require.NoError(t, err, "error generating key: %s", err)

	// trustmanager.KeyFileStore and trustmanager.KeyMemoryStore both cache the unlocked
	// private key, so to test private key unlocking we need a new instance using the
	// same underlying storage; this also makes trustmanager.KeyMemoryStore (and
	// c.cryptoServiceFactory()) unsuitable.
	retriever = passphrase.ConstantRetriever("incorrect password")
	store, err = trustmanager.NewKeyFileStore(tempBaseDir, retriever)
	require.NoError(t, err)
	cryptoService = NewCryptoService(store)

	_, _, err = cryptoService.GetPrivateKey(pubKey.ID())
	require.EqualError(t, err, trustmanager.ErrAttemptsExceeded{}.Error())
}

// asserts that removing key that exists succeeds
func (c CryptoServiceTester) TestRemoveCreatedKey(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()

	tufKey, err := cryptoService.Create(c.role, c.gun, c.keyAlgo)
	require.NoError(t, err, c.errorMsg("error creating key"))
	require.NotNil(t, cryptoService.GetKey(tufKey.ID()))

	// Test RemoveKey
	err = cryptoService.RemoveKey(tufKey.ID())
	require.NoError(t, err, c.errorMsg("could not remove key"))
	retrievedKey := cryptoService.GetKey(tufKey.ID())
	require.Nil(t, retrievedKey, c.errorMsg("remove didn't work"))
}

// asserts that removing key will remove it from all keystores
func (c CryptoServiceTester) TestRemoveFromMultipleKeystores(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()
	cryptoService.keyStores = append(cryptoService.keyStores,
		trustmanager.NewKeyMemoryStore(passphraseRetriever))

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err, c.errorMsg("error creating key"))

	for _, store := range cryptoService.keyStores {
		err := store.AddKey(trustmanager.KeyInfo{Role: data.CanonicalRootRole, Gun: ""}, privKey)
		require.NoError(t, err)
	}

	require.NotNil(t, cryptoService.GetKey(privKey.ID()))

	// Remove removes it from all key stores
	err = cryptoService.RemoveKey(privKey.ID())
	require.NoError(t, err, c.errorMsg("could not remove key"))

	for _, store := range cryptoService.keyStores {
		_, _, err := store.GetKey(privKey.ID())
		require.Error(t, err)
	}
}

// asserts that listing keys works with multiple keystores, and that the
// same keys are deduplicated
func (c CryptoServiceTester) TestListFromMultipleKeystores(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()
	cryptoService.keyStores = append(cryptoService.keyStores,
		trustmanager.NewKeyMemoryStore(passphraseRetriever))

	expectedKeysIDs := make(map[string]bool) // just want to be able to index by key

	for i := 0; i < 3; i++ {
		privKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err, c.errorMsg("error creating key"))
		expectedKeysIDs[privKey.ID()] = true

		// adds one different key to each keystore, and then one key to
		// both keystores
		for j, store := range cryptoService.keyStores {
			if i == j || i == 2 {
				store.AddKey(trustmanager.KeyInfo{Role: data.CanonicalRootRole, Gun: ""}, privKey)
			}
		}
	}
	// sanity check - each should have 2
	for _, store := range cryptoService.keyStores {
		require.Len(t, store.ListKeys(), 2, c.errorMsg("added keys wrong"))
	}

	keyList := cryptoService.ListKeys("root")
	require.Len(t, keyList, 4,
		c.errorMsg(
			"ListKeys should have 4 keys (not necessarily unique) but does not: %v", keyList))
	for _, k := range keyList {
		_, ok := expectedKeysIDs[k]
		require.True(t, ok, c.errorMsg("Unexpected key %s", k))
	}

	keyMap := cryptoService.ListAllKeys()
	require.Len(t, keyMap, 3,
		c.errorMsg("ListAllKeys should have 3 unique keys but does not: %v", keyMap))

	for k, role := range keyMap {
		_, ok := expectedKeysIDs[k]
		require.True(t, ok)
		require.Equal(t, data.RoleName("root"), role)
	}
}

// asserts that adding a key adds to just the first keystore
// and adding an existing key either succeeds if the role matches or fails if it does not
func (c CryptoServiceTester) TestAddKey(t *testing.T) {
	cryptoService := c.cryptoServiceFactory()
	cryptoService.keyStores = append(cryptoService.keyStores,
		trustmanager.NewKeyMemoryStore(passphraseRetriever))

	privKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	// Add the key to the targets role
	require.NoError(t, cryptoService.AddKey(data.CanonicalTargetsRole, c.gun, privKey))

	// Check that we added the key and its info to only the first keystore
	retrievedKey, retrievedRole, err := cryptoService.keyStores[0].GetKey(privKey.ID())
	require.NoError(t, err)
	require.Equal(t, privKey.Private(), retrievedKey.Private())
	require.Equal(t, data.CanonicalTargetsRole, retrievedRole)

	retrievedKeyInfo, err := cryptoService.keyStores[0].GetKeyInfo(privKey.ID())
	require.NoError(t, err)
	require.Equal(t, data.CanonicalTargetsRole, retrievedKeyInfo.Role)
	require.Equal(t, c.gun, retrievedKeyInfo.Gun)

	// The key should not exist in the second keystore
	_, _, err = cryptoService.keyStores[1].GetKey(privKey.ID())
	require.Error(t, err)
	_, err = cryptoService.keyStores[1].GetKeyInfo(privKey.ID())
	require.Error(t, err)

	// We should be able to successfully get the key from the cryptoservice level
	retrievedKey, retrievedRole, err = cryptoService.GetPrivateKey(privKey.ID())
	require.NoError(t, err)
	require.Equal(t, privKey.Private(), retrievedKey.Private())
	require.Equal(t, data.CanonicalTargetsRole, retrievedRole)
	retrievedKeyInfo, err = cryptoService.GetKeyInfo(privKey.ID())
	require.NoError(t, err)
	require.Equal(t, data.CanonicalTargetsRole, retrievedKeyInfo.Role)
	require.Equal(t, c.gun, retrievedKeyInfo.Gun)

	// Add the same key to the targets role, since the info is the same we should have no error
	require.NoError(t, cryptoService.AddKey(data.CanonicalTargetsRole, c.gun, privKey))

	// Try to add the same key to the snapshot role, which should error due to the role mismatch
	require.Error(t, cryptoService.AddKey(data.CanonicalSnapshotRole, c.gun, privKey))
}

// Prints out an error message with information about the key algorithm,
// role, and test name. Ideally we could generate different tests given
// data, without having to put for loops in one giant test function, but
// that involves a lot of boilerplate.  So as a compromise, everything will
// still be run in for loops in one giant test function, but we can at
// least provide an error message stating what data/helper test function
// failed.
func (c CryptoServiceTester) errorMsg(message string, args ...interface{}) string {
	pc := make([]uintptr, 10) // at least 1 entry needed
	runtime.Callers(2, pc)    // the caller of errorMsg
	f := runtime.FuncForPC(pc[0])
	return fmt.Sprintf("%s (role: %s, keyAlgo: %s): %s", f.Name(), c.role,
		c.keyAlgo, fmt.Sprintf(message, args...))
}

func testCryptoService(t *testing.T, gun data.GUN) {
	roles := []data.RoleName{
		data.CanonicalRootRole,
		data.CanonicalTargetsRole,
		data.CanonicalSnapshotRole,
		data.CanonicalTimestampRole,
	}

	for _, role := range roles {
		for algo := range algoToSigType {
			cst := CryptoServiceTester{
				role:    role,
				keyAlgo: algo,
				gun:     gun,
			}
			cst.TestAddKey(t)
			cst.TestCreateAndGetKey(t)
			cst.TestCreateAndGetWhenMultipleKeystores(t)
			cst.TestGetNonexistentKey(t)
			cst.TestSignWithKey(t)
			cst.TestSignNoMatchingKeys(t)
			cst.TestGetPrivateKeyMultipleKeystores(t)
			cst.TestRemoveCreatedKey(t)
			cst.TestRemoveFromMultipleKeystores(t)
			cst.TestListFromMultipleKeystores(t)
			cst.TestGetPrivateKeyPasswordInvalid(t)
			cst.TestGetPrivateKeyAttemptsExceeded(t)
		}
	}
}

func TestCryptoServiceWithNonEmptyGUN(t *testing.T) {
	testCryptoService(t, "org/repo")
}

func TestCryptoServiceWithEmptyGUN(t *testing.T) {
	testCryptoService(t, "")
}

// CryptoSigner conforms to the signed.CryptoService interface behavior
func TestCryptoSignerInterfaceBehavior(t *testing.T) {
	cs := NewCryptoService(trustmanager.NewKeyMemoryStore(passphraseRetriever))
	interfaces.EmptyCryptoServiceInterfaceBehaviorTests(t, cs)
	interfaces.CreateGetKeyCryptoServiceInterfaceBehaviorTests(t, cs, data.ECDSAKey)

	cs = NewCryptoService(trustmanager.NewKeyMemoryStore(passphraseRetriever))
	interfaces.CreateListKeyCryptoServiceInterfaceBehaviorTests(t, cs, data.ECDSAKey)

	cs = NewCryptoService(trustmanager.NewKeyMemoryStore(passphraseRetriever))
	interfaces.AddGetKeyCryptoServiceInterfaceBehaviorTests(t, cs, data.ECDSAKey)

	cs = NewCryptoService(trustmanager.NewKeyMemoryStore(passphraseRetriever))
	interfaces.AddListKeyCryptoServiceInterfaceBehaviorTests(t, cs, data.ECDSAKey)
}
