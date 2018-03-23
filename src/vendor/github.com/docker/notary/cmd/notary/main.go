package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/docker/notary"
	"github.com/docker/notary/passphrase"
	"github.com/docker/notary/tuf/data"
	"github.com/docker/notary/version"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	configDir        = ".notary/"
	defaultServerURL = "https://notary-server:4443"
)

type usageTemplate struct {
	Use   string
	Short string
	Long  string
}

type cobraRunE func(cmd *cobra.Command, args []string) error

func (u usageTemplate) ToCommand(run cobraRunE) *cobra.Command {
	c := cobra.Command{
		Use:   u.Use,
		Short: u.Short,
		Long:  u.Long,
	}
	if run != nil {
		// newer versions of cobra support a run function that returns an error,
		// but in the meantime, this should help ease the transition
		c.RunE = run
	}
	return &c
}

func pathRelativeToCwd(path string) string {
	if path == "" || filepath.IsAbs(path) {
		return path
	}
	cwd, err := os.Getwd()
	if err != nil {
		return filepath.Clean(path)
	}
	return filepath.Clean(filepath.Join(cwd, path))
}

type notaryCommander struct {
	// this needs to be set
	getRetriever func() notary.PassRetriever

	// these are for command line parsing - no need to set
	debug             bool
	verbose           bool
	trustDir          string
	configFile        string
	remoteTrustServer string

	tlsCAFile   string
	tlsCertFile string
	tlsKeyFile  string
}

func (n *notaryCommander) parseConfig() (*viper.Viper, error) {
	n.setVerbosityLevel()

	// Get home directory for current user
	homeDir, err := homedir.Dir()
	if err != nil {
		return nil, fmt.Errorf("cannot get current user home directory: %v", err)
	}
	if homeDir == "" {
		return nil, fmt.Errorf("cannot get current user home directory")
	}

	config := viper.New()

	// By default our trust directory (where keys are stored) is in ~/.notary/
	defaultTrustDir := filepath.Join(homeDir, filepath.Dir(configDir))

	// If there was a commandline configFile set, we parse that.
	// If there wasn't we attempt to find it on the default location ~/.notary/config.json
	if n.configFile != "" {
		config.SetConfigFile(n.configFile)
	} else {
		config.SetConfigFile(filepath.Join(defaultTrustDir, "config.json"))
	}

	// Setup the configuration details into viper
	config.SetDefault("trust_dir", defaultTrustDir)
	config.SetDefault("remote_server", map[string]string{"url": defaultServerURL})

	// Find and read the config file
	if err := config.ReadInConfig(); err != nil {
		logrus.Debugf("Configuration file not found, using defaults")

		// If we were passed in a configFile via command linen flags, bail if it doesn't exist,
		// otherwise ignore it: we can use the defaults
		if n.configFile != "" || !os.IsNotExist(err) {
			return nil, fmt.Errorf("error opening config file: %v", err)
		}
	}

	// At this point we either have the default value or the one set by the config.
	// Either way, some command-line flags have precedence and overwrites the value
	if n.trustDir != "" {
		config.Set("trust_dir", pathRelativeToCwd(n.trustDir))
	}
	if n.tlsCAFile != "" {
		config.Set("remote_server.root_ca", pathRelativeToCwd(n.tlsCAFile))
	}
	if n.tlsCertFile != "" {
		config.Set("remote_server.tls_client_cert", pathRelativeToCwd(n.tlsCertFile))
	}
	if n.tlsKeyFile != "" {
		config.Set("remote_server.tls_client_key", pathRelativeToCwd(n.tlsKeyFile))
	}
	if n.remoteTrustServer != "" {
		config.Set("remote_server.url", n.remoteTrustServer)
	}

	// Expands all the possible ~/ that have been given, either through -d or config
	// If there is no error, use it, if not, just attempt to use whatever the user gave us
	expandedTrustDir, err := homedir.Expand(config.GetString("trust_dir"))
	if err == nil {
		config.Set("trust_dir", expandedTrustDir)
	}
	logrus.Debugf("Using the following trust directory: %s", config.GetString("trust_dir"))

	return config, nil
}

func (n *notaryCommander) GetCommand() *cobra.Command {
	notaryCmd := cobra.Command{
		Use:           "notary",
		Short:         "Notary allows the creation of trusted collections.",
		Long:          "Notary allows the creation and management of collections of signed targets, allowing the signing and validation of arbitrary content.",
		SilenceUsage:  true, // we don't want to print out usage for EVERY error
		SilenceErrors: true, // we do our own error reporting with fatalf
		Run:           func(cmd *cobra.Command, args []string) { cmd.Usage() },
	}
	notaryCmd.SetOutput(os.Stdout)
	notaryCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number of notary",
		Long:  "Print the version number of notary",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("notary\n Version:    %s\n Git commit: %s\n", version.NotaryVersion, version.GitCommit)
		},
	})

	notaryCmd.PersistentFlags().StringVarP(
		&n.trustDir, "trustDir", "d", "", "Directory where the trust data is persisted to")
	notaryCmd.PersistentFlags().StringVarP(
		&n.configFile, "configFile", "c", "", "Path to the configuration file to use")
	notaryCmd.PersistentFlags().BoolVarP(&n.verbose, "verbose", "v", false, "Verbose output")
	notaryCmd.PersistentFlags().BoolVarP(&n.debug, "debug", "D", false, "Debug output")
	notaryCmd.PersistentFlags().StringVarP(&n.remoteTrustServer, "server", "s", "", "Remote trust server location")
	notaryCmd.PersistentFlags().StringVar(&n.tlsCAFile, "tlscacert", "", "Trust certs signed only by this CA")
	notaryCmd.PersistentFlags().StringVar(&n.tlsCertFile, "tlscert", "", "Path to TLS certificate file")
	notaryCmd.PersistentFlags().StringVar(&n.tlsKeyFile, "tlskey", "", "Path to TLS key file")

	cmdKeyGenerator := &keyCommander{
		configGetter: n.parseConfig,
		getRetriever: n.getRetriever,
		input:        os.Stdin,
	}

	cmdDelegationGenerator := &delegationCommander{
		configGetter: n.parseConfig,
		retriever:    n.getRetriever(),
	}

	cmdTUFGenerator := &tufCommander{
		configGetter: n.parseConfig,
		retriever:    n.getRetriever(),
	}

	notaryCmd.AddCommand(cmdKeyGenerator.GetCommand())
	notaryCmd.AddCommand(cmdDelegationGenerator.GetCommand())

	cmdTUFGenerator.AddToCommand(&notaryCmd)

	return &notaryCmd
}

func main() {
	notaryCommander := &notaryCommander{getRetriever: getPassphraseRetriever}
	notaryCmd := notaryCommander.GetCommand()
	if err := notaryCmd.Execute(); err != nil {
		notaryCmd.Println("")
		fatalf(err.Error())
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Printf("* fatal: "+format+"\n", args...)
	os.Exit(1)
}

func askConfirm(input io.Reader) bool {
	var res string
	if _, err := fmt.Fscanln(input, &res); err != nil {
		return false
	}
	if strings.EqualFold(res, "y") || strings.EqualFold(res, "yes") {
		return true
	}
	return false
}

func getPassphraseRetriever() notary.PassRetriever {
	baseRetriever := passphrase.PromptRetriever()
	env := map[string]string{
		"root":       os.Getenv("NOTARY_ROOT_PASSPHRASE"),
		"targets":    os.Getenv("NOTARY_TARGETS_PASSPHRASE"),
		"snapshot":   os.Getenv("NOTARY_SNAPSHOT_PASSPHRASE"),
		"delegation": os.Getenv("NOTARY_DELEGATION_PASSPHRASE"),
	}

	return func(keyName string, alias string, createNew bool, numAttempts int) (string, bool, error) {
		if v := env[alias]; v != "" {
			return v, numAttempts > 1, nil
		}
		// For delegation roles, we can also try the "delegation" alias if it is specified
		// Note that we don't check if the role name is for a delegation to allow for names like "user"
		// since delegation keys can be shared across repositories
		// This cannot be a base role or imported key, though.
		if v := env["delegation"]; !data.IsBaseRole(data.RoleName(alias)) && v != "" {
			return v, numAttempts > 1, nil
		}
		return baseRetriever(keyName, alias, createNew, numAttempts)
	}
}

// Set the logging level to fatal on default, or the most specific level the user specified (debug or error)
func (n *notaryCommander) setVerbosityLevel() {
	if n.debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if n.verbose {
		logrus.SetLevel(logrus.ErrorLevel)
	} else {
		logrus.SetLevel(logrus.FatalLevel)
	}
	logrus.SetOutput(os.Stderr)
}
