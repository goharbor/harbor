package storage

import (
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/testutils"
	"github.com/stretchr/testify/require"
)

// Produce a series of tufMeta objects and updates given a TUF repo
func metaFromRepo(t *testing.T, gun data.GUN, version int) map[string]StoredTUFMeta {
	tufRepo, _, err := testutils.EmptyRepo(gun, "targets/a", "targets/a/b")
	require.NoError(t, err)

	tufRepo.Root.Signed.Version = version - 1
	tufRepo.Timestamp.Signed.Version = version - 1
	tufRepo.Snapshot.Signed.Version = version - 1
	for _, signedObj := range tufRepo.Targets {
		signedObj.Signed.Version = version - 1
	}

	metaBytes, err := testutils.SignAndSerialize(tufRepo)
	require.NoError(t, err)

	tufMeta := make(map[string]StoredTUFMeta)
	for role, tufdata := range metaBytes {
		tufMeta[role.String()] = SampleCustomTUFObj(gun, role, version, tufdata)
	}

	return tufMeta
}

// TUFMetaStore's GetCurrent walks from the current timestamp metadata
// to the snapshot specified in the checksum, to potentially other role metadata by checksum
func testTUFMetaStoreGetCurrent(t *testing.T, s MetaStore) {
	tufDBStore := NewTUFMetaStorage(s)
	var gun data.GUN = "testGUN"

	initialRootTUF := SampleCustomTUFObj(gun, data.CanonicalRootRole, 1, nil)
	ConsistentEmptyGetCurrentTest(t, tufDBStore, initialRootTUF)

	// put an initial piece of root metadata data in the database,
	// there isn't enough state to retrieve it since we require a timestamp and snapshot in our walk
	require.NoError(t, s.UpdateCurrent(gun, MakeUpdate(initialRootTUF)), "Adding root to empty store failed.")
	ConsistentMissingTSAndSnapGetCurrentTest(t, tufDBStore, initialRootTUF)

	// Note that get by checksum succeeds, since it does not try to walk timestamp/snapshot
	_, _, err := tufDBStore.GetChecksum(gun, data.CanonicalRootRole, initialRootTUF.SHA256)
	require.NoError(t, err)

	// Now add metadata from a valid TUF repo to ensure that we walk correctly.
	tufMetaByRole := metaFromRepo(t, gun, 2)
	updates := make([]MetaUpdate, 0, len(tufMetaByRole))
	for _, tufObj := range tufMetaByRole {
		updates = append(updates, MakeUpdate(tufObj))
	}
	require.NoError(t, s.UpdateMany(gun, updates))

	// GetCurrent on all of these roles should succeed
	for _, tufobj := range tufMetaByRole {
		ConsistentGetCurrentFoundTest(t, tufDBStore, tufobj)
	}

	// Delete snapshot by just wiping out everything in the store and adding only
	// the non-snapshot data
	require.NoError(t, s.Delete(gun), "unable to delete metadata")
	updates = make([]MetaUpdate, 0, len(updates)-1)
	for role, tufObj := range tufMetaByRole {
		if role != data.CanonicalSnapshotRole.String() {
			updates = append(updates, MakeUpdate(tufObj))
		}
	}
	require.NoError(t, s.UpdateMany(gun, updates))
	_, _, err = s.GetCurrent(gun, data.CanonicalSnapshotRole)
	require.IsType(t, ErrNotFound{}, err)

	// GetCurrent on all roles should still succeed - snapshot lookup because of caching,
	// and targets and root because the snapshot is cached
	for _, tufobj := range tufMetaByRole {
		ConsistentGetCurrentFoundTest(t, tufDBStore, tufobj)
	}

	// add another orphaned root, but ensure that we still get the previous root
	// since the new root isn't in a timestamp/snapshot chain
	orphanedRootTUF := SampleCustomTUFObj(gun, data.CanonicalRootRole, 3, []byte("orphanedRoot"))
	require.NoError(t, s.UpdateCurrent(gun, MakeUpdate(orphanedRootTUF)), "unable to create orphaned root in store")

	// a GetCurrent for this gun and root gets us the previous root, which is linked in timestamp and snapshot
	ConsistentGetCurrentFoundTest(t, tufDBStore, tufMetaByRole[data.CanonicalRootRole.String()])
	// the orphaned root fails on a GetCurrent even though it's in the underlying store
	ConsistentTSAndSnapGetDifferentCurrentTest(t, tufDBStore, orphanedRootTUF)
}

func ConsistentGetCurrentFoundTest(t *testing.T, s *TUFMetaStorage, rec StoredTUFMeta) {
	_, byt, err := s.GetCurrent(rec.Gun, rec.Role)
	require.NoError(t, err)
	require.Equal(t, rec.Data, byt)
}

// Checks that both the walking metastore and underlying metastore do not contain the TUF file
func ConsistentEmptyGetCurrentTest(t *testing.T, s *TUFMetaStorage, rec StoredTUFMeta) {
	_, byt, err := s.GetCurrent(rec.Gun, rec.Role)
	require.Nil(t, byt)
	require.Error(t, err, "There should be an error getting an empty table")
	require.IsType(t, ErrNotFound{}, err, "Should get a not found error")

	_, byt, err = s.MetaStore.GetCurrent(rec.Gun, rec.Role)
	require.Nil(t, byt)
	require.Error(t, err, "There should be an error getting an empty table")
	require.IsType(t, ErrNotFound{}, err, "Should get a not found error")
}

// Check that we can't get the "current" specified role because we can't walk from timestamp --> snapshot --> role
// Also checks that the role metadata still exists in the underlying store
func ConsistentMissingTSAndSnapGetCurrentTest(t *testing.T, s *TUFMetaStorage, rec StoredTUFMeta) {
	_, byt, err := s.GetCurrent(rec.Gun, rec.Role)
	require.Nil(t, byt)
	require.Error(t, err, "There should be an error because there is no timestamp or snapshot to use on GetCurrent")
	_, byt, err = s.MetaStore.GetCurrent(rec.Gun, rec.Role)
	require.Equal(t, rec.Data, byt)
	require.NoError(t, err)
}

// Check that we can get the "current" specified role but it is different from the provided TUF file because
// the most valid walk from timestamp --> snapshot --> role gets a different
func ConsistentTSAndSnapGetDifferentCurrentTest(t *testing.T, s *TUFMetaStorage, rec StoredTUFMeta) {
	_, byt, err := s.GetCurrent(rec.Gun, rec.Role)
	require.NotEqual(t, rec.Data, byt)
	require.NoError(t, err)
	_, byt, err = s.MetaStore.GetCurrent(rec.Gun, rec.Role)
	require.Equal(t, rec.Data, byt)
	require.NoError(t, err)
}
