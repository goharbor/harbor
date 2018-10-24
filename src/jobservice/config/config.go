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

// Package config provides functions to handle the configurations of job service.
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os/user"
	"strconv"
	"strings"

	"github.com/goharbor/harbor/src/jobservice/utils"
	yaml "gopkg.in/yaml.v2"
)

const (
	jobServiceProtocol           = "JOB_SERVICE_PROTOCOL"
	jobServicePort               = "JOB_SERVICE_PORT"
	jobServiceHTTPCert           = "JOB_SERVICE_HTTPS_CERT"
	jobServiceHTTPKey            = "JOB_SERVICE_HTTPS_KEY"
	jobServiceWorkerPoolBackend  = "JOB_SERVICE_POOL_BACKEND"
	jobServiceWorkers            = "JOB_SERVICE_POOL_WORKERS"
	jobServiceRedisURL           = "JOB_SERVICE_POOL_REDIS_URL"
	jobServiceRedisNamespace     = "JOB_SERVICE_POOL_REDIS_NAMESPACE"
	jobServiceLoggerLevel        = "JOB_SERVICE_LOGGER_LEVEL"
	jobServiceCoreServerEndpoint = "CORE_URL"
	jobServiceAuthSecret         = "JOBSERVICE_SECRET"

	// JobServiceProtocolHTTPS points to the 'https' protocol
	JobServiceProtocolHTTPS = "https"
	// JobServiceProtocolHTTP points to the 'http' protocol
	JobServiceProtocolHTTP = "http"

	// JobServicePoolBackendRedis represents redis backend
	JobServicePoolBackendRedis = "redis"

	// Secret of UI
	uiAuthSecret = "CORE_SECRET"

	// Redis protocol schema
	redisSchema = "redis://"

	// All levels
	validLogLevels = "DEBUG,INFO,WARNING,ERROR,FATAL"

	// LoggerKindStdOut : STDOUT logger
	LoggerKindStdOut = "stdout"
	// LoggerKindStdError : STDERR logger
	LoggerKindStdError = "stderr"
	// LoggerKindFile : FILE logger
	LoggerKindFile = "file"
)

// DefaultConfig is the default configuration reference
var DefaultConfig = NewConfiguration()

// Configuration loads and keeps the related configuration items of job service.
type Configuration struct {
	// Protocol server listening on: https/http
	Protocol string `yaml:"protocol"`

	// Server listening port
	Port uint `yaml:"port"`

	AdminServer string `yaml:"admin_server"`

	// Additional config when using https
	HTTPSConfig *HTTPSConfig `yaml:"https_config,omitempty"`

	// Configurations of worker pool
	PoolConfig *PoolConfig `yaml:"worker_pool,omitempty"`

	// Logger configurations
	LoggerConfig []*LoggerConfig `yaml:"job_loggers,omitempty"`

	// Logger of job service itself
	ServiceLogger *ServiceLoggerConfig `yaml:"logger"`
}

// HTTPSConfig keeps additional configurations when using https protocol
type HTTPSConfig struct {
	Cert string `yaml:"cert"`
	Key  string `yaml:"key"`
}

// RedisPoolConfig keeps redis pool info.
type RedisPoolConfig struct {
	RedisURL  string `yaml:"redis_url"`
	Namespace string `yaml:"namespace"`
}

// PoolConfig keeps worker pool configurations.
type PoolConfig struct {
	// Worker concurrency
	WorkerCount  uint             `yaml:"workers"`
	Backend      string           `yaml:"backend"`
	RedisPoolCfg *RedisPoolConfig `yaml:"redis_pool,omitempty"`
}

// LoggerConfig keeps logger configurations.
type LoggerConfig struct {
	Kind          string `yaml:"kind"`
	BasePath      string `yaml:"path,omitempty"`
	LogLevel      string `yaml:"level"`
	ArchivePeriod uint   `yaml:"archive_period,omitempty"`
}

// ServiceLoggerConfig is logger settings of job service.
type ServiceLoggerConfig struct {
	LogLevel string `yaml:"level"`
}

// NewConfiguration is constructor of Configuration
func NewConfiguration() *Configuration {
	return &Configuration{
		LoggerConfig: make([]*LoggerConfig, 0),
	}
}

// Load the configuration options from the specified yaml file.
// If the yaml file is specified and existing, load configurations from yaml file first;
// If detecting env variables is specified, load configurations from env variables;
// Please pay attentions, the detected env variable will override the same configuration item loading from file.
//
// yamlFilePath	string: The path config yaml file
// readEnv       bool  : Whether detect the environment variables or not
func (c *Configuration) Load(yamlFilePath string, detectEnv bool) error {
	if !utils.IsEmptyStr(yamlFilePath) {
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
		// Load from env variables
		c.loadEnvs()
	}

	// translate redis url if needed
	if c.PoolConfig != nil && c.PoolConfig.RedisPoolCfg != nil {
		redisAddress := c.PoolConfig.RedisPoolCfg.RedisURL
		if !utils.IsEmptyStr(redisAddress) {
			if _, err := url.Parse(redisAddress); err != nil {
				if redisURL, ok := utils.TranslateRedisAddress(redisAddress); ok {
					c.PoolConfig.RedisPoolCfg.RedisURL = redisURL
				}
			} else {
				if !strings.HasPrefix(redisAddress, redisSchema) {
					c.PoolConfig.RedisPoolCfg.RedisURL = fmt.Sprintf("%s%s", redisSchema, redisAddress)
				}
			}
		}
	}

	// Validate settings
	return c.validate()
}

// GetAuthSecret get the auth secret from the env
func GetAuthSecret() string {
	return utils.ReadEnv(jobServiceAuthSecret)
}

// GetUIAuthSecret get the auth secret of UI side
func GetUIAuthSecret() string {
	return utils.ReadEnv(uiAuthSecret)
}

// GetAdminServerEndpoint return the admin server endpoint
func GetAdminServerEndpoint() string {
	return DefaultConfig.AdminServer
}

// GetFileLoggerSettings returns the settings of the file logger
func GetFileLoggerSettings() (string, uint, bool) {
	if len(DefaultConfig.LoggerConfig) > 0 {
		for _, logger := range DefaultConfig.LoggerConfig {
			if logger.Kind == LoggerKindFile {
				return logger.BasePath, logger.ArchivePeriod, true
			}
		}
	}

	return "", 0, false
}

// GetServiceLogLevel returns the log level of service logger
func GetServiceLogLevel() string {
	return DefaultConfig.ServiceLogger.LogLevel
}

// Load env variables
func (c *Configuration) loadEnvs() {
	prot := utils.ReadEnv(jobServiceProtocol)
	if !utils.IsEmptyStr(prot) {
		c.Protocol = prot
	}

	p := utils.ReadEnv(jobServicePort)
	if !utils.IsEmptyStr(p) {
		if po, err := strconv.Atoi(p); err == nil {
			c.Port = uint(po)
		}
	}

	// Only when protocol is https
	if c.Protocol == JobServiceProtocolHTTPS {
		cert := utils.ReadEnv(jobServiceHTTPCert)
		if !utils.IsEmptyStr(cert) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Cert = cert
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Cert: cert,
				}
			}
		}

		certKey := utils.ReadEnv(jobServiceHTTPKey)
		if !utils.IsEmptyStr(certKey) {
			if c.HTTPSConfig != nil {
				c.HTTPSConfig.Key = certKey
			} else {
				c.HTTPSConfig = &HTTPSConfig{
					Key: certKey,
				}
			}
		}
	}

	backend := utils.ReadEnv(jobServiceWorkerPoolBackend)
	if !utils.IsEmptyStr(backend) {
		if c.PoolConfig == nil {
			c.PoolConfig = &PoolConfig{}
		}
		c.PoolConfig.Backend = backend
	}

	workers := utils.ReadEnv(jobServiceWorkers)
	if !utils.IsEmptyStr(workers) {
		if count, err := strconv.Atoi(workers); err == nil {
			if c.PoolConfig == nil {
				c.PoolConfig = &PoolConfig{}
			}
			c.PoolConfig.WorkerCount = uint(count)
		}
	}

	if c.PoolConfig != nil && c.PoolConfig.Backend == JobServicePoolBackendRedis {
		redisURL := utils.ReadEnv(jobServiceRedisURL)
		if !utils.IsEmptyStr(redisURL) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.RedisURL = redisURL
		}

		rn := utils.ReadEnv(jobServiceRedisNamespace)
		if !utils.IsEmptyStr(rn) {
			if c.PoolConfig.RedisPoolCfg == nil {
				c.PoolConfig.RedisPoolCfg = &RedisPoolConfig{}
			}
			c.PoolConfig.RedisPoolCfg.Namespace = rn
		}
	}

	// admin server
	if coreServer := utils.ReadEnv(jobServiceCoreServerEndpoint); !utils.IsEmptyStr(coreServer) {
		c.AdminServer = coreServer
	}

	if jsLoggerLevel := utils.ReadEnv(jobServiceLoggerLevel); !utils.IsEmptyStr(jsLoggerLevel) {
		if c.ServiceLogger == nil {
			c.ServiceLogger = &ServiceLoggerConfig{}
		}
		c.ServiceLogger.LogLevel = jsLoggerLevel
	}
}

// Check if the configurations are valid settings.
func (c *Configuration) validate() error {
	if c.Protocol != JobServiceProtocolHTTPS &&
		c.Protocol != JobServiceProtocolHTTP {
		return fmt.Errorf("protocol should be %s or %s, but current setting is %s",
			JobServiceProtocolHTTP,
			JobServiceProtocolHTTPS,
			c.Protocol)
	}

	if !utils.IsValidPort(c.Port) {
		return fmt.Errorf("port number should be a none zero integer and less or equal 65535, but current is %d", c.Port)
	}

	if c.Protocol == JobServiceProtocolHTTPS {
		if c.HTTPSConfig == nil {
			return fmt.Errorf("certificate must be configured if serve with protocol %s", c.Protocol)
		}

		if utils.IsEmptyStr(c.HTTPSConfig.Cert) ||
			!utils.FileExists(c.HTTPSConfig.Cert) ||
			utils.IsEmptyStr(c.HTTPSConfig.Key) ||
			!utils.FileExists(c.HTTPSConfig.Key) {
			return fmt.Errorf("certificate for protocol %s is not correctly configured", c.Protocol)
		}
	}

	if c.PoolConfig == nil {
		return errors.New("no worker pool is configured")
	}

	if c.PoolConfig.Backend != JobServicePoolBackendRedis {
		return fmt.Errorf("worker pool backend %s does not support", c.PoolConfig.Backend)
	}

	// When backend is redis
	if c.PoolConfig.Backend == JobServicePoolBackendRedis {
		if c.PoolConfig.RedisPoolCfg == nil {
			return fmt.Errorf("redis pool must be configured when backend is set to '%s'", c.PoolConfig.Backend)
		}
		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.RedisURL) {
			return errors.New("URL of redis pool is empty")
		}

		if !strings.HasPrefix(c.PoolConfig.RedisPoolCfg.RedisURL, redisSchema) {
			return errors.New("Invalid redis URL")
		}

		if _, err := url.Parse(c.PoolConfig.RedisPoolCfg.RedisURL); err != nil {
			return fmt.Errorf("Invalid redis URL: %s", err.Error())
		}

		if utils.IsEmptyStr(c.PoolConfig.RedisPoolCfg.Namespace) {
			return errors.New("namespace of redis pool is required")
		}
	}

	if len(c.LoggerConfig) == 0 {
		return errors.New("missing logger config")
	}

	for _, loggerCfg := range c.LoggerConfig {
		validLoggerKinds := []string{LoggerKindFile, LoggerKindStdError, LoggerKindStdOut}
		if !strings.Contains(strings.Join(validLoggerKinds, ","), strings.Replace(loggerCfg.Kind, ",", "", -1)) {
			return fmt.Errorf("'%s' logger is not supported, only '%v' support", loggerCfg.Kind, validLoggerKinds)
		}
		if !strings.Contains(validLogLevels, strings.Replace(loggerCfg.LogLevel, ",", "", -1)) {
			return fmt.Errorf("logger level '%s' is invalid for '%s' logger, levels can only be one of: %s", loggerCfg.LogLevel, loggerCfg.Kind, validLogLevels)
		}

		// Extra checking for file logger
		if loggerCfg.Kind == LoggerKindFile {
			if strings.Contains(loggerCfg.BasePath, "~") {
				loggerCfg.BasePath = strings.Replace(loggerCfg.BasePath, "~", userHome(), 1)
			}

			if !utils.DirExists(loggerCfg.BasePath) {
				return fmt.Errorf("logger path '%s' of '%s' logger should be an existing dir", loggerCfg.BasePath, loggerCfg.Kind)
			}

			if loggerCfg.ArchivePeriod == 0 {
				return fmt.Errorf("logger archive period of '%s' logger should be greater than 0", loggerCfg.Kind)
			}
		}
	}

	if _, err := url.Parse(c.AdminServer); err != nil {
		return fmt.Errorf("invalid admin server endpoint: %s", err)
	}

	if c.ServiceLogger == nil || !strings.Contains(validLogLevels, strings.Replace(c.ServiceLogger.LogLevel, ",", "", -1)) {
		return fmt.Errorf("logger level '%s' is invalid for service logger, levels can only be one of: %s", c.ServiceLogger.LogLevel, validLogLevels)
	}

	return nil // valid
}

func userHome() string {
	u, err := user.Current()
	if err != nil {
		// return the pattern
		return "~"
	}

	return u.HomeDir
}
