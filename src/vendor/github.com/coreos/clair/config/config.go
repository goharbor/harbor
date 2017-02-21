// Copyright 2015 clair authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"errors"
	"io/ioutil"
	"os"
	"time"

	"github.com/fernet/fernet-go"
	"gopkg.in/yaml.v2"
)

// ErrDatasourceNotLoaded is returned when the datasource variable in the configuration file is not loaded properly
var ErrDatasourceNotLoaded = errors.New("could not load configuration: no database source specified")

// File represents a YAML configuration file that namespaces all Clair
// configuration under the top-level "clair" key.
type File struct {
	Clair Config `yaml:"clair"`
}

// Config is the global configuration for an instance of Clair.
type Config struct {
	Database *DatabaseConfig
	Updater  *UpdaterConfig
	Notifier *NotifierConfig
	API      *APIConfig
}

// DatabaseConfig is the configuration used to specify how Clair connects
// to a database.
type DatabaseConfig struct {
	Source    string
	CacheSize int
}

// UpdaterConfig is the configuration for the Updater service.
type UpdaterConfig struct {
	Interval time.Duration
}

// NotifierConfig is the configuration for the Notifier service and its registered notifiers.
type NotifierConfig struct {
	Attempts         int
	RenotifyInterval time.Duration
	Params           map[string]interface{} `yaml:",inline"`
}

// APIConfig is the configuration for the API service.
type APIConfig struct {
	Port                      int
	HealthPort                int
	Timeout                   time.Duration
	PaginationKey             string
	CertFile, KeyFile, CAFile string
}

// DefaultConfig is a configuration that can be used as a fallback value.
func DefaultConfig() Config {
	return Config{
		Database: &DatabaseConfig{
			CacheSize: 16384,
		},
		Updater: &UpdaterConfig{
			Interval: 1 * time.Hour,
		},
		API: &APIConfig{
			Port:       6060,
			HealthPort: 6061,
			Timeout:    900 * time.Second,
		},
		Notifier: &NotifierConfig{
			Attempts:         5,
			RenotifyInterval: 2 * time.Hour,
		},
	}
}

// Load is a shortcut to open a file, read it, and generate a Config.
// It supports relative and absolute paths. Given "", it returns DefaultConfig.
func Load(path string) (config *Config, err error) {
	var cfgFile File
	cfgFile.Clair = DefaultConfig()
	if path == "" {
		return &cfgFile.Clair, nil
	}

	f, err := os.Open(os.ExpandEnv(path))
	if err != nil {
		return
	}
	defer f.Close()

	d, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(d, &cfgFile)
	if err != nil {
		return
	}
	config = &cfgFile.Clair

	if config.Database.Source == "" {
		err = ErrDatasourceNotLoaded
		return
	}

	// Generate a pagination key if none is provided.
	if config.API.PaginationKey == "" {
		var key fernet.Key
		if err = key.Generate(); err != nil {
			return
		}
		config.API.PaginationKey = key.Encode()
	} else {
		_, err = fernet.DecodeKey(config.API.PaginationKey)
		if err != nil {
			return
		}
	}

	return
}
