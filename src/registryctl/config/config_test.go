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
	"bytes"
	"context"
	"testing"

	dcontext "github.com/docker/distribution/context"
	_ "github.com/docker/distribution/registry/storage/driver/filesystem"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestConfigDoesNotExists(t *testing.T) {
	cfg := &Configuration{}
	err := cfg.Load("./config.not-existing.yaml", false)
	assert.NotNil(t, err)
}

func TestConfigLoadingWithEnv(t *testing.T) {
	preserveLogrusLevel(t)
	t.Setenv("REGISTRYCTL_PROTOCOL", "https")
	t.Setenv("PORT", "1000")
	t.Setenv("LOG_LEVEL", "DEBUG")

	cfg := &Configuration{}
	err := cfg.Load("../config_test.yml", true)
	assert.Nil(t, err)
	assert.Equal(t, "https", cfg.Protocol)
	assert.Equal(t, "1000", cfg.Port)
	assert.Equal(t, "DEBUG", cfg.LogLevel)
	assert.Equal(t, "../reg_conf_test.yml", cfg.RegistryConfig)
	assert.Equal(t, logrus.DebugLevel, logrus.GetLevel())
}

func TestConfigLoadingWithYml(t *testing.T) {
	preserveLogrusLevel(t)
	cfg := &Configuration{}
	err := cfg.Load("../config_test.yml", false)
	assert.Nil(t, err)
	assert.Equal(t, "http", cfg.Protocol)
	assert.Equal(t, "1234", cfg.Port)
	assert.Equal(t, "ERROR", cfg.LogLevel)
	assert.Equal(t, "../reg_conf_test.yml", cfg.RegistryConfig)
	assert.True(t, cfg.StorageDriver.Name() == "filesystem")
	assert.Equal(t, logrus.ErrorLevel, logrus.GetLevel())
}

func TestGetLogLevel(t *testing.T) {
	preserveLogrusLevel(t)
	err := DefaultConfig.Load("../config_test.yml", false)
	assert.Nil(t, err)
	assert.Equal(t, "ERROR", GetLogLevel())
}

func TestGetJobAuthSecret(t *testing.T) {
	t.Setenv("JOBSERVICE_SECRET", "test_job_secret")
	assert.Equal(t, "test_job_secret", GetJobAuthSecret())
}

func TestGetUIAuthSecret(t *testing.T) {
	t.Setenv("CORE_SECRET", "test_core_secret")
	assert.Equal(t, "test_core_secret", GetUIAuthSecret())
}

func TestConfigureLogging(t *testing.T) {
	tests := []struct {
		name  string
		level string
		want  logrus.Level
	}{
		{"warning", "warning", logrus.WarnLevel},
		{"warn alias", "warn", logrus.WarnLevel},
		{"uppercase", "WARNING", logrus.WarnLevel},
		{"debug", "debug", logrus.DebugLevel},
		{"info", "info", logrus.InfoLevel},
		{"error", "error", logrus.ErrorLevel},
		{"fatal", "fatal", logrus.FatalLevel},
		{"empty defaults to info", "", logrus.InfoLevel},
		{"invalid defaults to info", "invalid", logrus.InfoLevel},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			previous := logrus.GetLevel()
			t.Cleanup(func() {
				logrus.SetLevel(previous)
			})

			// Ensure the test verifies that configuration sets the level,
			// rather than accidentally retaining an existing default.
			logrus.SetLevel(logrus.PanicLevel)

			cfg := &Configuration{LogLevel: tt.level}
			cfg.configureLogging()

			assert.Equal(t, tt.want, logrus.GetLevel())
		})
	}
}

func TestConfigureLoggingSuppressesDistributionInfo(t *testing.T) {
	logger := logrus.StandardLogger()
	previousLevel := logger.GetLevel()
	previousOutput := logger.Out
	t.Cleanup(func() {
		logger.SetLevel(previousLevel)
		logger.SetOutput(previousOutput)
	})

	var output bytes.Buffer
	logger.SetOutput(&output)

	cfg := &Configuration{LogLevel: "info"}
	cfg.configureLogging()

	distributionLogger := dcontext.GetLogger(context.Background())
	distributionLogger.Infof("Deleting blob: %s", "test-blob")
	assert.Contains(t, output.String(), "Deleting blob: test-blob")

	output.Reset()
	cfg.LogLevel = "warning"
	cfg.configureLogging()

	distributionLogger.Infof("Deleting blob: %s", "test-blob")
	distributionLogger.Warn("test storage warning")

	assert.NotContains(t, output.String(), "Deleting blob:")
	assert.Contains(t, output.String(), "test storage warning")
}

func preserveLogrusLevel(t *testing.T) {
	t.Helper()
	previous := logrus.GetLevel()
	t.Cleanup(func() {
		logrus.SetLevel(previous)
	})
}
