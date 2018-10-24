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
	"os"
	"testing"
)

func TestConfigLoadingFailed(t *testing.T) {
	cfg := NewConfiguration()
	if err := cfg.Load("./config.not-existing.yaml", false); err == nil {
		t.Fatalf("Load config from none-existing document, expect none nil error but got '%s'\n", err)
	}
}

func TestConfigLoadingSucceed(t *testing.T) {
	if err := CreateLogDir(); err != nil {
		t.Fatal(err)
	}

	cfg := NewConfiguration()
	if err := cfg.Load("../config_test.yml", false); err != nil {
		t.Fatalf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}

	if err := RemoveLogDir(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigLoadingWithEnv(t *testing.T) {
	if err := CreateLogDir(); err != nil {
		t.Error(err)
	}
	setENV()

	cfg := NewConfiguration()
	if err := cfg.Load("../config_test.yml", true); err != nil {
		t.Fatalf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}

	if cfg.Protocol != "https" {
		t.Fatalf("expect protocol 'https', but got '%s'\n", cfg.Protocol)
	}
	if cfg.Port != 8989 {
		t.Fatalf("expect port 8989 but got '%d'\n", cfg.Port)
	}
	if cfg.PoolConfig.WorkerCount != 8 {
		t.Fatalf("expect workcount 8 but go '%d'\n", cfg.PoolConfig.WorkerCount)
	}
	if cfg.PoolConfig.RedisPoolCfg.RedisURL != "redis://arbitrary_username:password@8.8.8.8:6379/0" {
		t.Fatalf("expect redis URL 'localhost' but got '%s'\n", cfg.PoolConfig.RedisPoolCfg.RedisURL)
	}
	if cfg.PoolConfig.RedisPoolCfg.Namespace != "ut_namespace" {
		t.Fatalf("expect redis namespace 'ut_namespace' but got '%s'\n", cfg.PoolConfig.RedisPoolCfg.Namespace)
	}
	if cfg.ServiceLogger == nil || cfg.ServiceLogger.LogLevel != "ERROR" {
		t.Fatal("expect 'ERROR' log level of service logger but got invalid one")
	}

	unsetENV()
	if err := RemoveLogDir(); err != nil {
		t.Fatal(err)
	}
}

func TestDefaultConfig(t *testing.T) {
	if err := CreateLogDir(); err != nil {
		t.Fatal(err)
	}

	if err := DefaultConfig.Load("../config_test.yml", true); err != nil {
		t.Fatalf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}

	if endpoint := GetAdminServerEndpoint(); endpoint != "http://127.0.0.1:8888" {
		t.Fatalf("expect default admin server endpoint 'http://127.0.0.1:8888' but got '%s'\n", endpoint)
	}

	redisURL := DefaultConfig.PoolConfig.RedisPoolCfg.RedisURL
	if redisURL != "redis://redis:6379" {
		t.Fatalf("expect redisURL '%s' but got '%s'\n", "redis://redis:6379", redisURL)
	}

	if len(DefaultConfig.LoggerConfig) != 2 {
		t.Fatalf("expect stdout and file 2 loggers but got %d", len(DefaultConfig.LoggerConfig))
	}

	found := 0
	for _, logger := range DefaultConfig.LoggerConfig {
		switch logger.Kind {
		case "stderr":
			found++
			if logger.LogLevel != "ERROR" {
				t.Fatalf("expect log level 'ERROR' of '%s' logger but got '%s'\n", logger.Kind, logger.LogLevel)
			}
		case "file":
			found++
			if logger.LogLevel != "INFO" {
				t.Fatalf("expect log level 'INFO' of '%s' logger but got '%s'\n", logger.Kind, logger.LogLevel)
			}
			if logger.BasePath != "/tmp/job_logs" {
				t.Fatalf("expect log base path '/tmp/job_logs' of '%s' logger but got '%s'\n", logger.Kind, logger.BasePath)
			}
			if logger.ArchivePeriod != 5 {
				t.Fatalf("expect log archive period 5 of '%s' logger but got '%d'\n", logger.Kind, logger.ArchivePeriod)
			}
		}
	}

	if found != 2 {
		t.Fatalf("expect stderr and file 2 loggers but got %d", found)
	}

	// Utility functions
	if GetServiceLogLevel() != "INFO" {
		t.Fatalf("expect default service logger level is 'INFO' but got '%s'", GetServiceLogLevel())
	}

	if _, _, ok := GetFileLoggerSettings(); !ok {
		t.Fatal("failed to get file logger settings")
	}

	if err := RemoveLogDir(); err != nil {
		t.Fatal(err)
	}
}

func setENV() {
	os.Setenv("JOB_SERVICE_PROTOCOL", "https")
	os.Setenv("JOB_SERVICE_PORT", "8989")
	os.Setenv("JOB_SERVICE_HTTPS_CERT", "../server.crt")
	os.Setenv("JOB_SERVICE_HTTPS_KEY", "../server.key")
	os.Setenv("JOB_SERVICE_POOL_BACKEND", "redis")
	os.Setenv("JOB_SERVICE_POOL_WORKERS", "8")
	os.Setenv("JOB_SERVICE_POOL_REDIS_URL", "8.8.8.8:6379,100,password,0")
	os.Setenv("JOB_SERVICE_POOL_REDIS_NAMESPACE", "ut_namespace")
	os.Setenv("JOB_SERVICE_LOGGER_LEVEL", "ERROR")
}

func unsetENV() {
	os.Unsetenv("JOB_SERVICE_PROTOCOL")
	os.Unsetenv("JOB_SERVICE_PORT")
	os.Unsetenv("JOB_SERVICE_HTTPS_CERT")
	os.Unsetenv("JOB_SERVICE_HTTPS_KEY")
	os.Unsetenv("JOB_SERVICE_POOL_BACKEND")
	os.Unsetenv("JOB_SERVICE_POOL_WORKERS")
	os.Unsetenv("JOB_SERVICE_POOL_REDIS_URL")
	os.Unsetenv("JOB_SERVICE_POOL_REDIS_NAMESPACE")
	os.Unsetenv("JOB_SERVICE_LOGGER_LEVEL")
}

func CreateLogDir() error {
	return os.MkdirAll("/tmp/job_logs", 0755)
}

func RemoveLogDir() error {
	return os.Remove("/tmp/job_logs")
}
