package keydbstore

import (
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

// gets a key from the DB store, and asserts that the key is the expected key
func requireGetKeySuccess(t *testing.T, dbKeyService signed.CryptoService, expectedRole string, expectedKey data.PrivateKey) {
	retrKey, retrRole, err := dbKeyService.GetPrivateKey(expectedKey.ID())
	require.NoError(t, err)
	require.Equal(t, retrKey.ID(), expectedKey.ID())
	require.Equal(t, retrKey.Algorithm(), expectedKey.Algorithm())
	require.Equal(t, retrKey.Public(), expectedKey.Public())
	require.Equal(t, retrKey.Private(), expectedKey.Private())
	require.EqualValues(t, retrRole, expectedRole)
}

func requireGetPubKeySuccess(t *testing.T, dbKeyService signed.CryptoService, expectedRole string, expectedPubKey data.PublicKey) {
	retrPubKey := dbKeyService.GetKey(expectedPubKey.ID())
	require.Equal(t, retrPubKey.Public(), expectedPubKey.Public())
	require.Equal(t, retrPubKey.ID(), expectedPubKey.ID())
	require.Equal(t, retrPubKey.Algorithm(), expectedPubKey.Algorithm())
}

// closes the DB connection first so we can test that the successful get was
// from the cache
func requireGetKeySuccessFromCache(t *testing.T, cachedStore, underlyingStore signed.CryptoService, expectedRole string, expectedKey data.PrivateKey) {
	require.NoError(t, underlyingStore.RemoveKey(expectedKey.ID()))
	requireGetKeySuccess(t, cachedStore, expectedRole, expectedKey)
}

func requireGetKeyFailure(t *testing.T, dbStore signed.CryptoService, keyID string) {
	_, _, err := dbStore.GetPrivateKey(keyID)
	require.Error(t, err)
	k := dbStore.GetKey(keyID)
	require.Nil(t, k)
}

type unAddableKeyService struct {
	signed.CryptoService
}

func (u unAddableKeyService) AddKey(_ data.RoleName, _ data.GUN, _ data.PrivateKey) error {
	return fmt.Errorf("can't add to keyservice")
}

type unRemoveableKeyService struct {
	signed.CryptoService
	failToRemove bool
}

func (u unRemoveableKeyService) RemoveKey(keyID string) error {
	if u.failToRemove {
		return fmt.Errorf("can't remove from keystore")
	}
	return u.CryptoService.RemoveKey(keyID)
}

// Getting a key, on success, populates the cache.
func TestGetSuccessPopulatesCache(t *testing.T) {
	underlying := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(constRetriever))
	cached := NewCachedKeyService(underlying)

	testKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	// nothing there yet
	requireGetKeyFailure(t, cached, testKey.ID())

	// Add key to underlying store only
	err = underlying.AddKey(data.CanonicalTimestampRole, "gun", testKey)
	require.NoError(t, err)

	// getting for the first time is successful, and after that getting from cache should be too
	requireGetKeySuccess(t, cached, data.CanonicalTimestampRole.String(), testKey)
	requireGetKeySuccessFromCache(t, cached, underlying, data.CanonicalTimestampRole.String(), testKey)
}

// Creating a key, on success, populates the cache, but does not do so on failure
func TestAddKeyPopulatesCacheIfSuccessful(t *testing.T) {
	underlying := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(constRetriever))
	cached := NewCachedKeyService(underlying)

	testKeys := make([]data.PrivateKey, 2)
	for i := 0; i < 2; i++ {
		privKey, err := utils.GenerateECDSAKey(rand.Reader)
		require.NoError(t, err)
		testKeys[i] = privKey
	}

	// Writing in the key service succeeds
	err := cached.AddKey(data.CanonicalTimestampRole, "gun", testKeys[0])
	require.NoError(t, err)

	// Now even if it's deleted from the underlying database, it's fine because it's cached
	requireGetKeySuccessFromCache(t, cached, underlying, data.CanonicalTimestampRole.String(), testKeys[0])

	// Writing in the key service fails
	cached = NewCachedKeyService(unAddableKeyService{underlying})
	err = cached.AddKey(data.CanonicalTimestampRole, "gun", testKeys[1])
	require.Error(t, err)

	// And now it can't be found in either DB
	requireGetKeyFailure(t, cached, testKeys[1].ID())
}

// Deleting a key, no matter whether we succeed in the underlying layer or not, evicts the cached key.
func TestDeleteKeyRemovesKeyFromCache(t *testing.T) {
	underlying := cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(constRetriever))
	cached := NewCachedKeyService(underlying)

	testKey, err := utils.GenerateECDSAKey(rand.Reader)
	require.NoError(t, err)

	// Write the key, which puts it in the cache
	err = cached.AddKey(data.CanonicalTimestampRole, "gun", testKey)
	require.NoError(t, err)

	// Deleting removes the key from the cache and the underlying store
	err = cached.RemoveKey(testKey.ID())
	require.NoError(t, err)
	requireGetKeyFailure(t, cached, testKey.ID())

	// Now set up an underlying store where the key can't be deleted
	failingUnderlying := unRemoveableKeyService{CryptoService: underlying, failToRemove: true}
	cached = NewCachedKeyService(failingUnderlying)
	err = cached.AddKey(data.CanonicalTimestampRole, "gun", testKey)
	require.NoError(t, err)

	// Deleting fails to remove the key from the underlying store
	err = cached.RemoveKey(testKey.ID())
	require.Error(t, err)
	requireGetKeySuccess(t, failingUnderlying, data.CanonicalTimestampRole.String(), testKey)

	// now actually remove the key from the underlying store to test that it's gone from the cache
	failingUnderlying.failToRemove = false
	require.NoError(t, failingUnderlying.RemoveKey(testKey.ID()))

	// and it's not in the cache
	requireGetKeyFailure(t, cached, testKey.ID())
}
