// Common configuration elements that may be resused

package utils

import (
	"crypto/tls"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	bugsnag_hook "github.com/Shopify/logrus-bugsnag"
	"github.com/Sirupsen/logrus"
	"github.com/bugsnag/bugsnag-go"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/spf13/viper"

	"github.com/docker/notary"
)

// Storage is a configuration about what storage backend a server should use
type Storage struct {
	Backend string
	Source  string
}

// RethinkDBStorage is configuration about a RethinkDB backend service
type RethinkDBStorage struct {
	Storage
	CA       string
	Cert     string
	DBName   string
	Key      string
	Username string
	Password string
}

// GetPathRelativeToConfig gets a configuration key which is a path, and if
// it is not empty or an absolute path, returns the absolute path relative
// to the configuration file
func GetPathRelativeToConfig(configuration *viper.Viper, key string) string {
	configFile := configuration.ConfigFileUsed()
	p := configuration.GetString(key)
	if p == "" || filepath.IsAbs(p) {
		return p
	}
	return filepath.Clean(filepath.Join(filepath.Dir(configFile), p))
}

// ParseServerTLS tries to parse out valid server TLS options from a Viper.
// The cert/key files are relative to the config file used to populate the instance
// of viper.
func ParseServerTLS(configuration *viper.Viper, tlsRequired bool) (*tls.Config, error) {
	//  unmarshalling into objects does not seem to pick up env vars
	tlsOpts := tlsconfig.Options{
		CertFile: GetPathRelativeToConfig(configuration, "server.tls_cert_file"),
		KeyFile:  GetPathRelativeToConfig(configuration, "server.tls_key_file"),
		CAFile:   GetPathRelativeToConfig(configuration, "server.client_ca_file"),
	}
	if tlsOpts.CAFile != "" {
		tlsOpts.ClientAuth = tls.RequireAndVerifyClientCert
	}

	if !tlsRequired {
		cert, key, ca := tlsOpts.CertFile, tlsOpts.KeyFile, tlsOpts.CAFile
		if cert == "" && key == "" && ca == "" {
			return nil, nil
		}

		if (cert == "" && key != "") || (cert != "" && key == "") || (cert == "" && key == "" && ca != "") {
			return nil, fmt.Errorf(
				"either include both a cert and key file, or no TLS information at all to disable TLS")
		}
	}

	return tlsconfig.Server(tlsOpts)
}

// ParseLogLevel tries to parse out a log level from a Viper.  If there is no
// configuration, defaults to the provided error level
func ParseLogLevel(configuration *viper.Viper, defaultLevel logrus.Level) (
	logrus.Level, error) {

	logStr := configuration.GetString("logging.level")
	if logStr == "" {
		return defaultLevel, nil
	}
	return logrus.ParseLevel(logStr)
}

// ParseSQLStorage tries to parse out Storage from a Viper.  If backend and
// URL are not provided, returns a nil pointer.  Storage is required (if
// a backend is not provided, an error will be returned.)
func ParseSQLStorage(configuration *viper.Viper) (*Storage, error) {
	store := Storage{
		Backend: configuration.GetString("storage.backend"),
		Source:  configuration.GetString("storage.db_url"),
	}

	switch {
	case store.Backend != notary.MySQLBackend && store.Backend != notary.SQLiteBackend && store.Backend != notary.PostgresBackend:
		return nil, fmt.Errorf(
			"%s is not a supported SQL backend driver",
			store.Backend,
		)
	case store.Source == "":
		return nil, fmt.Errorf(
			"must provide a non-empty database source for %s",
			store.Backend,
		)
	}
	return &store, nil
}

// ParseRethinkDBStorage tries to parse out Storage from a Viper.  If backend and
// URL are not provided, returns a nil pointer.  Storage is required (if
// a backend is not provided, an error will be returned.)
func ParseRethinkDBStorage(configuration *viper.Viper) (*RethinkDBStorage, error) {
	store := RethinkDBStorage{
		Storage: Storage{
			Backend: configuration.GetString("storage.backend"),
			Source:  configuration.GetString("storage.db_url"),
		},
		CA:       GetPathRelativeToConfig(configuration, "storage.tls_ca_file"),
		Cert:     GetPathRelativeToConfig(configuration, "storage.client_cert_file"),
		Key:      GetPathRelativeToConfig(configuration, "storage.client_key_file"),
		DBName:   configuration.GetString("storage.database"),
		Username: configuration.GetString("storage.username"),
		Password: configuration.GetString("storage.password"),
	}

	switch {
	case store.Backend != notary.RethinkDBBackend:
		return nil, fmt.Errorf(
			"%s is not a supported RethinkDB backend driver",
			store.Backend,
		)
	case store.Source == "":
		return nil, fmt.Errorf(
			"must provide a non-empty host:port for %s",
			store.Backend,
		)
	case store.CA == "":
		return nil, fmt.Errorf(
			"cowardly refusal to connect to %s without a CA cert",
			store.Backend,
		)
	case store.Cert == "" || store.Key == "":
		return nil, fmt.Errorf(
			"cowardly refusal to connect to %s without a client cert and key",
			store.Backend,
		)
	case store.DBName == "":
		return nil, fmt.Errorf(
			"%s requires a specific database to connect to",
			store.Backend,
		)
	case store.Username == "":
		return nil, fmt.Errorf(
			"%s requires a username to connect to the db",
			store.Backend,
		)
	}

	return &store, nil
}

// ParseBugsnag tries to parse out a Bugsnag Configuration from a Viper.
// If no values are provided, returns a nil pointer.
func ParseBugsnag(configuration *viper.Viper) (*bugsnag.Configuration, error) {
	// can't unmarshal because we can't add tags to the bugsnag.Configuration
	// struct
	bugconf := bugsnag.Configuration{
		APIKey:       configuration.GetString("reporting.bugsnag.api_key"),
		ReleaseStage: configuration.GetString("reporting.bugsnag.release_stage"),
		Endpoint:     configuration.GetString("reporting.bugsnag.endpoint"),
	}
	if bugconf.APIKey == "" && bugconf.ReleaseStage == "" && bugconf.Endpoint == "" {
		return nil, nil
	}
	if bugconf.APIKey == "" {
		return nil, fmt.Errorf("must provide an API key for bugsnag")
	}
	return &bugconf, nil
}

// utilities for setting up/acting on common configurations

// SetupViper sets up an instance of viper to also look at environment
// variables
func SetupViper(v *viper.Viper, envPrefix string) {
	v.SetEnvPrefix(envPrefix)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
}

// SetUpBugsnag configures bugsnag and sets up a logrus hook
func SetUpBugsnag(config *bugsnag.Configuration) error {
	if config != nil {
		bugsnag.Configure(*config)
		hook, err := bugsnag_hook.NewBugsnagHook()
		if err != nil {
			return err
		}
		logrus.AddHook(hook)
		logrus.Debug("Adding logrus hook for Bugsnag")
	}
	return nil
}

// ParseViper tries to parse out a Viper from a configuration file.
func ParseViper(v *viper.Viper, configFile string) error {
	filename := filepath.Base(configFile)
	ext := filepath.Ext(configFile)
	configPath := filepath.Dir(configFile)

	v.SetConfigType(strings.TrimPrefix(ext, "."))
	v.SetConfigName(strings.TrimSuffix(filename, ext))
	v.AddConfigPath(configPath)

	if err := v.ReadInConfig(); err != nil {
		return fmt.Errorf("Could not read config at :%s, viper error: %v", configFile, err)
	}
	return nil
}

// AdjustLogLevel increases/decreases the log level, return error if the operation is invaild.
func AdjustLogLevel(increment bool) error {
	lvl := logrus.GetLevel()

	// The log level seems not possible, in the foreseeable future,
	// out of range [Panic, Debug]
	if increment {
		if lvl == logrus.DebugLevel {
			return fmt.Errorf("log level can not be set higher than %s", "Debug")
		}
		lvl++
	} else {
		if lvl == logrus.PanicLevel {
			return fmt.Errorf("log level can not be set lower than %s", "Panic")
		}
		lvl--
	}

	logrus.SetLevel(lvl)
	return nil
}

// SetupSignalTrap is a utility to trap supported signals hand handle them (currently by increasing logging)
func SetupSignalTrap(handler func(os.Signal)) chan os.Signal {
	if len(notary.NotarySupportedSignals) == 0 {
		return nil

	}
	c := make(chan os.Signal, 1)
	signal.Notify(c, notary.NotarySupportedSignals...)
	go func() {
		for {
			handler(<-c)
		}
	}()

	return c
}
