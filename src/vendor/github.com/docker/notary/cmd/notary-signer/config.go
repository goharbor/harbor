package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/Sirupsen/logrus"
	"github.com/docker/distribution/health"
	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/notary"
	"github.com/docker/notary/cryptoservice"
	"github.com/docker/notary/passphrase"
	pb "github.com/docker/notary/proto"
	"github.com/docker/notary/signer"
	"github.com/docker/notary/signer/api"
	"github.com/docker/notary/signer/keydbstore"
	"github.com/docker/notary/storage"
	"github.com/docker/notary/storage/rethinkdb"
	"github.com/docker/notary/trustmanager"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/tuf/signed"
	tufutils "github.com/docker/notary/tuf/utils"
	"github.com/docker/notary/utils"
	"github.com/spf13/viper"
	ghealth "google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"gopkg.in/dancannon/gorethink.v3"
)

const (
	envPrefix       = "NOTARY_SIGNER"
	defaultAliasEnv = "DEFAULT_ALIAS"
)

func parseSignerConfig(configFilePath string, doBootstrap bool) (signer.Config, error) {
	config := viper.New()
	utils.SetupViper(config, envPrefix)

	// parse viper config
	if err := utils.ParseViper(config, configFilePath); err != nil {
		return signer.Config{}, err
	}

	// default is error level
	lvl, err := utils.ParseLogLevel(config, logrus.ErrorLevel)
	if err != nil {
		return signer.Config{}, err
	}
	logrus.SetLevel(lvl)

	// parse bugsnag config
	bugsnagConf, err := utils.ParseBugsnag(config)
	if err != nil {
		return signer.Config{}, err
	}
	utils.SetUpBugsnag(bugsnagConf)

	// parse server config
	grpcAddr, tlsConfig, err := getAddrAndTLSConfig(config)
	if err != nil {
		return signer.Config{}, err
	}

	// setup the cryptoservices
	cryptoServices, err := setUpCryptoservices(config, notary.NotarySupportedBackends, doBootstrap)
	if err != nil {
		return signer.Config{}, err
	}

	return signer.Config{
		GRPCAddr:       grpcAddr,
		TLSConfig:      tlsConfig,
		CryptoServices: cryptoServices,
	}, nil
}

func getEnv(env string) string {
	v := viper.New()
	utils.SetupViper(v, envPrefix)
	return v.GetString(strings.ToUpper(env))
}

func passphraseRetriever(keyName, alias string, createNew bool, attempts int) (passphrase string, giveup bool, err error) {
	passphrase = getEnv(alias)

	if passphrase == "" {
		return "", false, errors.New("expected env variable to not be empty: " + alias)
	}

	return passphrase, false, nil
}

// Reads the configuration file for storage setup, and sets up the cryptoservice
// mapping
func setUpCryptoservices(configuration *viper.Viper, allowedBackends []string, doBootstrap bool) (
	signer.CryptoServiceIndex, error) {
	backend := configuration.GetString("storage.backend")

	if !tufutils.StrSliceContains(allowedBackends, backend) {
		return nil, fmt.Errorf("%s is not an allowed backend, must be one of: %s", backend, allowedBackends)
	}

	var keyService signed.CryptoService
	switch backend {
	case notary.MemoryBackend:
		keyService = cryptoservice.NewCryptoService(trustmanager.NewKeyMemoryStore(
			passphrase.ConstantRetriever("memory-db-ignore")))
	case notary.RethinkDBBackend:
		var sess *gorethink.Session
		storeConfig, err := utils.ParseRethinkDBStorage(configuration)
		if err != nil {
			return nil, err
		}
		defaultAlias, err := getDefaultAlias(configuration)
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
		s := keydbstore.NewRethinkDBKeyStore(storeConfig.DBName, storeConfig.Username, storeConfig.Password, passphraseRetriever, defaultAlias, sess)
		health.RegisterPeriodicFunc("DB operational", time.Minute, s.CheckHealth)

		if doBootstrap {
			keyService = s
		} else {
			keyService = keydbstore.NewCachedKeyService(s)
		}
	case notary.MySQLBackend, notary.SQLiteBackend, notary.PostgresBackend:
		storeConfig, err := utils.ParseSQLStorage(configuration)
		if err != nil {
			return nil, err
		}
		defaultAlias, err := getDefaultAlias(configuration)
		if err != nil {
			return nil, err
		}
		dbStore, err := keydbstore.NewSQLKeyDBStore(
			passphraseRetriever, defaultAlias, storeConfig.Backend, storeConfig.Source)
		if err != nil {
			return nil, fmt.Errorf("failed to create a new keydbstore: %v", err)
		}

		health.RegisterPeriodicFunc(
			"DB operational", time.Minute, dbStore.HealthCheck)
		keyService = keydbstore.NewCachedKeyService(dbStore)
	}

	if doBootstrap {
		err := bootstrap(keyService)
		if err != nil {
			logrus.Fatal(err.Error())
		}
		os.Exit(0)
	}

	cryptoServices := make(signer.CryptoServiceIndex)
	cryptoServices[data.ED25519Key] = keyService
	cryptoServices[data.ECDSAKey] = keyService
	return cryptoServices, nil
}

func getDefaultAlias(configuration *viper.Viper) (string, error) {
	defaultAlias := configuration.GetString("storage.default_alias")
	if defaultAlias == "" {
		// backwards compatibility - support this environment variable
		defaultAlias = configuration.GetString(defaultAliasEnv)
	}

	if defaultAlias == "" {
		return "", fmt.Errorf("must provide a default alias for the key DB")
	}
	logrus.Debug("Default Alias: ", defaultAlias)
	return defaultAlias, nil
}

// set up the GRPC server
func setupGRPCServer(signerConfig signer.Config) (*grpc.Server, net.Listener, error) {

	//RPC server setup
	kms := &api.KeyManagementServer{
		CryptoServices: signerConfig.CryptoServices,
	}
	ss := &api.SignerServer{
		CryptoServices: signerConfig.CryptoServices,
	}
	hs := ghealth.NewServer()

	lis, err := net.Listen("tcp", signerConfig.GRPCAddr)
	if err != nil {
		return nil, nil, fmt.Errorf("grpc server failed to listen on %s: %v",
			signerConfig.GRPCAddr, err)
	}

	creds := credentials.NewTLS(signerConfig.TLSConfig)
	opts := []grpc.ServerOption{grpc.Creds(creds)}
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterKeyManagementServer(grpcServer, kms)
	pb.RegisterSignerServer(grpcServer, ss)
	healthpb.RegisterHealthServer(grpcServer, hs)

	// Set status for both of the grpc service "KeyManagement" and "Signer", these are
	// the only two we have at present, if we add more grpc service in the future,
	// we should add a new line for that service here as well.
	hs.SetServingStatus(notary.HealthCheckKeyManagement, healthpb.HealthCheckResponse_SERVING)
	hs.SetServingStatus(notary.HealthCheckSigner, healthpb.HealthCheckResponse_SERVING)

	return grpcServer, lis, nil
}

func getAddrAndTLSConfig(configuration *viper.Viper) (string, *tls.Config, error) {
	tlsConfig, err := utils.ParseServerTLS(configuration, true)
	if err != nil {
		return "", nil, fmt.Errorf("unable to set up TLS: %s", err.Error())
	}

	grpcAddr := configuration.GetString("server.grpc_addr")
	if grpcAddr == "" {
		return "", nil, fmt.Errorf("grpc listen address required for server")
	}

	return grpcAddr, tlsConfig, nil
}

func bootstrap(s interface{}) error {
	store, ok := s.(storage.Bootstrapper)
	if !ok {
		return fmt.Errorf("store does not support bootstrapping")
	}
	return store.Bootstrap()
}
