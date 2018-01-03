package main

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/docker/notary"
	"github.com/docker/notary/storage"
)

func TestParseConfigError(t *testing.T) {
	_, err := parseConfig("notafile")
	require.Error(t, err)
}

func TestSetupGRPCServer(t *testing.T) {
	v := viper.New()
	v.SetDefault("storage.backend", notary.MemoryBackend)
	v.SetDefault("server.insecure", true)
	s, err := setupGRPCServer(v)
	require.NoError(t, err)
	require.IsType(t, grpc.NewServer(), s)

	v = viper.New()
	v.SetDefault("storage.backend", "not recognized")
	_, err = setupGRPCServer(v)
	require.Error(t, err)
}

func TestSetupStorage(t *testing.T) {
	v := viper.New()
	v.SetDefault("storage.backend", notary.MemoryBackend)
	s, err := setupStorage(v)
	require.NoError(t, err)
	require.IsType(t, &storage.MemoryStore{}, s)

	v = viper.New()
	v.SetDefault("storage.backend", notary.FileBackend)
	s, err = setupStorage(v)
	require.NoError(t, err)
	require.IsType(t, &storage.FilesystemStore{}, s)

	v = viper.New()
	v.SetDefault("storage.backend", "not recognized")
	_, err = setupStorage(v)
	require.Error(t, err)
}

func TestSetupNetListener(t *testing.T) {
	v := viper.New()
	v.SetDefault("server.addr", "invalidaddr")
	_, err := setupNetListener(v)
	require.Error(t, err)

	v = viper.New()
	v.SetDefault("server.addr", "127.0.0.1:9999")
	l, err := setupNetListener(v)
	require.NoError(t, err)
	l.Close()
}
