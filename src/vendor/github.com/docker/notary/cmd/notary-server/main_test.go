package main

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/docker/distribution/health"
	"github.com/docker/notary"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/signer/client"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/utils"
	_ "github.com/mattn/go-sqlite3"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const (
	Cert = "../../fixtures/notary-server.crt"
	Key  = "../../fixtures/notary-server.key"
	Root = "../../fixtures/root-ca.crt"
)

// initializes a viper object with test configuration
func configure(jsonConfig string) *viper.Viper {
	config := viper.New()
	config.SetConfigType("json")
	config.ReadConfig(bytes.NewBuffer([]byte(jsonConfig)))
	return config
}

func TestGetAddrAndTLSConfigInvalidTLS(t *testing.T) {
	invalids := []string{
		`{"server": {
				"http_addr": ":1234",
				"tls_key_file": "nope"
		}}`,
	}
	for _, configJSON := range invalids {
		_, _, err := getAddrAndTLSConfig(configure(configJSON))
		require.Error(t, err)
	}
}

func TestGetAddrAndTLSConfigNoHTTPAddr(t *testing.T) {
	_, _, err := getAddrAndTLSConfig(configure(fmt.Sprintf(`{
		"server": {
			"tls_cert_file": "%s",
			"tls_key_file": "%s"
		}
	}`, Cert, Key)))
	require.Error(t, err)
	require.Contains(t, err.Error(), "http listen address required for server")
}

func TestGetAddrAndTLSConfigSuccessWithTLS(t *testing.T) {
	httpAddr, tlsConf, err := getAddrAndTLSConfig(configure(fmt.Sprintf(`{
		"server": {
			"http_addr": ":2345",
			"tls_cert_file": "%s",
			"tls_key_file": "%s"
		}
	}`, Cert, Key)))
	require.NoError(t, err)
	require.Equal(t, ":2345", httpAddr)
	require.NotNil(t, tlsConf)
}

func TestGetAddrAndTLSConfigSuccessWithoutTLS(t *testing.T) {
	httpAddr, tlsConf, err := getAddrAndTLSConfig(configure(
		`{"server": {"http_addr": ":2345"}}`))
	require.NoError(t, err)
	require.Equal(t, ":2345", httpAddr)
	require.Nil(t, tlsConf)
}

func TestGetAddrAndTLSConfigWithClientTLS(t *testing.T) {
	httpAddr, tlsConf, err := getAddrAndTLSConfig(configure(fmt.Sprintf(`{
		"server": {
			"http_addr": ":2345",
			"tls_cert_file": "%s",
			"tls_key_file": "%s",
			"client_ca_file": "%s"
		}
	}`, Cert, Key, Root)))
	require.NoError(t, err)
	require.Equal(t, ":2345", httpAddr)
	require.NotNil(t, tlsConf.ClientCAs)
}

func fakeRegisterer(callCount *int) healthRegister {
	return func(_ string, _ time.Duration, _ health.CheckFunc) {
		(*callCount)++
	}

}

// If neither "remote" nor "local" is passed for "trust_service.type", an
// error is returned.
func TestGetInvalidTrustService(t *testing.T) {
	invalids := []string{
		`{"trust_service": {"type": "bruhaha", "key_algorithm": "rsa"}}`,
		`{}`,
	}
	var registerCalled = 0

	for _, config := range invalids {
		_, _, err := getTrustService(configure(config),
			getNotarySigner, fakeRegisterer(&registerCalled))
		require.Error(t, err)
		require.Contains(t, err.Error(),
			"must specify either a \"local\" or \"remote\" type for trust_service")
	}
	// no health function ever registered
	require.Equal(t, 0, registerCalled)
}

// If a local trust service is specified, a local trust service will be used
// with an ED22519 algorithm no matter what algorithm was specified.  No health
// function is configured.
func TestGetLocalTrustService(t *testing.T) {
	localConfig := `{"trust_service": {"type": "local", "key_algorithm": "meh"}}`

	var registerCalled = 0

	trust, algo, err := getTrustService(configure(localConfig),
		getNotarySigner, fakeRegisterer(&registerCalled))
	require.NoError(t, err)
	require.IsType(t, &signed.Ed25519{}, trust)
	require.Equal(t, data.ED25519Key, algo)

	// no health function ever registered
	require.Equal(t, 0, registerCalled)
}

// Invalid key algorithms result in an error if a remote trust service was
// specified.
func TestGetTrustServiceInvalidKeyAlgorithm(t *testing.T) {
	configTemplate := `
	{
		"trust_service": {
			"type": "remote",
			"hostname": "blah",
			"port": "1234",
			"key_algorithm": "%s"
		}
	}`
	badKeyAlgos := []string{
		fmt.Sprintf(configTemplate, ""),
		fmt.Sprintf(configTemplate, data.ECDSAx509Key),
		fmt.Sprintf(configTemplate, "random"),
	}
	var registerCalled = 0

	for _, config := range badKeyAlgos {
		_, _, err := getTrustService(configure(config),
			getNotarySigner, fakeRegisterer(&registerCalled))
		require.Error(t, err)
		require.Contains(t, err.Error(), "invalid key algorithm")
	}
	// no health function ever registered
	require.Equal(t, 0, registerCalled)
}

// template to be used for testing TLS parsing with the trust service
var trustTLSConfigTemplate = `
	{
		"trust_service": {
			"type": "remote",
			"hostname": "notary-signer",
			"port": "1234",
			"key_algorithm": "ecdsa",
			%s
		}
	}`

// Client cert and Key either both have to be empty or both have to be
// provided.
func TestGetTrustServiceTLSMissingCertOrKey(t *testing.T) {
	configs := []string{
		fmt.Sprintf(`"tls_client_cert": "%s"`, Cert),
		fmt.Sprintf(`"tls_client_key": "%s"`, Key),
	}
	var registerCalled = 0

	for _, clientTLSConfig := range configs {
		jsonConfig := fmt.Sprintf(trustTLSConfigTemplate, clientTLSConfig)
		config := configure(jsonConfig)
		_, _, err := getTrustService(config, getNotarySigner,
			fakeRegisterer(&registerCalled))
		require.Error(t, err)
		require.True(t,
			strings.Contains(err.Error(), "either pass both client key and cert, or neither"))
	}
	// no health function ever registered
	require.Equal(t, 0, registerCalled)
}

// If no TLS configuration is provided for the host server, no TLS config will
// be set for the trust service.
func TestGetTrustServiceNoTLSConfig(t *testing.T) {
	config := `{
		"trust_service": {
			"type": "remote",
			"hostname": "notary-signer",
			"port": "1234",
			"key_algorithm": "ecdsa"
		}
	}`
	var registerCalled = 0

	var tlsConfig *tls.Config
	var fakeNewSigner = func(_, _ string, c *tls.Config) (*client.NotarySigner, error) {
		tlsConfig = c
		return &client.NotarySigner{}, nil
	}

	trust, algo, err := getTrustService(configure(config),
		fakeNewSigner, fakeRegisterer(&registerCalled))
	require.NoError(t, err)
	require.IsType(t, &client.NotarySigner{}, trust)
	require.Equal(t, "ecdsa", algo)
	require.Nil(t, tlsConfig.RootCAs)
	require.Nil(t, tlsConfig.Certificates)
	// health function registered
	require.Equal(t, 1, registerCalled)
}

// The rest of the functionality of getTrustService depends upon
// utils.ConfigureClientTLS, so this test just asserts that if successful,
// the correct tls.Config is returned based on all the configuration parameters
func TestGetTrustServiceTLSSuccess(t *testing.T) {
	keypair, err := tls.LoadX509KeyPair(Cert, Key)
	require.NoError(t, err, "Unable to load cert and key for testing")

	tlspart := fmt.Sprintf(`"tls_client_cert": "%s", "tls_client_key": "%s"`,
		Cert, Key)

	var registerCalled = 0

	var tlsConfig *tls.Config
	var fakeNewSigner = func(_, _ string, c *tls.Config) (*client.NotarySigner, error) {
		tlsConfig = c
		return &client.NotarySigner{}, nil
	}

	trust, algo, err := getTrustService(
		configure(fmt.Sprintf(trustTLSConfigTemplate, tlspart)),
		fakeNewSigner, fakeRegisterer(&registerCalled))
	require.NoError(t, err)
	require.IsType(t, &client.NotarySigner{}, trust)
	require.Equal(t, "ecdsa", algo)
	require.Len(t, tlsConfig.Certificates, 1)
	require.True(t, reflect.DeepEqual(keypair, tlsConfig.Certificates[0]))
	// health function registered
	require.Equal(t, 1, registerCalled)
}

// The rest of the functionality of getTrustService depends upon
// utils.ConfigureServerTLS, so this test just asserts that if it fails,
// the error is propagated.
func TestGetTrustServiceTLSFailure(t *testing.T) {
	tlspart := fmt.Sprintf(`"tls_client_cert": "none", "tls_client_key": "%s"`,
		Key)

	var registerCalled = 0

	_, _, err := getTrustService(
		configure(fmt.Sprintf(trustTLSConfigTemplate, tlspart)),
		getNotarySigner, fakeRegisterer(&registerCalled))

	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(),
		"Unable to configure TLS to the trust service"))

	// no health function ever registered
	require.Equal(t, 0, registerCalled)
}

// Just to ensure that errors are propagated
func TestGetStoreInvalid(t *testing.T) {
	config := `{"storage": {"backend": "asdf", "db_url": "does_not_matter_what_value_this_is"}}`

	var registerCalled = 0

	_, err := getStore(configure(config), fakeRegisterer(&registerCalled), false)
	require.Error(t, err)

	// no health function ever registered
	require.Equal(t, 0, registerCalled)
}

func TestGetStoreDBStore(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "sqlite3")
	require.NoError(t, err)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	config := fmt.Sprintf(`{"storage": {"backend": "%s", "db_url": "%s"}}`,
		notary.SQLiteBackend, tmpFile.Name())

	var registerCalled = 0

	store, err := getStore(configure(config), fakeRegisterer(&registerCalled), false)
	require.NoError(t, err)
	_, ok := store.(storage.TUFMetaStorage)
	require.True(t, ok)

	// health function registered
	require.Equal(t, 1, registerCalled)
}

func TestGetStoreRethinkDBStoreConnectionFails(t *testing.T) {
	config := fmt.Sprintf(
		`{"storage": {
			"backend": "%s",
			"db_url": "host:port",
			"tls_ca_file": "/tls/ca.pem",
			"client_cert_file": "/tls/cert.pem",
			"client_key_file": "/tls/key.pem",
			"database": "rethinkdbtest"
			}
		}`,
		notary.RethinkDBBackend)

	var registerCalled = 0

	_, err := getStore(configure(config), fakeRegisterer(&registerCalled), false)
	require.Error(t, err)
}

func TestGetMemoryStore(t *testing.T) {
	var registerCalled = 0

	config := fmt.Sprintf(`{"storage": {"backend": "%s"}}`, notary.MemoryBackend)
	store, err := getStore(configure(config), fakeRegisterer(&registerCalled), false)
	require.NoError(t, err)
	_, ok := store.(*storage.MemStorage)
	require.True(t, ok)

	// no health function ever registered
	require.Equal(t, 0, registerCalled)
}

func TestGetCacheConfig(t *testing.T) {
	defaults := `{}`
	valid := `{"caching": {"max_age": {"current_metadata": 0, "consistent_metadata": 31536000}}}`
	invalids := []string{
		`{"caching": {"max_age": {"current_metadata": 0, "consistent_metadata": 31539000}}}`,
		`{"caching": {"max_age": {"current_metadata": -1, "consistent_metadata": 300}}}`,
		`{"caching": {"max_age": {"current_metadata": "hello", "consistent_metadata": 300}}}`,
	}

	current, consistent, err := getCacheConfig(configure(defaults))
	require.NoError(t, err)
	require.Equal(t,
		utils.PublicCacheControl{MaxAgeInSeconds: int(notary.CurrentMetadataCacheMaxAge.Seconds()),
			MustReValidate: true}, current)
	require.Equal(t,
		utils.PublicCacheControl{MaxAgeInSeconds: int(notary.ConsistentMetadataCacheMaxAge.Seconds())}, consistent)

	current, consistent, err = getCacheConfig(configure(valid))
	require.NoError(t, err)
	require.Equal(t, utils.NoCacheControl{}, current)
	require.Equal(t, utils.PublicCacheControl{MaxAgeInSeconds: 31536000}, consistent)

	for _, invalid := range invalids {
		_, _, err := getCacheConfig(configure(invalid))
		require.Error(t, err)
	}
}

func TestGetGUNPRefixes(t *testing.T) {
	valids := map[string][]string{
		`{}`: nil,
		`{"repositories": {"gun_prefixes": []}}`:         nil,
		`{"repositories": {}}`:                           nil,
		`{"repositories": {"gun_prefixes": ["hello/"]}}`: {"hello/"},
	}
	invalids := []string{
		`{"repositories": {"gun_prefixes": " / "}}`,
		`{"repositories": {"gun_prefixes": "nope"}}`,
		`{"repositories": {"gun_prefixes": ["nope"]}}`,
		`{"repositories": {"gun_prefixes": ["/nope/"]}}`,
		`{"repositories": {"gun_prefixes": ["../nope/"]}}`,
	}

	for valid, expected := range valids {
		prefixes, err := getRequiredGunPrefixes(configure(valid))
		require.NoError(t, err)
		require.Equal(t, expected, prefixes)
	}
	for _, invalid := range invalids {
		_, err := getRequiredGunPrefixes(configure(invalid))
		require.Error(t, err, "expected error with %s", invalid)
	}
}

// For sanity, make sure we can always parse the sample config
func TestSampleConfig(t *testing.T) {
	var registerCalled = 0
	_, _, err := parseServerConfig("../../fixtures/server-config.json", fakeRegisterer(&registerCalled), false)
	require.NoError(t, err)

	// once for the DB, once for the trust service
	require.Equal(t, registerCalled, 2)
}
