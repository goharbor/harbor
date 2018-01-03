package timestamp

import (
	"bytes"
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

func TestTimestampExpired(t *testing.T) {
	ts := &data.SignedTimestamp{
		Signatures: nil,
		Signed: data.Timestamp{
			SignedCommon: data.SignedCommon{Expires: time.Now().AddDate(-1, 0, 0)},
		},
	}
	require.True(t, timestampExpired(ts), "Timestamp should have expired")
}

func TestTimestampNotExpired(t *testing.T) {
	ts := &data.SignedTimestamp{
		Signatures: nil,
		Signed: data.Timestamp{
			SignedCommon: data.SignedCommon{Expires: time.Now().AddDate(1, 0, 0)},
		},
	}
	require.False(t, timestampExpired(ts), "Timestamp should NOT have expired")
}

func TestGetTimestampKey(t *testing.T) {
	store := storage.NewMemStorage()
	crypto := signed.NewEd25519()
	k, err := GetOrCreateTimestampKey("gun", store, crypto, data.ED25519Key)
	require.Nil(t, err, "Expected nil error")
	require.NotNil(t, k, "Key should not be nil")

	k2, err := GetOrCreateTimestampKey("gun", store, crypto, data.ED25519Key)

	require.Nil(t, err, "Expected nil error")

	// Note that this cryptoservice does not perform any rate-limiting, unlike the notary-signer,
	// so we get a different key until we've published valid TUF metadata in the store
	require.NotEqual(t, k, k2, "Received same key when attempting to recreate.")
	require.NotNil(t, k2, "Key should not be nil")
}

// If there is no previous timestamp or the previous timestamp is corrupt, then
// even if everything else is in place, getting the timestamp fails
func TestGetTimestampNoPreviousTimestamp(t *testing.T) {
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	for _, timestampJSON := range [][]byte{nil, []byte("invalid JSON")} {
		store := storage.NewMemStorage()

		var gun data.GUN = "gun"
		// so we know it's not a failure in getting root or snapshot
		require.NoError(t,
			store.UpdateCurrent(gun, storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0,
				Data: meta[data.CanonicalRootRole]}))
		require.NoError(t,
			store.UpdateCurrent(gun, storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0,
				Data: meta[data.CanonicalSnapshotRole]}))

		if timestampJSON != nil {
			require.NoError(t,
				store.UpdateCurrent(gun,
					storage.MetaUpdate{Role: data.CanonicalTimestampRole, Version: 0, Data: timestampJSON}))
		}

		_, _, err = GetOrCreateTimestamp(gun, store, crypto)
		require.Error(t, err, "GetTimestamp should have failed")
		if timestampJSON == nil {
			require.IsType(t, storage.ErrNotFound{}, err)
		} else {
			require.IsType(t, &json.SyntaxError{}, err)
		}
	}
}

// If there WAS a pre-existing timestamp, and it is not expired, then just return it (it doesn't
// load any other metadata that it doesn't need, like root)
func TestGetTimestampReturnsPreviousTimestampIfUnexpired(t *testing.T) {
	store := storage.NewMemStorage()
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0, Data: meta[data.CanonicalSnapshotRole]}))
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalTimestampRole, Version: 0, Data: meta[data.CanonicalTimestampRole]}))

	_, gottenTimestamp, err := GetOrCreateTimestamp("gun", store, crypto)
	require.NoError(t, err, "GetTimestamp should not have failed")
	require.True(t, bytes.Equal(meta[data.CanonicalTimestampRole], gottenTimestamp))
}

func TestGetTimestampOldTimestampExpired(t *testing.T) {
	store := storage.NewMemStorage()
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	// create an expired timestamp
	_, err = repo.SignTimestamp(time.Now().AddDate(-1, -1, -1))
	require.True(t, repo.Timestamp.Signed.Expires.Before(time.Now()))
	require.NoError(t, err)
	timestampJSON, err := json.Marshal(repo.Timestamp)
	require.NoError(t, err)

	// set all the metadata
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: meta[data.CanonicalRootRole]}))
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0, Data: meta[data.CanonicalSnapshotRole]}))
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalTimestampRole, Version: 1, Data: timestampJSON}))

	_, gottenTimestamp, err := GetOrCreateTimestamp("gun", store, crypto)
	require.NoError(t, err, "GetTimestamp errored")

	require.False(t, bytes.Equal(timestampJSON, gottenTimestamp),
		"Timestamp was not regenerated when old one was expired")

	signedMeta := &data.SignedMeta{}
	require.NoError(t, json.Unmarshal(gottenTimestamp, signedMeta))
	// the new metadata is not expired
	require.True(t, signedMeta.Signed.Expires.After(time.Now()))
}

// If the root or snapshot is missing or corrupt, no timestamp can be generated
func TestCannotMakeNewTimestampIfNoRootOrSnapshot(t *testing.T) {
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	// create an expired timestamp
	_, err = repo.SignTimestamp(time.Now().AddDate(-1, -1, -1))
	require.True(t, repo.Timestamp.Signed.Expires.Before(time.Now()))
	require.NoError(t, err)
	timestampJSON, err := json.Marshal(repo.Timestamp)
	require.NoError(t, err)

	rootJSON := meta[data.CanonicalRootRole]
	snapJSON := meta[data.CanonicalSnapshotRole]

	invalids := []struct {
		test map[string][]byte
		err  error
	}{
		{
			test: map[string][]byte{data.CanonicalRootRole.String(): rootJSON, data.CanonicalSnapshotRole.String(): []byte("invalid JSON")},
			err:  storage.ErrNotFound{},
		},
		{
			test: map[string][]byte{data.CanonicalRootRole.String(): []byte("invalid JSON"), data.CanonicalSnapshotRole.String(): snapJSON},
			err:  &json.SyntaxError{},
		},
		{
			test: map[string][]byte{data.CanonicalRootRole.String(): rootJSON},
			err:  storage.ErrNotFound{},
		},
		{
			test: map[string][]byte{data.CanonicalSnapshotRole.String(): snapJSON},
			err:  storage.ErrNotFound{},
		},
	}

	for _, test := range invalids {
		dataToSet := test.test
		store := storage.NewMemStorage()
		for roleName, jsonBytes := range dataToSet {
			require.NoError(t, store.UpdateCurrent("gun",
				storage.MetaUpdate{Role: data.RoleName(roleName), Version: 0, Data: jsonBytes}))
		}
		require.NoError(t, store.UpdateCurrent("gun",
			storage.MetaUpdate{Role: data.CanonicalTimestampRole, Version: 1, Data: timestampJSON}))

		_, _, err := GetOrCreateTimestamp("gun", store, crypto)
		require.Error(t, err, "GetTimestamp errored")
		require.IsType(t, test.err, err)
	}
}

func TestCreateTimestampNoKeyInCrypto(t *testing.T) {
	store := storage.NewMemStorage()
	repo, _, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	meta, err := testutils.SignAndSerialize(repo)
	require.NoError(t, err)

	// create an expired timestamp
	_, err = repo.SignTimestamp(time.Now().AddDate(-1, -1, -1))
	require.True(t, repo.Timestamp.Signed.Expires.Before(time.Now()))
	require.NoError(t, err)
	timestampJSON, err := json.Marshal(repo.Timestamp)
	require.NoError(t, err)

	// set all the metadata so we know the failure to sign is just because of the key
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: meta[data.CanonicalRootRole]}))
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalSnapshotRole, Version: 0, Data: meta[data.CanonicalSnapshotRole]}))
	require.NoError(t, store.UpdateCurrent("gun",
		storage.MetaUpdate{Role: data.CanonicalTimestampRole, Version: 1, Data: timestampJSON}))

	// pass it a new cryptoservice without the key
	_, _, err = GetOrCreateTimestamp("gun", store, signed.NewEd25519())
	require.Error(t, err)
	require.IsType(t, signed.ErrInsufficientSignatures{}, err)
}

type FailingStore struct {
	*storage.MemStorage
}

func (f FailingStore) GetCurrent(gun data.GUN, role data.RoleName) (*time.Time, []byte, error) {
	return nil, nil, fmt.Errorf("failing store failed")
}

func TestGetTimestampKeyCreateWithFailingStore(t *testing.T) {
	store := FailingStore{storage.NewMemStorage()}
	crypto := signed.NewEd25519()
	k, err := GetOrCreateTimestampKey("gun", store, crypto, data.ED25519Key)
	require.Error(t, err, "Expected error")
	require.Nil(t, k, "Key should be nil")
}

type CorruptedStore struct {
	*storage.MemStorage
}

func (c CorruptedStore) GetCurrent(gun data.GUN, role data.RoleName) (*time.Time, []byte, error) {
	return &time.Time{}, []byte("junk"), nil
}

func TestGetTimestampKeyCreateWithCorruptedStore(t *testing.T) {
	store := CorruptedStore{storage.NewMemStorage()}
	crypto := signed.NewEd25519()
	k, err := GetOrCreateTimestampKey("gun", store, crypto, data.ED25519Key)
	require.Error(t, err, "Expected error")
	require.Nil(t, k, "Key should be nil")
}

func TestGetTimestampKeyCreateWithInvalidAlgo(t *testing.T) {
	store := storage.NewMemStorage()
	crypto := signed.NewEd25519()
	k, err := GetOrCreateTimestampKey("gun", store, crypto, "notactuallyanalgorithm")
	require.Error(t, err, "Expected error")
	require.Nil(t, k, "Key should be nil")
}

func TestGetTimestampKeyExistingMetadata(t *testing.T) {
	repo, crypto, err := testutils.EmptyRepo("gun")
	require.NoError(t, err)

	sgnd, err := repo.SignRoot(data.DefaultExpires(data.CanonicalRootRole), nil)
	require.NoError(t, err)
	rootJSON, err := json.Marshal(sgnd)
	require.NoError(t, err)
	store := storage.NewMemStorage()
	require.NoError(t,
		store.UpdateCurrent("gun", storage.MetaUpdate{Role: data.CanonicalRootRole, Version: 0, Data: rootJSON}))

	timestampRole, err := repo.Root.BuildBaseRole(data.CanonicalTimestampRole)
	require.NoError(t, err)
	key, ok := timestampRole.Keys[repo.Root.Signed.Roles[data.CanonicalTimestampRole].KeyIDs[0]]
	require.True(t, ok)

	k, err := GetOrCreateTimestampKey("gun", store, crypto, data.ED25519Key)
	require.Nil(t, err, "Expected nil error")
	require.NotNil(t, k, "Key should not be nil")
	require.Equal(t, key, k, "Did not receive same key when attempting to recreate.")
	require.NotNil(t, k, "Key should not be nil")

	k2, err := GetOrCreateTimestampKey("gun", store, crypto, data.ED25519Key)

	require.Nil(t, err, "Expected nil error")

	require.Equal(t, k, k2, "Did not receive same key when attempting to recreate.")
	require.NotNil(t, k2, "Key should not be nil")

	// try wiping out the cryptoservice data, and ensure we create a new key because the signer doesn't hold the key specified by TUF
	crypto = signed.NewEd25519()
	k3, err := GetOrCreateTimestampKey("gun", store, crypto, data.ED25519Key)
	require.Nil(t, err, "Expected nil error")
	require.NotEqual(t, k, k3, "Received same key when attempting to recreate.")
	require.NotEqual(t, k2, k3, "Received same key when attempting to recreate.")
	require.NotNil(t, k3, "Key should not be nil")
}
