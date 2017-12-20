package snapshot

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/docker/go/canonical/json"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/tuf/testutils"
	"github.com/stretchr/testify/require"
)

func TestSnapshotExpired(t *testing.T) {
	sn := &data.SignedSnapshot{
		Signatures: nil,
		Signed: data.Snapshot{
			SignedCommon: data.SignedCommon{Expires: time.Now().AddDate(-1, 0, 0)},
		},
	}
	require.True(t, snapshotExpired(sn), "Snapshot should have expired")
}

func TestSnapshotNotExpired(t *testing.T) {
	sn := &data.SignedSnapshot{
		Signatures: nil,
		Signed: data.Snapshot{
			SignedCommon: data.SignedCommon{Expires: time.Now().AddDate(1, 0, 0)},
		},
	}
	require.False(t, snapshotExpired(sn), "Snapshot should NOT have expired")
}

func TestGetSnapshotKeyCreate(t *testing.T) {
	store := storage.NewMemStorage()
	crypto := signed.NewEd25519()
	k, err := GetOrCreateSnapshotKey("gun", store, crypto, data.ED25519Key)
	require.Nil(t, err, "Expected nil error")
	require.NotNil(t, k, "Key should not be nil")

	k2, err := GetOrCreateSnapshotKey("gun", store, crypto, data.ED25519Key)

	require.Nil(t, err, "Expected nil error")

	// trying to get the key for the same gun and role will create a new key unless we have existing TUF metadata
	require.NotEqual(t, k, k2, "Did not receive same key when attempting to recreate.")
	require.NotNil(t, k2, "Key should not be nil")
}

type FailingStore struct {
	*storage.MemStorage
}

func (f FailingStore) GetCurrent(gun data.GUN, role data.RoleName) (*time.Time, []byte, error) {
	return nil, nil, fmt.Errorf("failing store failed")
}

func TestGetSnapshotKeyCreateWithFailingStore(t *testing.T) {
	store := FailingStore{storage.NewMemStorage()}
	crypto := signed.NewEd25519()
	k, err := GetOrCreateSnapshotKey("gun", store, crypto, data.ED25519Key)
	require.Error(t, err, "Expected error")
	require.Nil(t, k, "Key should be nil")
}

type CorruptedStore struct {
	*storage.MemStorage
}

func (c CorruptedStore) GetCurrent(gun data.GUN, role data.RoleName) (*time.Time, []byte, error) {
	return &time.Time{}, []byte("junk"), nil
}

func TestGetSnapshotKeyCreateWithCorruptedStore(t *testing.T) {
	store := CorruptedStore{storage.NewMemStorage()}
	crypto := signed.NewEd25519()
	k, err := GetOrCreateSnapshotKey("gun", store, crypto, data.ED25519Key)
	require.Error(t, err, "Expected error")
	require.Nil(t, k, "Key should be nil")
}

func TestGetSnapshotKeyCreateWithInvalidAlgo(t *testing.T) {
	store := storage.NewMemStorage()
	crypto := signed.NewEd25519()
	k, err := GetOrCreateSnapshotKey("gun", store, crypto, "notactuallyanalgorithm")
	require.Error(t, err, "Expected error")
	require.Nil(t, k, "Key should be nil")
}

func TestGetSnapshotKeyExistingMetadata(t *testing.T) {
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	sgnd, err := repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	rootJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)
	store := storage.NewMemStorage()
	require.NoError(t,
		store.UpdateCurrent("gun", storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: rootJSON}))

	snapshotRole, err := repo.Root.BuildBaseRole(data.CanonicalSnapshotRole)
	require.NoError(t, err)
	key, ok := snapshotRole.Keys[repo.Root.Signed.Roles[data.CanonicalSnapshotRole].KeyIDs[0]]
	require.True(t, ok)

	k, err := GetOrCreateSnapshotKey("gun", store, crypto, data.ED25519Key)
	require.Nil(t, err, "Expected nil error")
	require.NotNil(t, k, "Key should not be nil")
	require.Equal(t, key, k, "Did not receive same key when attempting to recreate.")
	require.NotNil(t, k, "Key should not be nil")

	k2, err := GetOrCreateSnapshotKey("gun", store, crypto, data.ED25519Key)

	require.Nil(t, err, "Expected nil error")

	require.Equal(t, k, k2, "Did not receive same key when attempting to recreate.")
	require.NotNil(t, k2, "Key should not be nil")

	// try wiping out the cryptoservice data, and ensure we create a new key because the signer doesn't hold the key specified by TUF
	crypto = signed.NewEd25519()
	k3, err := GetOrCreateSnapshotKey("gun", store, crypto, data.ED25519Key)
	require.Nil(t, err, "Expected nil error")
	require.NotEqual(t, k, k3, "Received same key when attempting to recreate.")
	require.NotEqual(t, k2, k3, "Received same key when attempting to recreate.")
	require.NotNil(t, k3, "Key should not be nil")
}

// If there is no previous snapshot or the previous snapshot is corrupt, then
// even if everything else is in place, getting the snapshot fails
func TestGetSnapshotNoPreviousSnapshot(t *testing.T) {
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	sgnd, err := repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	rootJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)

	for _, snapshotJSON := range [][]byte{nil, []byte("invalid JSON")} {
		store := storage.NewMemStorage()

		// so we know it's not a failure in getting root
		require.NoError(t,
			store.UpdateCurrent("gun", storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: rootJSON}))

		if snapshotJSON != nil {
			require.NoError(t,
				store.UpdateCurrent("gun",
					storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0, Data: snapshotJSON}))
		}

		hashBytes := sha256.Sum256(snapshotJSON)
		hashHex := hex.EncodeToString(hashBytes[:])

		_, _, err = GetOrCreateSnapshot("gun", hashHex, store, crypto)
		require.Error(t, err, "GetSnapshot should have failed")
		if snapshotJSON == nil {
			require.IsType(t, storage.ErrNotFound{}, err)
		} else {
			require.IsType(t, &json.SyntaxError{}, err)
		}
	}
}

// If there WAS a pre-existing snapshot, and it is not expired, then just return it (it doesn't
// load any other metadata that it doesn't need)
func TestGetSnapshotReturnsPreviousSnapshotIfUnexpired(t *testing.T) {
	store := storage.NewMemStorage()
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	// create an expired snapshot
	sgnd, err := repo.SignSnapshot(data.DefaultExpires(data.CanonicalSnapshotRole))
	require.NoError(t, err)
	snapshotJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)

	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0, Data: snapshotJSON}))

	hashBytes := sha256.Sum256(snapshotJSON)
	hashHex := hex.EncodeToString(hashBytes[:])

	// test when db is missing the role data (no root)
	_, gottenSnapshot, err := GetOrCreateSnapshot("gun", hashHex, store, crypto)
	require.NoError(t, err, "GetSnapshot should not have failed")
	require.True(t, bytes.Equal(snapshotJSON, gottenSnapshot))
}

func TestGetSnapshotOldSnapshotExpired(t *testing.T) {
	store := storage.NewMemStorage()
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	sgnd, err := repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	rootJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)

	// create an expired snapshot
	sgnd, err = repo.SignSnapshot(time.Now().AddDate(-1, -1, -1))
	require.True(t, repo.Snapshot.Signed.Expires.Before(time.Now()))
	require.NoError(t, err)
	snapshotJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)

	// set all the metadata
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: rootJSON}))
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0, Data: snapshotJSON}))

	hashBytes := sha256.Sum256(snapshotJSON)
	hashHex := hex.EncodeToString(hashBytes[:])

	_, gottenSnapshot, err := GetOrCreateSnapshot("gun", hashHex, store, crypto)
	require.NoError(t, err, "GetSnapshot errored")

	require.False(t, bytes.Equal(snapshotJSON, gottenSnapshot),
		"Snapshot was not regenerated when old one was expired")

	signedMeta := &data.SignedMeta{}
	require.NoError(t, json.Unmarshal(gottenSnapshot, signedMeta))
	// the new metadata is not expired
	require.True(t, signedMeta.Signed.Expires.After(time.Now()))
}

// If the root is missing or corrupt, no snapshot can be generated
func TestCannotMakeNewSnapshotIfNoRoot(t *testing.T) {
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	// create an expired snapshot
	_, err = repo.SignSnapshot(time.Now().AddDate(-1, -1, -1))
	require.True(t, repo.Snapshot.Signed.Expires.Before(time.Now()))
	require.NoError(t, err)
	snapshotJSON, err := json.Marshal(repo.Snapshot)
	require.NoError(t, err)

	for _, rootJSON := range [][]byte{nil, []byte("invalid JSON")} {
		store := storage.NewMemStorage()

		if rootJSON != nil {
			require.NoError(t, store.UpdateCurrent("gun",
				storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: rootJSON}))
		}
		require.NoError(t, store.UpdateCurrent("gun",
			storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 1, Data: snapshotJSON}))

		hashBytes := sha256.Sum256(snapshotJSON)
		hashHex := hex.EncodeToString(hashBytes[:])

		_, _, err := GetOrCreateSnapshot("gun", hashHex, store, crypto)
		require.Error(t, err, "GetSnapshot errored")

		if rootJSON == nil { // missing metadata
			require.IsType(t, storage.ErrNotFound{}, err)
		} else {
			require.IsType(t, &json.SyntaxError{}, err)
		}
	}
}

func TestCreateSnapshotNoKeyInCrypto(t *testing.T) {
	store := storage.NewMemStorage()
	repo, _, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	sgnd, err := repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	rootJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)

	// create an expired snapshot
	sgnd, err = repo.SignSnapshot(time.Now().AddDate(-1, -1, -1))
	require.True(t, repo.Snapshot.Signed.Expires.Before(time.Now()))
	require.NoError(t, err)
	snapshotJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)

	// set all the metadata so we know the failure to sign is just because of the key
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: rootJSON}))
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0, Data: snapshotJSON}))

	hashBytes := sha256.Sum256(snapshotJSON)
	hashHex := hex.EncodeToString(hashBytes[:])

	// pass it a new cryptoservice without the key
	_, _, err = GetOrCreateSnapshot("gun", hashHex, store, signed.NewEd25519())
	require.Error(t, err)
	require.IsType(t, signed.ErrInsufficientSignatures{}, err)
}
