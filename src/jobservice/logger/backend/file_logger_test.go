package backend

import (
	"os"
	"path"
	"testing"
)

// Test file logger creation with non existing file path
func TestFileLoggerCreation(t *testing.T) {
	if _, err := NewFileLogger("DEBUG", "/non-existing/a.log", 4); err == nil {
		t.Fatalf("expect non nil error but got nil when creating file logger with non existing path")
	}
}

// Test file logger
func TestFileLogger(t *testing.T) {
	l, err := NewFileLogger("DEBUG", path.Join(os.TempDir(), "TestFileLogger.log"), 4)
	if err != nil {
		t.Fatal(err)
	}

	l.Debug("TestFileLogger")
	l.Info("TestFileLogger")
	l.Warning("TestFileLogger")
	l.Error("TestFileLogger")
	l.Debugf("%s", "TestFileLogger")
	l.Warningf("%s", "TestFileLogger")
	l.Errorf("%s", "TestFileLogger")
}
