package remoteks

import (
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"

	"github.com/docker/notary/storage"
)

func TestNewGRPCStorage(t *testing.T) {
	s := NewGRPCStorage(nil)
	require.IsType(t, &GRPCStorage{}, s)
}

// Set, Get, List, Remove, List, Get
func TestGRPCStorage(t *testing.T) {
	name := "testfile"
	bytes := []byte{'1'}
	ctx := context.Background()

	s := NewGRPCStorage(storage.NewMemoryStore(nil))
	msg := &SetMsg{
		FileName: name,
		Data:     bytes,
	}
	_, err := s.Set(ctx, msg)
	require.NoError(t, err)

	resp, err := s.Get(ctx, &FileNameMsg{FileName: name})
	require.NoError(t, err)
	require.Equal(t, bytes, resp.Data)

	ls, err := s.ListFiles(ctx, nil)
	require.NoError(t, err)
	require.Len(t, ls.FileNames, 1)
	require.Equal(t, name, ls.FileNames[0])

	_, err = s.Remove(ctx, &FileNameMsg{FileName: name})
	require.NoError(t, err)

	ls, err = s.ListFiles(ctx, nil)
	require.NoError(t, err)
	require.Len(t, ls.FileNames, 0)

	_, err = s.Get(ctx, &FileNameMsg{FileName: name})
	require.Error(t, err)
}
