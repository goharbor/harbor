package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type storeFactory func() MetadataStore

// Verifies that the metadata store can get and set metadata
func testGetSetMeta(t *testing.T, factory storeFactory) {
	s := factory()
	metaBytes, err := s.GetSized("root", 300)
	require.Error(t, err)
	require.Nil(t, metaBytes)
	require.IsType(t, ErrMetaNotFound{}, err)

	content := []byte("root bytes")
	require.NoError(t, s.Set("root", content))

	metaBytes, err = s.GetSized("root", 300)
	require.NoError(t, err)
	require.Equal(t, content, metaBytes)
}

// Verifies that the metadata store can delete metadata
func testRemove(t *testing.T, factory storeFactory) {
	s := factory()

	require.NoError(t, s.Set("root", []byte("test data")))

	require.NoError(t, s.Remove("root"))
	_, err := s.GetSized("root", 300)
	require.Error(t, err)
	require.IsType(t, ErrMetaNotFound{}, err)

	// delete metadata should be successful even if the metadata doesn't exist
	require.NoError(t, s.Remove("root"))
}

func TestMemoryStoreMetadata(t *testing.T) {
	factory := func() MetadataStore {
		return NewMemoryStore(nil)
	}

	testGetSetMeta(t, factory)
	testRemove(t, factory)
}
