package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	"github.com/docker/notary/tuf/data"
	"github.com/stretchr/testify/require"
)

type StoredTUFMeta struct {
	Gun     data.GUN
	Role    data.RoleName
	SHA256  string
	Data    []byte
	Version int
}

func SampleCustomTUFObj(gun data.GUN, role data.RoleName, version int, tufdata []byte) StoredTUFMeta {
	if tufdata == nil {
		tufdata = []byte(fmt.Sprintf("%s_%s_%d", gun, role, version))
	}
	checksum := sha256.Sum256(tufdata)
	hexChecksum := hex.EncodeToString(checksum[:])
	return StoredTUFMeta{
		Gun:     gun,
		Role:    role,
		Version: version,
		SHA256:  hexChecksum,
		Data:    tufdata,
	}
}

func MakeUpdate(tufObj StoredTUFMeta) MetaUpdate {
	return MetaUpdate{
		Role:    tufObj.Role,
		Version: tufObj.Version,
		Data:    tufObj.Data,
	}
}

func assertExpectedTUFMetaInStore(t *testing.T, s MetaStore, expected []StoredTUFMeta, current bool) {
	for _, tufObj := range expected {
		var prevTime *time.Time
		if current {
			cDate, tufdata, err := s.GetCurrent(tufObj.Gun, tufObj.Role)
			require.NoError(t, err)
			require.Equal(t, tufObj.Data, tufdata)

			// the update date was sometime wthin the last minute
			require.True(t, cDate.After(time.Now().Add(-1*time.Minute)))
			require.True(t, cDate.Before(time.Now().Add(5*time.Second)))
			prevTime = cDate
		}

		checksumBytes := sha256.Sum256(tufObj.Data)
		checksum := hex.EncodeToString(checksumBytes[:])

		cDate, tufdata, err := s.GetChecksum(tufObj.Gun, tufObj.Role, checksum)
		require.NoError(t, err)
		require.Equal(t, tufObj.Data, tufdata)

		if current {
			require.True(t, prevTime.Equal(*cDate), "%s should be equal to %s", prevTime, cDate)
		} else {
			// the update date was sometime wthin the last minute
			require.True(t, cDate.After(time.Now().Add(-1*time.Minute)))
			require.True(t, cDate.Before(time.Now().Add(5*time.Second)))
		}
	}
}

// UpdateCurrent should succeed if there was no previous metadata of the same
// gun and role.  They should be gettable.
func testUpdateCurrentEmptyStore(t *testing.T, s MetaStore) []StoredTUFMeta {
	expected := make([]StoredTUFMeta, 0, 10)
	for _, role := range append(data.BaseRoles, "targets/a") {
		for _, gun := range []data.GUN{"gun1", "gun2"} {
			// Adding a new TUF file should succeed
			tufObj := SampleCustomTUFObj(gun, role, 1, nil)
			require.NoError(t, s.UpdateCurrent(tufObj.Gun, MakeUpdate(tufObj)))
			expected = append(expected, tufObj)
		}
	}

	assertExpectedTUFMetaInStore(t, s, expected, true)
	return expected
}

// UpdateCurrent will successfully add a new (higher) version of an existing TUF file,
// but will return an error if there is an older version of a TUF file.  oldVersionExists
// specifies whether the older version should already exist in the DB or not.
func testUpdateCurrentVersionCheck(t *testing.T, s MetaStore, oldVersionExists bool) []StoredTUFMeta {
	role, gun := data.CanonicalRootRole, data.GUN("testGUN")

	expected := []StoredTUFMeta{
		SampleCustomTUFObj(gun, role, 1, nil),
		SampleCustomTUFObj(gun, role, 2, nil),
		SampleCustomTUFObj(gun, role, 4, nil),
	}

	// starting meta is version 1
	require.NoError(t, s.UpdateCurrent(gun, MakeUpdate(expected[0])))

	// inserting meta version immediately above it and skipping ahead will succeed
	require.NoError(t, s.UpdateCurrent(gun, MakeUpdate(expected[1])))
	require.NoError(t, s.UpdateCurrent(gun, MakeUpdate(expected[2])))

	// Inserting a version that already exists, or that is lower than the current version, will fail
	version := 3
	if oldVersionExists {
		version = 4
	}

	tufObj := SampleCustomTUFObj(gun, role, version, nil)
	err := s.UpdateCurrent(gun, MakeUpdate(tufObj))
	require.Error(t, err, "Error should not be nil")
	require.IsType(t, ErrOldVersion{}, err,
		"Expected ErrOldVersion error type, got: %v", err)

	assertExpectedTUFMetaInStore(t, s, expected[:2], false)
	assertExpectedTUFMetaInStore(t, s, expected[2:], true)
	return expected
}

// GetVersion should successfully retrieve a version of an existing TUF file,
// but will return an error if the requested version does not exist.
func testGetVersion(t *testing.T, s MetaStore) {
	_, _, err := s.GetVersion("gun", "role", 2)
	require.IsType(t, ErrNotFound{}, err, "Expected error to be ErrNotFound")

	s.UpdateCurrent("gun", MetaUpdate{"role", 2, []byte("version2")})
	_, d, err := s.GetVersion("gun", "role", 2)
	require.Nil(t, err, "Expected error to be nil")
	require.Equal(t, []byte("version2"), d, "Data was incorrect")

	// Getting newer version fails
	_, _, err = s.GetVersion("gun", "role", 3)
	require.IsType(t, ErrNotFound{}, err, "Expected error to be ErrNotFound")

	// Getting another gun/role fails
	_, _, err = s.GetVersion("badgun", "badrole", 2)
	require.IsType(t, ErrNotFound{}, err, "Expected error to be ErrNotFound")
}

// UpdateMany succeeds if the updates do not conflict with each other or with what's
// already in the DB
func testUpdateManyNoConflicts(t *testing.T, s MetaStore) []StoredTUFMeta {
	var gun data.GUN = "testGUN"
	firstBatch := make([]StoredTUFMeta, 4)
	updates := make([]MetaUpdate, 4)
	for i, role := range data.BaseRoles {
		firstBatch[i] = SampleCustomTUFObj(gun, role, 1, nil)
		updates[i] = MakeUpdate(firstBatch[i])
	}

	require.NoError(t, s.UpdateMany(gun, updates))
	assertExpectedTUFMetaInStore(t, s, firstBatch, true)

	secondBatch := make([]StoredTUFMeta, 4)
	// no conflicts with what's in DB or with itself
	for i, role := range data.BaseRoles {
		secondBatch[i] = SampleCustomTUFObj(gun, role, 2, nil)
		updates[i] = MakeUpdate(secondBatch[i])
	}

	require.NoError(t, s.UpdateMany(gun, updates))
	// the first batch is still there, but are no longer the current ones
	assertExpectedTUFMetaInStore(t, s, firstBatch, false)
	assertExpectedTUFMetaInStore(t, s, secondBatch, true)

	// and no conflicts if the same role and gun but different version is included
	// in the same update.  Even if they're out of order.
	thirdBatch := make([]StoredTUFMeta, 2)
	role := data.CanonicalRootRole
	updates = updates[:2]
	for i, version := range []int{4, 3} {
		thirdBatch[i] = SampleCustomTUFObj(gun, role, version, nil)
		updates[i] = MakeUpdate(thirdBatch[i])
	}

	require.NoError(t, s.UpdateMany(gun, updates))

	// all the other data is still there, but are no longer the current ones
	assertExpectedTUFMetaInStore(t, s, append(firstBatch, secondBatch...), false)
	assertExpectedTUFMetaInStore(t, s, thirdBatch[:1], true)
	assertExpectedTUFMetaInStore(t, s, thirdBatch[1:], false)

	return append(append(firstBatch, secondBatch...), thirdBatch...)
}

// UpdateMany does not insert any rows (or at least rolls them back) if there
// are any conflicts.
func testUpdateManyConflictRollback(t *testing.T, s MetaStore) []StoredTUFMeta {
	var gun data.GUN = "testGUN"
	successBatch := make([]StoredTUFMeta, 4)
	updates := make([]MetaUpdate, 4)
	for i, role := range data.BaseRoles {
		successBatch[i] = SampleCustomTUFObj(gun, role, 1, nil)
		updates[i] = MakeUpdate(successBatch[i])
	}

	require.NoError(t, s.UpdateMany(gun, updates))

	before, err := s.GetChanges("0", 1000, "")
	if _, ok := s.(RethinkDB); !ok {
		require.NoError(t, err)
	}

	// conflicts with what's in DB
	badBatch := make([]StoredTUFMeta, 4)
	for i, role := range data.BaseRoles {
		version := 2
		if role == data.CanonicalTargetsRole {
			version = 1
		}
		tufdata := []byte(fmt.Sprintf("%s_%s_%d_bad", gun, role, version))
		badBatch[i] = SampleCustomTUFObj(gun, role, version, tufdata)
		updates[i] = MakeUpdate(badBatch[i])
	}

	// check no changes were written when there was a conflict+rollback
	after, err := s.GetChanges("0", 1000, "")
	if _, ok := s.(RethinkDB); !ok {
		require.NoError(t, err)
	}
	require.Equal(t, len(before), len(after))

	err = s.UpdateMany(gun, updates)
	require.Error(t, err)
	require.IsType(t, ErrOldVersion{}, err)

	// self-conflicting, in that it's a duplicate, but otherwise no DB conflicts
	duplicate := SampleCustomTUFObj(gun, data.CanonicalTimestampRole, 3, []byte("duplicate"))
	duplicateUpdate := MakeUpdate(duplicate)
	err = s.UpdateMany(gun, []MetaUpdate{duplicateUpdate, duplicateUpdate})
	require.Error(t, err)
	require.IsType(t, ErrOldVersion{}, err)

	assertExpectedTUFMetaInStore(t, s, successBatch, true)

	for _, tufObj := range append(badBatch, duplicate) {
		checksumBytes := sha256.Sum256(tufObj.Data)
		checksum := hex.EncodeToString(checksumBytes[:])

		_, _, err = s.GetChecksum(tufObj.Gun, tufObj.Role, checksum)
		require.Error(t, err)
		require.IsType(t, ErrNotFound{}, err)
	}

	return successBatch
}

// Delete will remove all TUF metadata, all versions, associated with a gun
func testDeleteSuccess(t *testing.T, s MetaStore) {
	var gun data.GUN = "testGUN"
	// If there is nothing in the DB, delete is a no-op success
	require.NoError(t, s.Delete(gun))

	// If there is data in the DB, all versions are deleted
	unexpected := make([]StoredTUFMeta, 0, 10)
	updates := make([]MetaUpdate, 0, 10)
	for version := 1; version < 3; version++ {
		for _, role := range append(data.BaseRoles, "targets/a") {
			tufObj := SampleCustomTUFObj(gun, role, version, nil)
			unexpected = append(unexpected, tufObj)
			updates = append(updates, MakeUpdate(tufObj))
		}
	}
	require.NoError(t, s.UpdateMany(gun, updates))
	assertExpectedTUFMetaInStore(t, s, unexpected[:5], false)
	assertExpectedTUFMetaInStore(t, s, unexpected[5:], true)

	require.NoError(t, s.Delete(gun))

	for _, tufObj := range unexpected {
		_, _, err := s.GetCurrent(tufObj.Gun, tufObj.Role)
		require.IsType(t, ErrNotFound{}, err)

		checksumBytes := sha256.Sum256(tufObj.Data)
		checksum := hex.EncodeToString(checksumBytes[:])

		_, _, err = s.GetChecksum(tufObj.Gun, tufObj.Role, checksum)
		require.Error(t, err)
		require.IsType(t, ErrNotFound{}, err)
	}

	// We can now write the same files without conflicts to the DB
	require.NoError(t, s.UpdateMany(gun, updates))
	assertExpectedTUFMetaInStore(t, s, unexpected[:5], false)
	assertExpectedTUFMetaInStore(t, s, unexpected[5:], true)

	// And delete them again successfully
	require.NoError(t, s.Delete(gun))
}

func testGetChanges(t *testing.T, s MetaStore) {
	// non-int changeID
	c, err := s.GetChanges("foo", 10, "")
	require.Error(t, err)
	require.Len(t, c, 0)

	// add some records
	require.NoError(t, s.UpdateMany("alpine", []MetaUpdate{
		{
			Role:    data.CanonicalTimestampRole,
			Version: 1,
			Data:    []byte{'1'},
		},
		{
			Role:    data.CanonicalTimestampRole,
			Version: 2,
			Data:    []byte{'2'},
		},
		{
			Role:    data.CanonicalTimestampRole,
			Version: 3,
			Data:    []byte{'3'},
		},
		{
			Role:    data.CanonicalTimestampRole,
			Version: 4,
			Data:    []byte{'4'},
		},
	}))
	require.NoError(t, s.UpdateMany("busybox", []MetaUpdate{
		{
			Role:    data.CanonicalTimestampRole,
			Version: 1,
			Data:    []byte{'5'},
		},
		{
			Role:    data.CanonicalTimestampRole,
			Version: 2,
			Data:    []byte{'6'},
		},
		{
			Role:    data.CanonicalTimestampRole,
			Version: 3,
			Data:    []byte{'7'},
		},
		{
			Role:    data.CanonicalTimestampRole,
			Version: 4,
			Data:    []byte{'8'},
		},
	}))

	// check non-error cases
	c, err = s.GetChanges("0", 8, "")
	require.NoError(t, err)
	require.Len(t, c, 8)

	for i := 0; i < 4; i++ {
		require.Equal(t, uint(i+1), c[i].ID)
		require.Equal(t, "alpine", c[i].GUN)
		require.Equal(t, i+1, c[i].Version)
	}
	for i := 4; i < 8; i++ {
		require.Equal(t, uint(i+1), c[i].ID)
		require.Equal(t, "busybox", c[i].GUN)
		require.Equal(t, i-3, c[i].Version)
	}

	c, err = s.GetChanges("-1", 4, "")
	require.NoError(t, err)
	require.Len(t, c, 4)
	for i := 0; i < 4; i++ {
		require.Equal(t, uint(i+5), c[i].ID)
		require.Equal(t, "busybox", c[i].GUN)
		require.Equal(t, i+1, c[i].Version)
	}

	c, err = s.GetChanges("10", 4, "")
	require.NoError(t, err)
	require.Len(t, c, 0)

	c, err = s.GetChanges("10", -4, "")
	require.NoError(t, err)
	require.Len(t, c, 4)
	for i := 0; i < 4; i++ {
		require.Equal(t, uint(i+5), c[i].ID)
		require.Equal(t, "busybox", c[i].GUN)
		require.Equal(t, i+1, c[i].Version)
	}

	c, err = s.GetChanges("7", -4, "")
	require.NoError(t, err)
	require.Len(t, c, 4)
	for i := 0; i < 2; i++ {
		require.Equal(t, uint(i+3), c[i].ID)
		require.Equal(t, "alpine", c[i].GUN)
		require.Equal(t, i+3, c[i].Version)
	}
	for i := 2; i < 4; i++ {
		require.Equal(t, uint(i+3), c[i].ID)
		require.Equal(t, "busybox", c[i].GUN)
		require.Equal(t, i-1, c[i].Version)
	}

	c, err = s.GetChanges("0", 8, "busybox")
	require.NoError(t, err)
	require.Len(t, c, 4)
	for i := 0; i < 4; i++ {
		require.Equal(t, uint(i+5), c[i].ID)
		require.Equal(t, "busybox", c[i].GUN)
		require.Equal(t, i+1, c[i].Version)
	}

	c, err = s.GetChanges("-1", -8, "busybox")
	require.NoError(t, err)
	require.Len(t, c, 4)
	for i := 0; i < 4; i++ {
		require.Equal(t, uint(i+5), c[i].ID)
		require.Equal(t, "busybox", c[i].GUN)
		require.Equal(t, i+1, c[i].Version)
	}

	// update a snapshot and confirm the most recent item of the changelist
	// hasn't changed (only timestamps should create changes)
	before, err := s.GetChanges("-1", -1, "")
	require.NoError(t, err)
	require.NoError(t, s.UpdateMany("alpine", []MetaUpdate{
		{
			Role:    data.CanonicalSnapshotRole,
			Version: 1,
			Data:    []byte{'1'},
		},
	}))
	after, err := s.GetChanges("-1", -1, "")
	require.NoError(t, err)
	require.Equal(t, before, after)

	// do a deletion and check is shows up.
	require.NoError(t, s.Delete("alpine"))
	c, err = s.GetChanges("-1", -1, "")
	require.NoError(t, err)
	require.Len(t, c, 1)
	require.Equal(t, changeCategoryDeletion, c[0].Category)
	require.Equal(t, "alpine", c[0].GUN)

	// do another deletion and check it doesn't show up because no records were deleted
	// after the first one
	require.NoError(t, s.Delete("alpine"))
	c, err = s.GetChanges("-1", -2, "")
	require.NoError(t, err)
	require.Len(t, c, 2)
	require.NotEqual(t, changeCategoryDeletion, c[0].Category)
	require.NotEqual(t, "alpine", c[0].GUN)
}
