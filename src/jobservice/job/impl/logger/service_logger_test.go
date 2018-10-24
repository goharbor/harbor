package logger

import (
	"testing"
)

// Very happy path
func TestServiceLogger(t *testing.T) {
	logger := NewServiceLogger("ERROR")

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
}
