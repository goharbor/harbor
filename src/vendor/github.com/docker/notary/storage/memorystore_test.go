package storage

import (
	"crypto/sha256"
	"testing"

	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/utils"
	"github.com/stretchr/testify/require"
)

func TestMemoryStoreMetadataOperations(t *testing.T) {
	s := NewMemoryStore(nil)

	// GetSized of a non-existent metadata fails
	_, err := s.GetSized("nonexistent", 0)
	require.Error(t, err)
	require.IsType(t, ErrMetaNotFound{}, err)

	// Once SetMeta succeeds, GetSized with the role name and the consistent name
	// should succeed
	metaContent := []byte("content")
	metaSize := int64(len(metaContent))
	shasum := sha256.Sum256(metaContent)
	invalidShasum := sha256.Sum256([]byte{})

	require.NoError(t, s.Set("exists", metaContent))
	require.NoError(t, s.SetMulti(map[string][]byte{"multi1": metaContent, "multi2": metaContent}))

	for _, metaName := range []string{"exists", "multi1", "multi2"} {
		role := data.RoleName(metaName)
		meta, err := s.GetSized(metaName, metaSize)
		require.NoError(t, err)
		require.Equal(t, metaContent, meta)

		meta, err = s.GetSized(utils.ConsistentName(role.String(), shasum[:]), metaSize)
		require.NoError(t, err)
		require.Equal(t, metaContent, meta)

		_, err = s.GetSized(utils.ConsistentName(role.String(), invalidShasum[:]), metaSize)
		require.Error(t, err)
		require.IsType(t, ErrMetaNotFound{}, err)
	}

	// Once Metadata is removed, it's no longer accessible
	err = s.RemoveAll()
	require.NoError(t, err)

	_, err = s.GetSized("exists", 0)
	require.Error(t, err)
	require.IsType(t, ErrMetaNotFound{}, err)
}

func TestMemoryStoreGetSized(t *testing.T) {
	content := []byte("content")
	s := NewMemoryStore(map[data.RoleName][]byte{"content": content})

	// we can get partial size
	meta, err := s.GetSized("content", 3)
	require.NoError(t, err)
	require.Equal(t, []byte("con"), meta)

	// we can get zero size
	meta, err = s.GetSized("content", 0)
	require.NoError(t, err)
	require.Equal(t, []byte{}, meta)

	// we can get the whole thing by passing NoSizeLimit (-1)
	meta, err = s.GetSized("content", NoSizeLimit)
	require.NoError(t, err)
	require.Equal(t, content, meta)

	// a size much larger than the actual length will return the whole thing
	meta, err = s.GetSized("content", 8000)
	require.NoError(t, err)
	require.Equal(t, content, meta)
}
