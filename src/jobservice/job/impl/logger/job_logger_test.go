package logger

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/utils"
)

func TestJobLogger(t *testing.T) {
	oldConfig := config.DefaultConfig.LoggerConfig
	defer func() {
		config.DefaultConfig.LoggerConfig = oldConfig
	}()

	fakeJobID := "fake_job_id"
	stdLogger := &config.LoggerConfig{
		Kind:     config.LoggerKindStdOut,
		LogLevel: "INFO",
	}
	fileLogger := &config.LoggerConfig{
		Kind:          config.LoggerKindFile,
		LogLevel:      "ERROR",
		BasePath:      os.TempDir(),
		ArchivePeriod: 5,
	}
	config.DefaultConfig.LoggerConfig = []*config.LoggerConfig{stdLogger, fileLogger}

	logFilePath := path.Join(os.TempDir(), fmt.Sprintf("%s.log", fakeJobID))
	defer func() {
		os.Remove(logFilePath)
	}()

	logger, err := New(fakeJobID)
	if err != nil {
		t.Fatal(err)
	}

	logger.Info("info")
	logger.Infof("infof=%s", "info")
	logger.Debug("debug")
	logger.Debugf("debugf=%s", "debug")
	logger.Error("error")
	logger.Errorf("errorf=%s", "error")
	logger.Fatal("fatal")
	logger.Fatalf("fatalf=%s", "fatal")
	logger.Warning("warning")
	logger.Warningf("warningf=%s", "warning")

	if !utils.FileExists(logFilePath) {
		t.Fatalf("expect log file %s existing but not", logFilePath)
	}
}
