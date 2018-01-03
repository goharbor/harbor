package remoteks

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"crypto/tls"
	"crypto/x509"
	"github.com/docker/notary/storage"
	"github.com/docker/notary/trustmanager"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

type TestError struct{}

func (err TestError) Error() string {
	return "test error"
}

type ErroringStorage struct{}

func (s ErroringStorage) Set(string, []byte) error {
	return TestError{}
}

func (s ErroringStorage) Remove(string) error {
	return TestError{}
}

func (s ErroringStorage) Get(string) ([]byte, error) {
	return nil, TestError{}
}

func (s ErroringStorage) ListFiles() []string {
	return nil
}

func (s ErroringStorage) Location() string {
	return "erroringstorage"
}

func getCertsDir(t *testing.T) string {
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(file)
	certsDir := filepath.Join(dir, "../../fixtures/")
	return certsDir
}

func getServerTLS(t *testing.T) *tls.Config {
	certDir := getCertsDir(t)
	cert, err := tls.LoadX509KeyPair(
		filepath.Join(certDir, "notary-escrow.crt"),
		filepath.Join(certDir, "notary-escrow.key"),
	)
	require.NoError(t, err)
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
}

func getClientTLS(t *testing.T) *tls.Config {
	certDir := getCertsDir(t)
	pool := x509.NewCertPool()
	cert, err := ioutil.ReadFile(filepath.Join(certDir, "root-ca.crt"))
	require.NoError(t, err)
	pool.AppendCertsFromPEM(
		cert,
	)
	return &tls.Config{
		RootCAs: pool,
	}
}

func setupTestServer(t *testing.T, addr string, store trustmanager.Storage) func() {
	s := grpc.NewServer(
		grpc.Creds(
			credentials.NewTLS(
				getServerTLS(t),
			),
		),
	)
	st := NewGRPCStorage(store)
	l, err := net.Listen(
		"tcp",
		addr,
	)
	require.NoError(t, err)
	RegisterStoreServer(s, st)
	go func() {
		err := s.Serve(l)
		t.Logf("server errored %s", err)
	}()
	return func() {
		s.Stop()
		l.Close()
	}
}

func TestRemoteStore(t *testing.T) {
	name := "testfile"
	bytes := []byte{'1'}
	addr := "localhost:9888"

	closer := setupTestServer(t, addr, storage.NewMemoryStore(nil))
	defer closer()

	c, err := NewRemoteStore(addr, getClientTLS(t), 0)
	require.NoError(t, err)

	loc := c.Location()
	require.Equal(t, "Remote Key Store @ "+addr, loc)

	err = c.Set(name, bytes)
	require.NoError(t, err)

	out, err := c.Get(name)
	require.NoError(t, err)
	require.Equal(t, bytes, out)

	ls := c.ListFiles()
	require.Len(t, ls, 1)
	require.Equal(t, name, ls[0])

	err = c.Remove(name)
	require.NoError(t, err)

	ls = c.ListFiles()
	require.Len(t, ls, 0)

	_, err = c.Get(name)
	require.Error(t, err)
}

// GRPC converts our errors into *grpc.rpcError types.
func TestErrors(t *testing.T) {
	name := "testfile"
	bytes := []byte{'1'}
	addr := "localhost:9887"

	closer := setupTestServer(t, addr, ErroringStorage{})
	defer closer()

	c, err := NewRemoteStore(addr, getClientTLS(t), 0)
	require.NoError(t, err)

	err = c.Set(name, bytes)
	require.Error(t, err)
	require.Equal(t, "test error", grpc.ErrorDesc(err))

	_, err = c.Get(name)
	require.Error(t, err)
	require.Equal(t, "test error", grpc.ErrorDesc(err))

	err = c.Remove(name)
	require.Error(t, err)
	require.Equal(t, "test error", grpc.ErrorDesc(err))
}
