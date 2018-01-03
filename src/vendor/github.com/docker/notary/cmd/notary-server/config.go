package main

import (
	"crypto/tls"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/docker/distribution/health"
	_ "github.com/docker/distribution/registry/auth/htpasswd"
	_ "github.com/docker/distribution/registry/auth/token"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/notary"
	"github.com/docker/notary/server"
	"github.com/docker/notary/server/storage"
	"github.com/docker/notary/signer/client"
	"github.com/docker/notary/storage/rethinkdb"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	"github.com/docker/notary/utils"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"gopkg.in/dancannon/gorethink.v3"
)

// gets the required gun prefixes accepted by this server
func getRequiredGunPrefixes(configuration *viper.Viper) ([]string, error) {
	prefixes := configuration.GetStringSlice("repositories.gun_prefixes")
	for _, prefix := range prefixes {
		p := path.Clean(strings.TrimSpace(prefix))
		if p+"/" != prefix || strings.HasPrefix(p, "/") || strings.HasPrefix(p, "..") {
			return nil, fmt.Errorf("invalid GUN prefix %s", prefix)
		}
	}
	return prefixes, nil
}

// get the address for the HTTP server, and parses the optional TLS
// configuration for the server - if no TLS configuration is specified,
// TLS is not enabled.
func getAddrAndTLSConfig(configuration *viper.Viper) (string, *tls.Config, error) {
	httpAddr := configuration.GetString("server.http_addr")
	if httpAddr == "" {
		return "", nil, fmt.Errorf("http listen address required for server")
	}

	tlsConfig, err := utils.ParseServerTLS(configuration, false)
	if err != nil {
		return "", nil, fmt.Errorf(err.Error())
	}
	return httpAddr, tlsConfig, nil
}

// sets up TLS for the GRPC connection to notary-signer
func grpcTLS(configuration *viper.Viper) (*tls.Config, error) {
	rootCA := utils.GetPathRelativeToConfig(configuration, "trust_service.tls_ca_file")
	clientCert := utils.GetPathRelativeToConfig(configuration, "trust_service.tls_client_cert")
	clientKey := utils.GetPathRelativeToConfig(configuration, "trust_service.tls_client_key")

	if clientCert == "" && clientKey != "" || clientCert != "" && clientKey == "" {
		return nil, fmt.Errorf("either pass both client key and cert, or neither")
	}

	tlsConfig, err := tlsconfig.Client(tlsconfig.Options{
		CAFile:   rootCA,
		CertFile: clientCert,
		KeyFile:  clientKey,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"Unable to configure TLS to the trust service: %s", err.Error())
	}
	return tlsConfig, nil
}

// parses the configuration and returns a backing store for the TUF files
func getStore(configuration *viper.Viper, hRegister healthRegister, doBootstrap bool) (
	storage.MetaStore, error) {
	var store storage.MetaStore
	backend := configuration.GetString("storage.backend")
	logrus.Infof("Using %s backend", backend)

	switch backend {
	case notary.MemoryBackend:
		return storage.NewMemStorage(), nil
	case notary.MySQLBackend, notary.SQLiteBackend, notary.PostgresBackend:
		storeConfig, err := utils.ParseSQLStorage(configuration)
		if err != nil {
			return nil, err
		}
		s, err := storage.NewSQLStorage(storeConfig.Backend, storeConfig.Source)
		if err != nil {
			return nil, fmt.Errorf("Error starting %s driver: %s", backend, err.Error())
		}
		store = *storage.NewTUFMetaStorage(s)
		hRegister("DB operational", 10*time.Second, s.CheckHealth)
	case notary.RethinkDBBackend:
		var sess *gorethink.Session
		storeConfig, err := utils.ParseRethinkDBStorage(configuration)
		if err != nil {
			return nil, err
		}
		tlsOpts := tlsconfig.Options{
			CAFile:   storeConfig.CA,
			CertFile: storeConfig.Cert,
			KeyFile:  storeConfig.Key,
		}
		if doBootstrap {
			sess, err = rethinkdb.AdminConnection(tlsOpts, storeConfig.Source)
		} else {
			sess, err = rethinkdb.UserConnection(tlsOpts, storeConfig.Source, storeConfig.Username, storeConfig.Password)
		}
		if err != nil {
			return nil, fmt.Errorf("Error starting %s driver: %s", backend, err.Error())
		}
		s := storage.NewRethinkDBStorage(storeConfig.DBName, storeConfig.Username, storeConfig.Password, sess)
		store = *storage.NewTUFMetaStorage(s)
		hRegister("DB operational", 10*time.Second, s.CheckHealth)
	default:
		return nil, fmt.Errorf("%s is not a supported storage backend", backend)
	}
	return store, nil
}

type signerFactory func(hostname, port string, tlsConfig *tls.Config) (*client.NotarySigner, error)
type healthRegister func(name string, duration time.Duration, check health.CheckFunc)

func getNotarySigner(hostname, port string, tlsConfig *tls.Config) (*client.NotarySigner, error) {
	conn, err := client.NewGRPCConnection(hostname, port, tlsConfig)
	if err != nil {
		return nil, err
	}
	return client.NewNotarySigner(conn), nil
}

// parses the configuration and determines which trust service and key algorithm
// to return
func getTrustService(configuration *viper.Viper, sFactory signerFactory,
	hRegister healthRegister) (signed.CryptoService, string, error) {

	switch configuration.GetString("trust_service.type") {
	case "local":
		logrus.Info("Using local signing service, which requires ED25519. " +
			"Ignoring all other trust_service parameters, including keyAlgorithm")
		return signed.NewEd25519(), data.ED25519Key, nil
	case "remote":
	default:
		return nil, "", fmt.Errorf(
			"must specify either a \"local\" or \"remote\" type for trust_service")
	}

	keyAlgo := configuration.GetString("trust_service.key_algorithm")
	if keyAlgo != data.ED25519Key && keyAlgo != data.ECDSAKey && keyAlgo != data.RSAKey {
		return nil, "", fmt.Errorf("invalid key algorithm configured: %s", keyAlgo)
	}

	clientTLS, err := grpcTLS(configuration)
	if err != nil {
		return nil, "", err
	}

	logrus.Info("Using remote signing service")

	notarySigner, err := sFactory(
		configuration.GetString("trust_service.hostname"),
		configuration.GetString("trust_service.port"),
		clientTLS,
	)

	if err != nil {
		return nil, "", err
	}

	duration := 10 * time.Second
	hRegister(
		"Trust operational",
		duration,
		func() error {
			err := notarySigner.CheckHealth(duration, notary.HealthCheckOverall)
			if err != nil {
				logrus.Error("Trust not fully operational: ", err.Error())
			}
			return err
		},
	)
	return notarySigner, keyAlgo, nil
}

// Parse the cache configurations for GET-ting current and checksummed metadata,
// returning the configuration for current (non-content-addressed) metadata
// first, then the configuration for consistent (content-addressed) metadata
// second. The configuration consists mainly of the max-age (an integer in seconds,
// just like in the Cache-Control header) for each type of metadata.
// The max-age must be between 0 and 31536000 (one year in seconds, which is
// the recommended maximum time data is cached), else parsing will return an error.
// A max-age of 0 will disable caching for that type of download (consistent or current).
func getCacheConfig(configuration *viper.Viper) (current, consistent utils.CacheControlConfig, err error) {
	cccs := make(map[string]utils.CacheControlConfig)
	currentOpt, consistentOpt := "current_metadata", "consistent_metadata"

	defaults := map[string]int{
		currentOpt:    int(notary.CurrentMetadataCacheMaxAge.Seconds()),
		consistentOpt: int(notary.ConsistentMetadataCacheMaxAge.Seconds()),
	}
	maxMaxAge := int(notary.CacheMaxAgeLimit.Seconds())

	for optionName, seconds := range defaults {
		m := configuration.GetString(fmt.Sprintf("caching.max_age.%s", optionName))
		if m != "" {
			seconds, err = strconv.Atoi(m)
			if err != nil || seconds < 0 || seconds > maxMaxAge {
				return nil, nil, fmt.Errorf(
					"must specify a cache-control max-age between 0 and %v", maxMaxAge)
			}
		}
		cccs[optionName] = utils.NewCacheControlConfig(seconds, optionName == currentOpt)
	}
	current = cccs[currentOpt]
	consistent = cccs[consistentOpt]
	return
}

func parseServerConfig(configFilePath string, hRegister healthRegister, doBootstrap bool) (context.Context, server.Config, error) {
	config := viper.New()
	utils.SetupViper(config, envPrefix)

	// parse viper config
	if err := utils.ParseViper(config, configFilePath); err != nil {
		return nil, server.Config{}, err
	}

	ctx := context.Background()

	// default is error level
	lvl, err := utils.ParseLogLevel(config, logrus.ErrorLevel)
	if err != nil {
		return nil, server.Config{}, err
	}
	logrus.SetLevel(lvl)

	prefixes, err := getRequiredGunPrefixes(config)
	if err != nil {
		return nil, server.Config{}, err
	}

	// parse bugsnag config
	bugsnagConf, err := utils.ParseBugsnag(config)
	if err != nil {
		return ctx, server.Config{}, err
	}
	utils.SetUpBugsnag(bugsnagConf)

	trust, keyAlgo, err := getTrustService(config, getNotarySigner, hRegister)
	if err != nil {
		return nil, server.Config{}, err
	}
	ctx = context.WithValue(ctx, notary.CtxKeyKeyAlgo, keyAlgo)

	store, err := getStore(config, hRegister, doBootstrap)
	if err != nil {
		return nil, server.Config{}, err
	}
	ctx = context.WithValue(ctx, notary.CtxKeyMetaStore, store)

	currentCache, consistentCache, err := getCacheConfig(config)
	if err != nil {
		return nil, server.Config{}, err
	}

	httpAddr, tlsConfig, err := getAddrAndTLSConfig(config)
	if err != nil {
		return nil, server.Config{}, err
	}

	return ctx, server.Config{
		Addr:                         httpAddr,
		TLSConfig:                    tlsConfig,
		Trust:                        trust,
		AuthMethod:                   config.GetString("auth.type"),
		AuthOpts:                     config.Get("auth.options"),
		RepoPrefixes:                 prefixes,
		CurrentCacheControlConfig:    currentCache,
		ConsistentCacheControlConfig: consistentCache,
	}, nil
}
