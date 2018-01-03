package storage

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOfflineStore(t *testing.T) {
	s := OfflineStore{}
	_, err := s.GetSized("", 0)
	require.Error(t, err)
	require.IsType(t, ErrOffline{}, err)

	err = s.Set("", nil)
	require.Error(t, err)
	require.IsType(t, ErrOffline{}, err)

	err = s.SetMulti(nil)
	require.Error(t, err)
	require.IsType(t, ErrOffline{}, err)

	_, err = s.GetKey("")
	require.Error(t, err)
	require.IsType(t, ErrOffline{}, err)

	_, err = s.RotateKey("")
	require.Error(t, err)
	require.IsType(t, ErrOffline{}, err)

	err = s.RemoveAll()
	require.Error(t, err)
	require.IsType(t, ErrOffline{}, err)
}

func TestErrOffline(t *testing.T) {
	var _ error = ErrOffline{}
}
