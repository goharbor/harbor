package storage

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRDBTUFFileJSONUnmarshalling(t *testing.T) {
	created := time.Now().AddDate(-1, -1, -1)
	updated := time.Now().AddDate(0, -5, 0)
	deleted := time.Time{}
	data := []byte("Hello world")

	createdMarshalled, err := json.Marshal(created)
	require.NoError(t, err)
	updatedMarshalled, err := json.Marshal(updated)
	require.NoError(t, err)
	deletedMarshalled, err := json.Marshal(deleted)
	require.NoError(t, err)
	dataMarshalled, err := json.Marshal(data)
	require.NoError(t, err)

	jsonBytes := []byte(fmt.Sprintf(`
	{
		"created_at": %s,
		"updated_at": %s,
		"deleted_at": %s,
		"gun_role_version": ["completely", "invalid", "garbage"],
		"gun": "namespaced/name",
		"role": "timestamp",
		"version": 5,
		"sha256": "56ee4a23129fc22c6cb4b4ba5f78d730c91ab6def514e80d807c947bb21f0d63",
		"data": %s,
		"timestamp_checksum": "ebe6b6e082c94ef24043f1786a7046432506c3d193a47c299ed48ff4413ad7b0"
	}
	`, createdMarshalled, updatedMarshalled, deletedMarshalled, dataMarshalled))

	unmarshalledAnon, err := TUFFilesRethinkTable.JSONUnmarshaller(jsonBytes)
	require.NoError(t, err)
	unmarshalled, ok := unmarshalledAnon.(RDBTUFFile)
	require.True(t, ok)

	// There is some weirdness with comparing time.Time due to a location pointer,
	// so let's use time.Time's equal function to compare times, and then re-assign
	// the timing struct to compare the rest of the RDBTUFFile struct
	require.True(t, created.Equal(unmarshalled.CreatedAt))
	require.True(t, updated.Equal(unmarshalled.UpdatedAt))
	require.True(t, deleted.Equal(unmarshalled.DeletedAt))

	expected := RDBTUFFile{
		Timing:         unmarshalled.Timing,
		GunRoleVersion: []interface{}{"namespaced/name", "timestamp", 5},
		Gun:            "namespaced/name",
		Role:           "timestamp",
		Version:        5,
		SHA256:         "56ee4a23129fc22c6cb4b4ba5f78d730c91ab6def514e80d807c947bb21f0d63",
		Data:           data,
		TSchecksum:     "ebe6b6e082c94ef24043f1786a7046432506c3d193a47c299ed48ff4413ad7b0",
	}
	require.Equal(t, expected, unmarshalled)
}

func TestRDBTUFFileJSONUnmarshallingFailure(t *testing.T) {
	validTimeMarshalled, err := json.Marshal(time.Now())
	require.NoError(t, err)
	dataMarshalled, err := json.Marshal([]byte("Hello world!"))
	require.NoError(t, err)

	invalids := []string{
		fmt.Sprintf(`
			{
				"created_at": "not a time",
				"updated_at": %s,
				"deleted_at": %s,
				"gun_role_version": ["completely", "invalid", "garbage"],
				"gun": "namespaced/name",
				"role": "timestamp",
				"version": 5,
				"sha256": "56ee4a23129fc22c6cb4b4ba5f78d730c91ab6def514e80d807c947bb21f0d63",
				"data": %s,
				"timestamp_checksum": "ebe6b6e082c94ef24043f1786a7046432506c3d193a47c299ed48ff4413ad7b0"
			}`, validTimeMarshalled, validTimeMarshalled, dataMarshalled),
		fmt.Sprintf(`
			{
				"created_at": %s,
				"updated_at": %s,
				"deleted_at": %s,
				"gun_role_version": ["completely", "invalid", "garbage"],
				"gun": "namespaced/name",
				"role": "timestamp",
				"version": 5,
				"sha256": "56ee4a23129fc22c6cb4b4ba5f78d730c91ab6def514e80d807c947bb21f0d63",
				"data": 1245,
				"timestamp_checksum": "ebe6b6e082c94ef24043f1786a7046432506c3d193a47c299ed48ff4413ad7b0"
			}`, validTimeMarshalled, validTimeMarshalled, validTimeMarshalled),
		fmt.Sprintf(`
			{
				"created_at": %s,
				"updated_at": %s,
				"deleted_at": %s,
				"gun_role_version": ["completely", "invalid", "garbage"],
				"gun": "namespaced/name",
				"role": "timestamp",
				"version": "not an int",
				"sha256": "56ee4a23129fc22c6cb4b4ba5f78d730c91ab6def514e80d807c947bb21f0d63",
				"data": %s,
				"timestamp_checksum": "ebe6b6e082c94ef24043f1786a7046432506c3d193a47c299ed48ff4413ad7b0"
			}`, validTimeMarshalled, validTimeMarshalled, validTimeMarshalled, dataMarshalled),
	}

	for _, invalid := range invalids {
		_, err := TUFFilesRethinkTable.JSONUnmarshaller([]byte(invalid))
		require.Error(t, err)
	}
}
