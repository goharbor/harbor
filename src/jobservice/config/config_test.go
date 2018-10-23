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
	cfg := &Configuration{}
	if err := cfg.Load("./config.not-existing.yaml", false); err == nil {
		t.Fatalf("Load config from none-existing document, expect none nil error but got '%s'\n", err)
	}
}

func TestConfigLoadingSucceed(t *testing.T) {
	if err := CreateLogDir(); err != nil {
		t.Fatal(err)
	}

	cfg := &Configuration{}
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

	cfg := &Configuration{}
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
	if cfg.LoggerConfig.BasePath != "/tmp" {
		t.Fatalf("expect log base path '/tmp' but got '%s'\n", cfg.LoggerConfig.BasePath)
	}
	if cfg.LoggerConfig.LogLevel != "DEBUG" {
		t.Fatalf("expect log level 'DEBUG' but got '%s'\n", cfg.LoggerConfig.LogLevel)
	}
	if cfg.LoggerConfig.ArchivePeriod != 5 {
		t.Fatalf("expect log archive period 5 but got '%d'\n", cfg.LoggerConfig.ArchivePeriod)
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

	if basePath := GetLogBasePath(); basePath != "/tmp/job_logs" {
		t.Fatalf("expect default logger base path '/tmp/job_logs' but got '%s'\n", basePath)
	}

	if lvl := GetLogLevel(); lvl != "INFO" {
		t.Fatalf("expect default logger level 'INFO' but got '%s'\n", lvl)
	}

	if period := GetLogArchivePeriod(); period != 1 {
		t.Fatalf("expect default log archive period 1 but got '%d'\n", period)
	}

	redisURL := DefaultConfig.PoolConfig.RedisPoolCfg.RedisURL
	if redisURL != "redis://redis:6379" {
		t.Fatalf("expect redisURL '%s' but got '%s'\n", "redis://redis:6379", redisURL)
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
	os.Setenv("JOB_SERVICE_LOGGER_BASE_PATH", "/tmp")
	os.Setenv("JOB_SERVICE_LOGGER_LEVEL", "DEBUG")
	os.Setenv("JOB_SERVICE_LOGGER_ARCHIVE_PERIOD", "5")
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
	os.Unsetenv("JOB_SERVICE_LOGGER_BASE_PATH")
	os.Unsetenv("JOB_SERVICE_LOGGER_LEVEL")
	os.Unsetenv("JOB_SERVICE_LOGGER_ARCHIVE_PERIOD")
}

func CreateLogDir() error {
	return os.MkdirAll("/tmp/job_logs", 0755)
}

func RemoveLogDir() error {
	return os.Remove("/tmp/job_logs")
}
