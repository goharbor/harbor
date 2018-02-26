package main

import (
	_ "expvar"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Sirupsen/logrus"
	"github.com/docker/notary/utils"
	"github.com/docker/notary/version"
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
)

const (
	jsonLogFormat = "json"
	debugAddr     = "localhost:8080"
)

type cmdFlags struct {
	debug       bool
	logFormat   string
	configFile  string
	doBootstrap bool
}

func setupFlags(flagStorage *cmdFlags) {
	// Setup flags
	flag.StringVar(&flagStorage.configFile, "config", "", "Path to configuration file")
	flag.BoolVar(&flagStorage.debug, "debug", false, "Show the version and exit")
	flag.StringVar(&flagStorage.logFormat, "logf", "json", "Set the format of the logs. Only 'json' and 'logfmt' are supported at the moment.")
	flag.BoolVar(&flagStorage.doBootstrap, "bootstrap", false, "Do any necessary setup of configured backend storage services")

	// this needs to be in init so that _ALL_ logs are in the correct format
	if flagStorage.logFormat == jsonLogFormat {
		logrus.SetFormatter(new(logrus.JSONFormatter))
	}

	flag.Usage = usage
}

func main() {
	flagStorage := cmdFlags{}
	setupFlags(&flagStorage)

	flag.Parse()

	if flagStorage.debug {
		go debugServer(debugAddr)
	}

	// when the signer starts print the version for debugging and issue logs later
	logrus.Infof("Version: %s, Git commit: %s", version.NotaryVersion, version.GitCommit)

	signerConfig, err := parseSignerConfig(flagStorage.configFile, flagStorage.doBootstrap)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	grpcServer, lis, err := setupGRPCServer(signerConfig)
	if err != nil {
		logrus.Fatal(err.Error())
	}

	if flagStorage.debug {
		log.Println("RPC server listening on", signerConfig.GRPCAddr)
	}

	c := utils.SetupSignalTrap(utils.LogLevelSignalHandle)
	if c != nil {
		defer signal.Stop(c)
	}

	grpcServer.Serve(lis)
}

func usage() {
	log.Println("usage:", os.Args[0], "<config>")
	flag.PrintDefaults()
}

// debugServer starts the debug server with pprof, expvar among other
// endpoints. The addr should not be exposed externally. For most of these to
// work, tls cannot be enabled on the endpoint, so it is generally separate.
func debugServer(addr string) {
	logrus.Infof("Debug server listening on %s", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		logrus.Fatalf("error listening on debug interface: %v", err)
	}
}
