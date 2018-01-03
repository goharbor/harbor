package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/docker/notary"
	"github.com/docker/notary/signer"
	"github.com/docker/notary/signer/keydbstore"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/testutils"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	Cert = "../../fixtures/notary-signer.crt"
	Key  = "../../fixtures/notary-signer.key"
)

// initializes a viper object with test configuration
func configure(jsonConfig string) *viper.Viper {
	config := viper.New()
	config.SetConfigType("json")
	config.ReadConfig(bytes.NewBuffer([]byte(jsonConfig)))
	return config
}

// If the TLS configuration is invalid, an error is returned.  This doesn't test
// all the cases of the TLS configuration being invalid, since it's just
// calling configuration.ParseTLSConfig - this test just makes sure the
// error is propagated.
func TestGetAddrAndTLSConfigInvalidTLS(t *testing.T) {
	invalids := []string{
		`{"server": {"grpc_addr": ":2345"}}`,
		`{"server": {
				"grpc_addr": ":2345",
				"tls_cert_file": "nope",
				"tls_key_file": "nope"
		}}`,
	}
	for _, configJSON := range invalids {
		_, _, err := getAddrAndTLSConfig(configure(configJSON))
		require.Error(t, err)
		require.Contains(t, err.Error(), "unable to set up TLS")
	}
}

// If a GRPC address is not provided, an error is returned.
func TestGetAddrAndTLSConfigNoGRPCAddr(t *testing.T) {
	_, _, err := getAddrAndTLSConfig(configure(fmt.Sprintf(`{
		"server": {
			"tls_cert_file": "%s",
			"tls_key_file": "%s"
		}
	}`, Cert, Key)))
	require.Error(t, err)
	require.Contains(t, err.Error(), "grpc listen address required for server")
}

// Success parsing a valid TLS config, HTTP address, and GRPC address.
func TestGetAddrAndTLSConfigSuccess(t *testing.T) {
	grpcAddr, tlsConf, err := getAddrAndTLSConfig(configure(fmt.Sprintf(`{
		"server": {
			"grpc_addr": ":1234",
			"tls_cert_file": "%s",
			"tls_key_file": "%s"
		}
	}`, Cert, Key)))
	require.NoError(t, err)
	require.Equal(t, ":1234", grpcAddr)
	require.NotNil(t, tlsConf)
}

// If a default alias is not provided to a DB backend, an error is returned.
func TestSetupCryptoServicesDBStoreNoDefaultAlias(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "sqlite3")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err = setUpCryptoservices(
		configure(fmt.Sprintf(
			`{"storage": {"backend": "%s", "db_url": "%s"}}`,
			notary.SQLiteBackend, tmpFile.Name())),
		[]string{notary.SQLiteBackend},
		false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must provide a default alias for the key DB")
}

// If a default alias is not provided to a rethinkdb backend, an error is returned.
func TestSetupCryptoServicesRethinkDBStoreNoDefaultAlias(t *testing.T) {
	_, err := setUpCryptoservices(
		configure(fmt.Sprintf(
			`{"storage": {
				"backend": "%s",
				"db_url": "host:port",
				"tls_ca_file": "/tls/ca.pem",
				"client_cert_file": "/tls/cert.pem",
				"client_key_file": "/tls/key.pem",
				"database": "rethinkdbtest",
				"username": "signer",
				"password": "password"
				}
			}`,
			notary.RethinkDBBackend)),
		[]string{notary.RethinkDBBackend}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must provide a default alias for the key DB")
}

func TestSetupCryptoServicesRethinkDBStoreConnectionFails(t *testing.T) {
	// We don't have a rethink instance up, so the Connection() call will fail
	_, err := setUpCryptoservices(
		configure(fmt.Sprintf(
			`{"storage": {
				"backend": "%s",
				"db_url": "host:port",
				"tls_ca_file": "../../fixtures/rethinkdb/ca.pem",
				"client_cert_file": "../../fixtures/rethinkdb/cert.pem",
				"client_key_file": "../../fixtures/rethinkdb/key.pem",
				"database": "rethinkdbtest",
				"username": "signer",
				"password": "password"
				},
				"default_alias": "timestamp"
			}`,
			notary.RethinkDBBackend)),
		[]string{notary.RethinkDBBackend}, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no such host")
}

// If a default alias *is* provided to a valid DB backend, a valid
// CryptoService is returned.  (This depends on ParseStorage, which is tested
// separately, so this doesn't test all the possible cases of storage
// success/failure).
func TestSetupCryptoServicesDBStoreSuccess(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "sqlite3")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	// Ensure that the private_key table exists
	db, err := gorm.Open("sqlite3", tmpFile.Name())
	require.NoError(t, err)
	var (
		gormKey = keydbstore.GormPrivateKey{}
		count   int
	)
	db.CreateTable(&gormKey)
	db.Model(&gormKey).Count(&count)
	require.Equal(t, 0, count)

	cryptoServices, err := setUpCryptoservices(
		configure(fmt.Sprintf(
			`{"storage": {"backend": "%s", "db_url": "%s"},
			"default_alias": "timestamp"}`,
			notary.SQLiteBackend, tmpFile.Name())),
		[]string{notary.SQLiteBackend}, false)
	require.NoError(t, err)
	require.Len(t, cryptoServices, 2)

	edService, ok := cryptoServices[data.ED25519Key]
	require.True(t, ok)

	ecService, ok := cryptoServices[data.ECDSAKey]
	require.True(t, ok)

	require.Equal(t, edService, ecService)

	// since the keystores are not exposed by CryptoService, try creating
	// a key and seeing if it is in the sqlite DB.
	os.Setenv("NOTARY_SIGNER_TIMESTAMP", "password")
	defer os.Unsetenv("NOTARY_SIGNER_TIMESTAMP")

	_, err = ecService.Create("timestamp", "", data.ECDSAKey)
	require.NoError(t, err)
	db.Model(&gormKey).Count(&count)
	require.Equal(t, 1, count)
}

// If a memory backend is specified, then a default alias is not needed, and
// a valid CryptoService is returned.
func TestSetupCryptoServicesMemoryStore(t *testing.T) {
	config := configure(fmt.Sprintf(`{"storage": {"backend": "%s"}}`,
		notary.MemoryBackend))
	cryptoServices, err := setUpCryptoservices(config,
		[]string{notary.SQLiteBackend, notary.MemoryBackend}, false)
	require.NoError(t, err)
	require.Len(t, cryptoServices, 2)

	edService, ok := cryptoServices[data.ED25519Key]
	require.True(t, ok)

	ecService, ok := cryptoServices[data.ECDSAKey]
	require.True(t, ok)

	require.Equal(t, edService, ecService)

	// since the keystores are not exposed by CryptoService, try creating
	// and getting the key
	pubKey, err := ecService.Create("", "", data.ECDSAKey)
	require.NoError(t, err)
	privKey, _, err := ecService.GetPrivateKey(pubKey.ID())
	require.NoError(t, err)
	require.NotNil(t, privKey)
}

func TestSetupCryptoServicesInvalidStore(t *testing.T) {
	config := configure(fmt.Sprintf(`{"storage": {"backend": "%s"}}`,
		"invalid_backend"))
	_, err := setUpCryptoservices(config,
		[]string{notary.SQLiteBackend, notary.MemoryBackend, notary.RethinkDBBackend}, false)
	require.Error(t, err)
	require.Equal(t, err.Error(), fmt.Sprintf("%s is not an allowed backend, must be one of: %s", "invalid_backend", []string{notary.SQLiteBackend, notary.MemoryBackend, notary.RethinkDBBackend}))
}

func TestSetupGRPCServerInvalidAddress(t *testing.T) {
	_, _, err := setupGRPCServer(signer.Config{GRPCAddr: "nope", CryptoServices: make(signer.CryptoServiceIndex)})
	require.Error(t, err)
	require.Contains(t, err.Error(), "grpc server failed to listen on nope")
}

func TestSetupGRPCServerSuccess(t *testing.T) {
	tlsConf := tls.Config{InsecureSkipVerify: true}
	grpcServer, lis, err := setupGRPCServer(signer.Config{
		GRPCAddr:       ":7899",
		TLSConfig:      &tlsConf,
		CryptoServices: make(signer.CryptoServiceIndex),
	})
	require.NoError(t, err)
	defer lis.Close()
	require.Equal(t, "[::]:7899", lis.Addr().String())
	require.Equal(t, "tcp", lis.Addr().Network())
	require.NotNil(t, grpcServer)
}

func TestBootstrap(t *testing.T) {
	var ks trustmanager.KeyStore
	err := bootstrap(ks)
	require.Error(t, err)
	tb := &testutils.TestBootstrapper{}
	err = bootstrap(tb)
	require.NoError(t, err)
	require.True(t, tb.Booted)
}

func TestGetEnv(t *testing.T) {
	os.Setenv("NOTARY_SIGNER_TIMESTAMP", "password")
	defer os.Unsetenv("NOTARY_SIGNER_TIMESTAMP")

	require.Equal(t, "password", getEnv("timestamp"))
}

func TestPassphraseRetrieverInvalid(t *testing.T) {
	_, _, err := passphraseRetriever("fakeKey", "fakeAlias", false, 1)
	require.Error(t, err)
}

// For sanity, make sure we can always parse the sample config
func TestSampleConfig(t *testing.T) {
	// We need to provide a default alias for the key DB.
	//
	// Generally it will be done during the building process
	// if using signer.Dockerfile.
	os.Setenv("NOTARY_SIGNER_DEFAULT_ALIAS", "timestamp_1")
	defer os.Unsetenv("NOTARY_SIGNER_DEFAULT_ALIAS")
	_, err := parseSignerConfig("../../fixtures/signer-config-local.json", false)
	require.NoError(t, err)
}
