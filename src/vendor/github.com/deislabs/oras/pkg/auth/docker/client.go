package docker

import (
	"os"

	"github.com/deislabs/oras/pkg/auth"

	"github.com/docker/cli/cli/config"
	"github.com/docker/cli/cli/config/configfile"
	"github.com/docker/cli/cli/config/credentials"
	"github.com/pkg/errors"
)

// Client provides authentication operations for docker registries.
type Client struct {
	configs []*configfile.ConfigFile
}

// NewClient creates a new auth client based on provided config paths.
// If not config path is provided, the default path is used.
// Credentials are read from the first config and fall backs to next.
// All changes will only be written to the first config file.
func NewClient(configPaths ...string) (auth.Client, error) {
	if len(configPaths) == 0 {
		cfg, err := config.Load(config.Dir())
		if err != nil {
			return nil, err
		}
		if !cfg.ContainsAuth() {
			cfg.CredentialsStore = credentials.DetectDefaultStore(cfg.CredentialsStore)
		}

		return &Client{
			configs: []*configfile.ConfigFile{cfg},
		}, nil
	}

	var configs []*configfile.ConfigFile
	for _, path := range configPaths {
		cfg, err := loadConfigFile(path)
		if err != nil {
			return nil, errors.Wrap(err, path)
		}
		configs = append(configs, cfg)
	}

	return &Client{
		configs: configs,
	}, nil
}

func (c *Client) primaryCredentialsStore(hostname string) credentials.Store {
	return c.configs[0].GetCredentialsStore(hostname)
}

// loadConfigFile reads the configuration files from the given path.
func loadConfigFile(path string) (*configfile.ConfigFile, error) {
	cfg := configfile.New(path)
	if _, err := os.Stat(path); err == nil {
		file, err := os.Open(path)
		if err != nil {
			return nil, err
		}
		defer file.Close()
		if err := cfg.LoadFromReader(file); err != nil {
			return nil, err
		}
	} else if !os.IsNotExist(err) {
		return nil, err
	}
	if !cfg.ContainsAuth() {
		cfg.CredentialsStore = credentials.DetectDefaultStore(cfg.CredentialsStore)
	}
	return cfg, nil
}
