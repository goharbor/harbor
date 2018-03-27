// Copyright 2018 The Harbor Authors. All rights reserved.
package config

import (
	"os"
	"testing"
)

func TestConfigLoadingFailed(t *testing.T) {
	cfg := &Configuration{}
	if err := cfg.Load("./config.not-existing.yaml", false); err == nil {
		t.Errorf("Load config from none-existing document, expect none nil error but got '%s'\n", err)
	}
}

func TestConfigLoadingSucceed(t *testing.T) {
	if err := CreateLogDir(); err != nil {
		t.Error(err)
	}

	cfg := &Configuration{}
	if err := cfg.Load("../config_test.yml", false); err != nil {
		t.Errorf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}

	if err := RemoveLogDir(); err != nil {
		t.Error(err)
	}
}

func TestConfigLoadingWithEnv(t *testing.T) {
	if err := CreateLogDir(); err != nil {
		t.Error(err)
	}
	setENV()

	cfg := &Configuration{}
	if err := cfg.Load("../config_test.yml", true); err != nil {
		t.Errorf("Load config from yaml file, expect nil error but got error '%s'\n", err)
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
	if cfg.PoolConfig.RedisPoolCfg.Host != "localhost" {
		t.Errorf("expect redis host 'localhost' but got '%s'\n", cfg.PoolConfig.RedisPoolCfg.Host)
	}
	if cfg.PoolConfig.RedisPoolCfg.Port != 7379 {
		t.Errorf("expect redis port '7379' but got '%d'\n", cfg.PoolConfig.RedisPoolCfg.Port)
	}
	if cfg.PoolConfig.RedisPoolCfg.Namespace != "ut_namespace" {
		t.Errorf("expect redis namespace 'ut_namespace' but got '%s'\n", cfg.PoolConfig.RedisPoolCfg.Namespace)
	}
	if cfg.LoggerConfig.BasePath != "/tmp" {
		t.Errorf("expect log base path '/tmp' but got '%s'\n", cfg.LoggerConfig.BasePath)
	}
	if cfg.LoggerConfig.LogLevel != "DEBUG" {
		t.Errorf("expect log level 'DEBUG' but got '%s'\n", cfg.LoggerConfig.LogLevel)
	}
	if cfg.LoggerConfig.ArchivePeriod != 5 {
		t.Errorf("expect log archive period 5 but got '%d'\n", cfg.LoggerConfig.ArchivePeriod)
	}

	unsetENV()
	if err := RemoveLogDir(); err != nil {
		t.Error(err)
	}
}

func TestDefaultConfig(t *testing.T) {
	if err := CreateLogDir(); err != nil {
		t.Error(err)
	}

	if err := DefaultConfig.Load("../config_test.yml", true); err != nil {
		t.Errorf("Load config from yaml file, expect nil error but got error '%s'\n", err)
	}

	if endpoint := GetAdminServerEndpoint(); endpoint != "http://localhost:9010/" {
		t.Errorf("expect default admin server endpoint 'http://localhost:9010/' but got '%s'\n", endpoint)
	}

	if basePath := GetLogBasePath(); basePath != "/tmp/job_logs" {
		t.Errorf("expect default logger base path '/tmp/job_logs' but got '%s'\n", basePath)
	}

	if lvl := GetLogLevel(); lvl != "INFO" {
		t.Errorf("expect default logger level 'INFO' but got '%s'\n", lvl)
	}

	if period := GetLogArchivePeriod(); period != 1 {
		t.Errorf("expect default log archive period 1 but got '%d'\n", period)
	}

	if err := RemoveLogDir(); err != nil {
		t.Error(err)
	}
}

func setENV() {
	os.Setenv("JOB_SERVICE_PROTOCOL", "https")
	os.Setenv("JOB_SERVICE_PORT", "8989")
	os.Setenv("JOB_SERVICE_HTTPS_CERT", "../server.crt")
	os.Setenv("JOB_SERVICE_HTTPS_KEY", "../server.key")
	os.Setenv("JOB_SERVICE_POOL_BACKEND", "redis")
	os.Setenv("JOB_SERVICE_POOL_WORKERS", "8")
	os.Setenv("JOB_SERVICE_POOL_REDIS_HOST", "localhost")
	os.Setenv("JOB_SERVICE_POOL_REDIS_PORT", "7379")
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
	os.Unsetenv("JOB_SERVICE_POOL_REDIS_HOST")
	os.Unsetenv("JOB_SERVICE_POOL_REDIS_PORT")
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
