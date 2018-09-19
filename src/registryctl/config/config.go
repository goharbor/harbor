// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"
)

// DefaultConfig ...
var DefaultConfig = &Configuration{}

// Configuration loads the configuration of registry controller.
type Configuration struct {
	Protocol    string `yaml:"protocol"`
	Port        string `yaml:"port"`
	LogLevel    string `yaml:"log_level"`
	HTTPSConfig struct {
		Cert string `yaml:"cert"`
		Key  string `yaml:"key"`
	} `yaml:"https_config,omitempty"`
}

// Load the configuration options from the specified yaml file.
func (c *Configuration) Load(yamlFilePath string, detectEnv bool) error {
	if len(yamlFilePath) != 0 {
		// Try to load from file first
		data, err := ioutil.ReadFile(yamlFilePath)
		if err != nil {
			return err
		}
		if err = yaml.Unmarshal(data, c); err != nil {
			return err
		}
	}

	if detectEnv {
		c.loadEnvs()
	}

	return nil
}

// GetLogLevel returns the log level
func GetLogLevel() string {
	return DefaultConfig.LogLevel
}

// GetJobAuthSecret get the auth secret from the env
func GetJobAuthSecret() string {
	return os.Getenv("JOBSERVICE_SECRET")
}

// GetUIAuthSecret get the auth secret of UI side
func GetUIAuthSecret() string {
	return os.Getenv("CORE_SECRET")
}

// loadEnvs Load env variables
func (c *Configuration) loadEnvs() {
	prot := os.Getenv("REGISTRYCTL_PROTOCOL")
	if len(prot) != 0 {
		c.Protocol = prot
	}

	p := os.Getenv("PORT")
	if len(p) != 0 {
		c.Port = p
	}

	// Only when protocol is https
	if c.Protocol == "HTTPS" {
		cert := os.Getenv("REGISTRYCTL_HTTPS_CERT")
		if len(cert) != 0 {
			c.HTTPSConfig.Cert = cert
		}

		certKey := os.Getenv("REGISTRYCTL_HTTPS_KEY")
		if len(certKey) != 0 {
			c.HTTPSConfig.Key = certKey
		}
	}

	loggerLevel := os.Getenv("LOG_LEVEL")
	if len(loggerLevel) != 0 {
		c.LogLevel = loggerLevel
	}

}
