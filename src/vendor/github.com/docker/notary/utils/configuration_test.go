package utils

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/bugsnag/bugsnag-go"
	"github.com/docker/notary"
	"github.com/docker/notary/tuf/utils"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

const envPrefix = "NOTARY_TESTING_ENV_PREFIX"

const (
	Cert = "../fixtures/notary-server.crt"
	Key  = "../fixtures/notary-server.key"
	Root = "../fixtures/root-ca.crt"
)

// initializes a viper object with test configuration
func configure(jsonConfig string) *viper.Viper {
	config := viper.New()
	SetupViper(config, envPrefix)
	config.SetConfigType("json")
	config.ReadConfig(bytes.NewBuffer([]byte(jsonConfig)))
	return config
}

// Sets the environment variables in the given map, prefixed by envPrefix.
func setupEnvironmentVariables(t *testing.T, vars map[string]string) {
	for k, v := range vars {
		err := os.Setenv(fmt.Sprintf("%s_%s", envPrefix, k), v)
		require.NoError(t, err)
	}
}

// Unsets whatever environment variables were set with this map
func cleanupEnvironmentVariables(t *testing.T, vars map[string]string) {
	for k := range vars {
		err := os.Unsetenv(fmt.Sprintf("%s_%s", envPrefix, k))
		require.NoError(t, err)
	}

}

// An error is returned if the log level is not parsable
func TestParseInvalidLogLevel(t *testing.T) {
	_, err := ParseLogLevel(configure(`{"logging": {"level": "horatio"}}`),
		logrus.DebugLevel)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a valid logrus Level")
}

// If there is no logging level configured it is set to the default level
func TestParseNoLogLevel(t *testing.T) {
	empties := []string{`{}`, `{"logging": {}}`}
	for _, configJSON := range empties {
		lvl, err := ParseLogLevel(configure(configJSON), logrus.DebugLevel)
		require.NoError(t, err)
		require.Equal(t, logrus.DebugLevel, lvl)
	}
}

// If there is logging level configured, it is set to the configured one
func TestParseLogLevel(t *testing.T) {
	lvl, err := ParseLogLevel(configure(`{"logging": {"level": "error"}}`),
		logrus.DebugLevel)
	require.NoError(t, err)
	require.Equal(t, logrus.ErrorLevel, lvl)
}

func TestParseLogLevelWithEnvironmentVariables(t *testing.T) {
	vars := map[string]string{"LOGGING_LEVEL": "error"}
	setupEnvironmentVariables(t, vars)
	defer cleanupEnvironmentVariables(t, vars)

	lvl, err := ParseLogLevel(configure(`{}`),
		logrus.DebugLevel)
	require.NoError(t, err)
	require.Equal(t, logrus.ErrorLevel, lvl)
}

// An error is returned if there's no API key
func TestParseInvalidBugsnag(t *testing.T) {
	_, err := ParseBugsnag(configure(
		`{"reporting": {"bugsnag": {"endpoint": "http://12345"}}}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "must provide an API key")
}

// If there's no bugsnag, a nil pointer is returned
func TestParseNoBugsnag(t *testing.T) {
	empties := []string{`{}`, `{"reporting": {}}`}
	for _, configJSON := range empties {
		bugconf, err := ParseBugsnag(configure(configJSON))
		require.NoError(t, err)
		require.Nil(t, bugconf)
	}
}

func TestParseBugsnag(t *testing.T) {
	config := configure(`{
		"reporting": {
			"bugsnag": {
				"api_key": "12345",
				"release_stage": "production",
				"endpoint": "http://1234.com"
			}
		}
	}`)

	expected := bugsnag.Configuration{
		APIKey:       "12345",
		ReleaseStage: "production",
		Endpoint:     "http://1234.com",
	}

	bugconf, err := ParseBugsnag(config)
	require.NoError(t, err)
	require.Equal(t, expected, *bugconf)
}

func TestParseBugsnagWithEnvironmentVariables(t *testing.T) {
	config := configure(`{
		"reporting": {
			"bugsnag": {
				"api_key": "12345",
				"release_stage": "staging"
			}
		}
	}`)

	vars := map[string]string{
		"REPORTING_BUGSNAG_RELEASE_STAGE": "production",
		"REPORTING_BUGSNAG_ENDPOINT":      "http://1234.com",
	}
	setupEnvironmentVariables(t, vars)
	defer cleanupEnvironmentVariables(t, vars)

	expected := bugsnag.Configuration{
		APIKey:       "12345",
		ReleaseStage: "production",
		Endpoint:     "http://1234.com",
	}

	bugconf, err := ParseBugsnag(config)
	require.NoError(t, err)
	require.Equal(t, expected, *bugconf)
}

// If the storage backend is invalid or not provided, an error is returned.
func TestParseInvalidStorageBackend(t *testing.T) {
	invalids := []string{
		`{"storage": {"backend": "etcd", "db_url": "1234"}}`,
		`{"storage": {"db_url": "12345"}}`,
		`{"storage": {}}`,
		`{}`,
	}
	for _, configJSON := range invalids {
		_, err := ParseSQLStorage(configure(configJSON))
		require.Error(t, err, fmt.Sprintf("'%s' should be an error", configJSON))
		require.Contains(t, err.Error(),
			"is not a supported SQL backend driver")
	}
}

// If there is no DB url for non-memory backends, an error is returned.
func TestParseInvalidSQLStorageNoDBSource(t *testing.T) {
	invalids := []string{
		`{"storage": {"backend": "%s"}}`,
		`{"storage": {"backend": "%s", "db_url": ""}}`,
	}
	for _, backend := range []string{notary.MySQLBackend, notary.SQLiteBackend, notary.PostgresBackend} {
		for _, configJSONFmt := range invalids {
			configJSON := fmt.Sprintf(configJSONFmt, backend)
			_, err := ParseSQLStorage(configure(configJSON))
			require.Error(t, err, fmt.Sprintf("'%s' should be an error", configJSON))
			require.Contains(t, err.Error(),
				fmt.Sprintf("must provide a non-empty database source for %s", backend))
		}
	}
}

// A supported backend with DB source will be successfully parsed.
func TestParseSQLStorageDBStore(t *testing.T) {
	config := configure(`{
		"storage": {
			"backend": "mysql",
			"db_url": "username:passord@tcp(hostname:1234)/dbname"
		}
	}`)

	expected := Storage{
		Backend: "mysql",
		Source:  "username:passord@tcp(hostname:1234)/dbname",
	}

	store, err := ParseSQLStorage(config)
	require.NoError(t, err)
	require.Equal(t, expected, *store)
}

// ParseRethinkDBStorage will reject non rethink databases
func TestParseRethinkStorageDBStoreInvalidBackend(t *testing.T) {
	config := configure(`{
		"storage": {
			"backend": "mysql",
			"db_url": "username:password@tcp(hostname:1234)/dbname",
			"tls_ca_file": "/tls/ca.pem",
			"client_cert_file": "/tls/cert.pem",
			"client_key_file": "/tls/key.pem",
			"database": "rethinkdbtest",
			"username": "user"
		}
	}`)

	_, err := ParseRethinkDBStorage(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "not a supported RethinkDB backend")
}

// ParseRethinkDBStorage will require a db_url for rethink databases
func TestParseRethinkStorageDBStoreEmptyDBUrl(t *testing.T) {
	config := configure(`{
		"storage": {
			"backend": "rethinkdb",
			"tls_ca_file": "/tls/ca.pem",
			"client_cert_file": "/tls/cert.pem",
			"client_key_file": "/tls/key.pem",
			"database": "rethinkdbtest",
			"username": "user",
			"password": "password"
		}
	}`)

	_, err := ParseRethinkDBStorage(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must provide a non-empty host:port")
}

// ParseRethinkDBStorage will require a dbname for rethink databases
func TestParseRethinkStorageDBStoreEmptyDBName(t *testing.T) {
	config := configure(`{
		"storage": {
			"backend": "rethinkdb",
			"db_url": "username:password@tcp(hostname:1234)/dbname",
			"tls_ca_file": "/tls/ca.pem",
			"client_cert_file": "/tls/cert.pem",
			"client_key_file": "/tls/key.pem",
			"username": "user"
		}
	}`)

	_, err := ParseRethinkDBStorage(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a specific database to connect to")
}

// ParseRethinkDBStorage will require a CA cert for rethink databases
func TestParseRethinkStorageDBStoreEmptyCA(t *testing.T) {
	config := configure(`{
		"storage": {
			"backend": "rethinkdb",
			"db_url": "username:password@tcp(hostname:1234)/dbname",
			"database": "rethinkdbtest",
			"client_cert_file": "/tls/cert.pem",
			"client_key_file": "/tls/key.pem",
			"username": "user"
		}
	}`)

	_, err := ParseRethinkDBStorage(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cowardly refusal to connect to rethinkdb without a CA cert")
}

// ParseRethinkDBStorage will require a client cert and key to connect to rethink databases
func TestParseRethinkStorageDBStoreEmptyCertAndKey(t *testing.T) {
	config := configure(`{
		"storage": {
			"backend": "rethinkdb",
			"db_url": "username:password@tcp(hostname:1234)/dbname",
			"database": "rethinkdbtest",
			"tls_ca_file": "/tls/ca.pem",
			"username": "user"
		}
	}`)

	_, err := ParseRethinkDBStorage(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cowardly refusal to connect to rethinkdb without a client cert")
}

// ParseRethinkDBStorage will require a username to connect to the database after bootstrapping
func TestParseRethinkStorageDBStoreEmptyUsername(t *testing.T) {
	config := configure(`{
		"storage": {
			"backend": "rethinkdb",
			"db_url": "username:password@tcp(hostname:1234)/dbname",
			"database": "rethinkdbtest",
			"client_cert_file": "/tls/cert.pem",
			"client_key_file": "/tls/key.pem",
			"tls_ca_file": "/tls/ca.pem"
		}
	}`)

	_, err := ParseRethinkDBStorage(config)
	require.Error(t, err)
	require.Contains(t, err.Error(), "requires a username to connect to the db")
}

func TestParseSQLStorageWithEnvironmentVariables(t *testing.T) {
	config := configure(`{
		"storage": {
			"db_url": "username:passord@tcp(hostname:1234)/dbname"
		}
	}`)

	vars := map[string]string{"STORAGE_BACKEND": "mysql"}
	setupEnvironmentVariables(t, vars)
	defer cleanupEnvironmentVariables(t, vars)

	expected := Storage{
		Backend: "mysql",
		Source:  "username:passord@tcp(hostname:1234)/dbname",
	}

	store, err := ParseSQLStorage(config)
	require.NoError(t, err)
	require.Equal(t, expected, *store)
}

// If TLS is required and the parameters are missing, an error is returned
func TestParseTLSNoTLSWhenRequired(t *testing.T) {
	invalids := []string{
		fmt.Sprintf(`{"server": {"tls_cert_file": "%s"}}`, Cert),
		fmt.Sprintf(`{"server": {"tls_key_file": "%s"}}`, Key),
	}
	for _, configJSON := range invalids {
		_, err := ParseServerTLS(configure(configJSON), true)
		require.Error(t, err)
		require.Contains(t, err.Error(), "no such file or directory")
	}
}

// If TLS is not required and the cert/key are partially provided, an error is returned
func TestParseTLSPartialTLS(t *testing.T) {
	invalids := []string{
		fmt.Sprintf(`{"server": {"tls_cert_file": "%s"}}`, Cert),
		fmt.Sprintf(`{"server": {"tls_key_file": "%s"}}`, Key),
	}
	for _, configJSON := range invalids {
		_, err := ParseServerTLS(configure(configJSON), false)
		require.Error(t, err)
		require.Contains(t, err.Error(),
			"either include both a cert and key file, or no TLS information at all to disable TLS")
	}
}

func TestParseTLSNoTLSNotRequired(t *testing.T) {
	config := configure(`{
		"server": {}
	}`)

	tlsConfig, err := ParseServerTLS(config, false)
	require.NoError(t, err)
	require.Nil(t, tlsConfig)
}

func TestParseTLSWithTLS(t *testing.T) {
	config := configure(fmt.Sprintf(`{
		"server": {
			"tls_cert_file": "%s",
			"tls_key_file": "%s",
			"client_ca_file": "%s"
		}
	}`, Cert, Key, Root))

	tlsConfig, err := ParseServerTLS(config, false)
	require.NoError(t, err)

	expectedCert, err := tls.LoadX509KeyPair(Cert, Key)
	require.NoError(t, err)

	expectedRoot, err := utils.LoadCertFromFile(Root)
	require.NoError(t, err)

	require.Len(t, tlsConfig.Certificates, 1)
	require.True(t, reflect.DeepEqual(expectedCert, tlsConfig.Certificates[0]))

	subjects := tlsConfig.ClientCAs.Subjects()
	require.Len(t, subjects, 1)
	require.True(t, bytes.Equal(expectedRoot.RawSubject, subjects[0]))
	require.Equal(t, tlsConfig.ClientAuth, tls.RequireAndVerifyClientCert)
}

func TestParseTLSWithTLSRelativeToConfigFile(t *testing.T) {
	currDir, err := os.Getwd()
	require.NoError(t, err)

	config := configure(fmt.Sprintf(`{
		"server": {
			"tls_cert_file": "%s",
			"tls_key_file": "%s",
			"client_ca_file": ""
		}
	}`, Cert, filepath.Clean(filepath.Join(currDir, Key))))
	config.SetConfigFile(filepath.Join(currDir, "me.json"))

	tlsConfig, err := ParseServerTLS(config, false)
	require.NoError(t, err)

	expectedCert, err := tls.LoadX509KeyPair(Cert, Key)
	require.NoError(t, err)

	require.Len(t, tlsConfig.Certificates, 1)
	require.True(t, reflect.DeepEqual(expectedCert, tlsConfig.Certificates[0]))

	require.Nil(t, tlsConfig.ClientCAs)
	require.Equal(t, tlsConfig.ClientAuth, tls.NoClientCert)
}

func TestParseTLSWithEnvironmentVariables(t *testing.T) {
	config := configure(fmt.Sprintf(`{
		"server": {
			"tls_cert_file": "%s",
			"client_ca_file": "nosuchfile"
		}
	}`, Cert))

	vars := map[string]string{
		"SERVER_TLS_KEY_FILE":   Key,
		"SERVER_CLIENT_CA_FILE": Root,
	}
	setupEnvironmentVariables(t, vars)
	defer cleanupEnvironmentVariables(t, vars)

	tlsConfig, err := ParseServerTLS(config, true)
	require.NoError(t, err)

	expectedCert, err := tls.LoadX509KeyPair(Cert, Key)
	require.NoError(t, err)

	expectedRoot, err := utils.LoadCertFromFile(Root)
	require.NoError(t, err)

	require.Len(t, tlsConfig.Certificates, 1)
	require.True(t, reflect.DeepEqual(expectedCert, tlsConfig.Certificates[0]))

	subjects := tlsConfig.ClientCAs.Subjects()
	require.Len(t, subjects, 1)
	require.True(t, bytes.Equal(expectedRoot.RawSubject, subjects[0]))
	require.Equal(t, tlsConfig.ClientAuth, tls.RequireAndVerifyClientCert)
}

func TestParseViperWithInvalidFile(t *testing.T) {
	v := viper.New()
	SetupViper(v, envPrefix)

	err := ParseViper(v, "Chronicle_Of_Dark_Secrets.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "Could not read config")
}

func TestParseViperWithValidFile(t *testing.T) {
	testDir, err := ioutil.TempDir("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(testDir)

	file, err := os.Create(filepath.Join(testDir, "Chronicle_Of_Dark_Secrets.json"))
	require.NoError(t, err)

	file.WriteString(`{"logging": {"level": "debug"}}`)

	v := viper.New()
	SetupViper(v, envPrefix)

	err = ParseViper(v, file.Name())
	require.NoError(t, err)

	require.Equal(t, "debug", v.GetString("logging.level"))
}

type logLevelTests struct {
	startLevel logrus.Level
	endLevel   logrus.Level
	increment  bool
}

const (
	optIncrement = true
	optDecrement = false
)

var logLevelExpectations = []logLevelTests{
	// highest: Debug, lowest: Panic.  Incrementing brings everything up one level, except debug which is max level
	{startLevel: logrus.DebugLevel, increment: optIncrement, endLevel: logrus.DebugLevel},
	{startLevel: logrus.InfoLevel, increment: optIncrement, endLevel: logrus.DebugLevel},
	{startLevel: logrus.WarnLevel, increment: optIncrement, endLevel: logrus.InfoLevel},
	{startLevel: logrus.ErrorLevel, increment: optIncrement, endLevel: logrus.WarnLevel},
	{startLevel: logrus.FatalLevel, increment: optIncrement, endLevel: logrus.ErrorLevel},
	{startLevel: logrus.PanicLevel, increment: optIncrement, endLevel: logrus.FatalLevel},

	// highest: Debug, lowest: Panic.  Decrementing brings everything down one level, except panic which is min level
	{startLevel: logrus.DebugLevel, increment: optDecrement, endLevel: logrus.InfoLevel},
	{startLevel: logrus.InfoLevel, increment: optDecrement, endLevel: logrus.WarnLevel},
	{startLevel: logrus.WarnLevel, increment: optDecrement, endLevel: logrus.ErrorLevel},
	{startLevel: logrus.ErrorLevel, increment: optDecrement, endLevel: logrus.FatalLevel},
	{startLevel: logrus.FatalLevel, increment: optDecrement, endLevel: logrus.PanicLevel},
	{startLevel: logrus.PanicLevel, increment: optDecrement, endLevel: logrus.PanicLevel},
}

func TestAdjustLogLevel(t *testing.T) {
	for _, expt := range logLevelExpectations {
		logrus.SetLevel(expt.startLevel)
		err := AdjustLogLevel(expt.increment)

		if expt.startLevel == expt.endLevel {
			require.Error(t, err) // because if it didn't change, that means AdjustLogLevel failed
		} else {
			require.NoError(t, err)
		}

		require.Equal(t, expt.endLevel, logrus.GetLevel())
	}
}

func testSetSignalTrap(t *testing.T) {
	var signalsPassedOn map[string]struct{}

	signalHandler := func(s os.Signal) {
		signalsPassedOn := make(map[string]struct{})
		signalsPassedOn[s.String()] = struct{}{}
	}
	c := SetupSignalTrap(signalHandler)

	if len(notary.NotarySupportedSignals) == 0 { // currently, windows only
		require.Nil(t, c)
	} else {
		require.NotNil(t, c)
		defer signal.Stop(c)
	}

	for _, s := range notary.NotarySupportedSignals {
		syscallSignal, ok := s.(syscall.Signal)
		require.True(t, ok)
		require.NoError(t, syscall.Kill(syscall.Getpid(), syscallSignal))
		require.Len(t, signalsPassedOn, 0)
		require.NotNil(t, signalsPassedOn[s.String()])
	}
}

// TODO: undo this extra indirection, needed for mocking notary.NotarySupportedSignals being empty, when we have
// a windows CI system running
func TestSetSignalTrap(t *testing.T) {
	testSetSignalTrap(t)
}

func TestSetSignalTrapMockWindows(t *testing.T) {
	old := notary.NotarySupportedSignals
	notary.NotarySupportedSignals = nil
	testSetSignalTrap(t)
	notary.NotarySupportedSignals = old
}
