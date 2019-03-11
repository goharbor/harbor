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
	cfg := &Configuration{}
	if err := cfg.Load("../config_test.yml", false); err != nil {
		t.Fatalf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}
}

func TestConfigLoadingWithEnv(t *testing.T) {
	setENV()

	cfg := &Configuration{}
	if err := cfg.Load("../config_test.yml", true); err != nil {
		t.Fatalf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}

	if cfg.Protocol != "https" {
		t.Errorf("expect protocol 'https', but got '%s'\n", cfg.Protocol)
	}
	if cfg.Port != 8989 {
		t.Errorf("expect port 8989 but got '%d'\n", cfg.Port)
	}
	if cfg.PoolConfig.WorkerCount != 8 {
		t.Errorf("expect workcount 8 but go '%d'\n", cfg.PoolConfig.WorkerCount)
	}
	if cfg.PoolConfig.RedisPoolCfg.RedisURL != "redis://arbitrary_username:password@8.8.8.8:6379/0" {
		t.Errorf("expect redis URL 'localhost' but got '%s'\n", cfg.PoolConfig.RedisPoolCfg.RedisURL)
	}
	if cfg.PoolConfig.RedisPoolCfg.Namespace != "ut_namespace" {
		t.Errorf("expect redis namespace 'ut_namespace' but got '%s'\n", cfg.PoolConfig.RedisPoolCfg.Namespace)
	}
	if GetAuthSecret() != "js_secret" {
		t.Errorf("expect auth secret 'js_secret' but got '%s'", GetAuthSecret())
	}
	if GetUIAuthSecret() != "core_secret" {
		t.Errorf("expect auth secret 'core_secret' but got '%s'", GetUIAuthSecret())
	}

	unsetENV()
}

func TestDefaultConfig(t *testing.T) {
	if err := DefaultConfig.Load("../config_test.yml", true); err != nil {
		t.Fatalf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}
	redisURL := DefaultConfig.PoolConfig.RedisPoolCfg.RedisURL
	if redisURL != "redis://localhost:6379" {
		t.Errorf("expect redisURL '%s' but got '%s'\n", "redis://localhost:6379", redisURL)
	}

	if len(DefaultConfig.JobLoggerConfigs) == 0 {
		t.Errorf("expect 2 job loggers configured but got %d", len(DefaultConfig.JobLoggerConfigs))
	}

	if len(DefaultConfig.LoggerConfigs) == 0 {
		t.Errorf("expect 1 loggers configured but got %d", len(DefaultConfig.LoggerConfigs))
	}

	// Only verify the complicated one
	theLogger := DefaultConfig.JobLoggerConfigs[1]
	if theLogger.Name != "FILE" {
		t.Fatalf("expect FILE logger but got %s", theLogger.Name)
	}
	if theLogger.Level != "INFO" {
		t.Errorf("expect INFO log level of FILE logger but got %s", theLogger.Level)
	}
	if len(theLogger.Settings) == 0 {
		t.Errorf("expect extra settings but got nothing")
	}
	if theLogger.Settings["base_dir"] != "/tmp/job_logs" {
		t.Errorf("expect extra setting base_dir to be '/tmp/job_logs' but got %s", theLogger.Settings["base_dir"])
	}
	if theLogger.Sweeper == nil {
		t.Fatalf("expect non nil sweeper of FILE logger but got nil")
	}
	if theLogger.Sweeper.Duration != 5 {
		t.Errorf("expect sweep duration to be 5 but got %d", theLogger.Sweeper.Duration)
	}
	if theLogger.Sweeper.Settings["work_dir"] != "/tmp/job_logs" {
		t.Errorf("expect work dir of sweeper of FILE logger to be '/tmp/job_logs' but got %s", theLogger.Sweeper.Settings["work_dir"])
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
	os.Setenv("JOBSERVICE_SECRET", "js_secret")
	os.Setenv("CORE_SECRET", "core_secret")
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
	os.Unsetenv("JOBSERVICE_SECRET")
	os.Unsetenv("CORE_SECRET")
}
