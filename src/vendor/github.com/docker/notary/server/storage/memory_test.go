// +build !mysqldb,!rethinkdb,!postgresqldb

package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func assertExpectedMemoryTUFMeta(t *testing.T, expected []StoredTUFMeta, s *MemStorage) {
	for _, tufObj := range expected {
		k := entryKey(tufObj.Gun, tufObj.Role)
		versionList, ok := s.tufMeta[k]
		require.True(t, ok, "Did not find this gun+role in store")
		byVersion := make(map[int]ver)
		for _, v := range versionList {
			byVersion[v.version] = v
		}

		v, ok := byVersion[tufObj.Version]
		require.True(t, ok, "Did not find version %d in store", tufObj.Version)
		require.Equal(t, tufObj.Data, v.data, "Data was incorrect")
	}
}

// UpdateCurrent should succeed if there was no previous metadata of the same
// gun and role.  They should be gettable.
func TestMemoryUpdateCurrentEmpty(t *testing.T) {
	s := NewMemStorage()
	expected := testUpdateCurrentEmptyStore(t, s)
	assertExpectedMemoryTUFMeta(t, expected, s)
}

// UpdateCurrent will successfully add a new (higher) version of an existing TUF file,
// but will return an error if the to-be-added version already exists in the DB.
func TestMemoryUpdateCurrentVersionCheckOldVersionExists(t *testing.T) {
	s := NewMemStorage()
	expected := testUpdateCurrentVersionCheck(t, s, true)
	assertExpectedMemoryTUFMeta(t, expected, s)
}

// UpdateCurrent will successfully add a new (higher) version of an existing TUF file,
// but will return an error if the to-be-added version does not exist in the DB, but
// is older than an existing version in the DB.
func TestMemoryUpdateCurrentVersionCheckOldVersionNotExist(t *testing.T) {
	s := NewMemStorage()
	expected := testUpdateCurrentVersionCheck(t, s, false)
	assertExpectedMemoryTUFMeta(t, expected, s)
}

// UpdateMany succeeds if the updates do not conflict with each other or with what's
// already in the DB
func TestMemoryUpdateManyNoConflicts(t *testing.T) {
	s := NewMemStorage()
	expected := testUpdateManyNoConflicts(t, s)
	assertExpectedMemoryTUFMeta(t, expected, s)
}

// UpdateMany does not insert any rows (or at least rolls them back) if there
// are any conflicts.
func TestMemoryUpdateManyConflictRollback(t *testing.T) {
	s := NewMemStorage()
	expected := testUpdateManyConflictRollback(t, s)
	assertExpectedMemoryTUFMeta(t, expected, s)
}

// Delete will remove all TUF metadata, all versions, associated with a gun
func TestMemoryDeleteSuccess(t *testing.T) {
	s := NewMemStorage()
	testDeleteSuccess(t, s)
	assertExpectedMemoryTUFMeta(t, nil, s)
}

func TestGetCurrent(t *testing.T) {
	s := NewMemStorage()

	_, _, err := s.GetCurrent("gun", "role")
	require.IsType(t, ErrNotFound{}, err, "Expected error to be ErrNotFound")

	s.UpdateCurrent("gun", MetaUpdate{"role", 1, []byte("test")})
	_, d, err := s.GetCurrent("gun", "role")
	require.Nil(t, err, "Expected error to be nil")
	require.Equal(t, []byte("test"), d, "Data was incorrect")
}

func TestGetChecksumNotFound(t *testing.T) {
	s := NewMemStorage()
	_, _, err := s.GetChecksum("gun", "root", "12345")
	require.Error(t, err)
	require.IsType(t, ErrNotFound{}, err)
}

func TestMemoryGetChanges(t *testing.T) {
	s := NewMemStorage()

	testGetChanges(t, s)
}

func TestGetVersion(t *testing.T) {
	s := NewMemStorage()
	testGetVersion(t, s)
}
