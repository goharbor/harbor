package logger

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/goharbor/harbor/src/jobservice/config"
	"github.com/goharbor/harbor/src/jobservice/logger/backend"
)

const (
	fakeLogFile = "f00000000000000000000000.log"
	fakeLogID   = "f00000000000000000000000"
	fakeJobID   = "f00000000000000000000001"
	fakeJobID2  = "f00000000000000000000002"
)

// Test one single std logger
func TestGetLoggerSingleStd(t *testing.T) {
	l, err := GetLogger(BackendOption("STD_OUTPUT", "DEBUG", nil))
	if err != nil {
		t.Fatal(err)
	}

	l.Debugf("Verify logger testing: %s", "case_1")

	lSettings := map[string]interface{}{}
	lSettings["output"] = backend.StdErr
	l, err = GetLogger(BackendOption("STD_OUTPUT", "ERROR", lSettings))
	if err != nil {
		t.Fatal(err)
	}

	l.Errorf("Verify logger testing: %s", "case_2")

	// With empty options
	l, err = GetLogger()
	if err != nil {
		t.Fatal(err)
	}

	l.Warningf("Verify logger testing: %s", "case_3")
}

// Test one single file logger
func TestGetLoggerSingleFile(t *testing.T) {
	_, err := GetLogger(BackendOption("FILE", "DEBUG", nil))
	if err == nil {
		t.Fatalf("expect non nil error when creating file logger with empty settings but got nil error: %s", "case_4")
	}

	lSettings := map[string]interface{}{}
	lSettings["base_dir"] = os.TempDir()
	lSettings["filename"] = fmt.Sprintf("%s.log", fakeJobID)
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), lSettings["filename"].(string))); err != nil {
			t.Error(err)
		}
	}()

	l, err := GetLogger(BackendOption("FILE", "DEBUG", lSettings))
	if err != nil {
		t.Fatal(err)
	}

	l.Debugf("Verify logger testing: %s", "case_5")
}

// Test getting multi loggers
func TestGetLoggersMulti(t *testing.T) {
	lSettings := map[string]interface{}{}
	lSettings["base_dir"] = os.TempDir()
	lSettings["filename"] = fmt.Sprintf("%s.log", fakeJobID2)
	defer func() {
		if err := os.Remove(path.Join(os.TempDir(), lSettings["filename"].(string))); err != nil {
			t.Error(err)
		}
	}()

	ops := make([]Option, 0)
	ops = append(
		ops,
		BackendOption("STD_OUTPUT", "DEBUG", nil),
		BackendOption("FILE", "DEBUG", lSettings),
	)

	l, err := GetLogger(ops...)
	if err != nil {
		t.Fatal(err)
	}

	l.Infof("Verify logger testing: %s", "case_6")
}

// Test getting sweepers
func TestGetSweeper(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, err := GetSweeper(ctx)
	if err == nil {
		t.Fatalf("expect non nil error but got nil error when getting sweeper with empty settings: %s", "case_7")
	}

	_, err = GetSweeper(ctx, SweeperOption("STD_OUTPUT", 1, nil))
	if err == nil {
		t.Fatalf("expect non nil error but got nil error when getting sweeper with name 'STD_OUTPUT': %s", "case_8")
	}

	sSettings := map[string]interface{}{}
	sSettings["work_dir"] = os.TempDir()
	s, err := GetSweeper(ctx, SweeperOption("FILE", 5, sSettings))
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Sweep()
	if err != nil {
		t.Fatalf("[%s] start sweeper error: %s", "case_9", err)
	}
}

// Test getting getters
func TestGetGetter(t *testing.T) {
	_, err := GetLogDataGetter()
	if err == nil {
		t.Fatalf("error should be returned if no options provided: %s", "case_10")
	}

	// no configured
	g, err := GetLogDataGetter(GetterOption("STD_OUTPUT", nil))
	if err != nil || g != nil {
		t.Fatalf("nil interface with nil error should be returned if no log data getter configured: %s", "case_11")
	}

	lSettings := map[string]interface{}{}
	_, err = GetLogDataGetter(GetterOption("FILE", lSettings))
	if err == nil {
		t.Fatalf("expect non nil error but got nil one: %s", "case_12")
	}

	lSettings["base_dir"] = os.TempDir()
	g, err = GetLogDataGetter(GetterOption("FILE", lSettings))
	if err != nil {
		t.Fatal(err)
	}

	logFile := path.Join(os.TempDir(), fakeLogFile)
	if err := ioutil.WriteFile(logFile, []byte("hello log getter"), 0644); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.Remove(logFile); err != nil {
			t.Error(err)
		}
	}()

	data, err := g.Retrieve(fakeLogID)
	if err != nil {
		t.Error(err)
	}

	if len(data) != 16 {
		t.Errorf("expect 16 bytes data but got %d bytes", len(data))
	}
}

// Test init
func TestLoggerInit(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	oldJobLoggerCfg := config.DefaultConfig.JobLoggerConfigs
	oldLoggerCfg := config.DefaultConfig.LoggerConfigs
	defer func() {
		config.DefaultConfig.JobLoggerConfigs = oldJobLoggerCfg
		config.DefaultConfig.LoggerConfigs = oldLoggerCfg
	}()

	config.DefaultConfig.JobLoggerConfigs = []*config.LoggerConfig{
		{
			Name:  "STD_OUTPUT",
			Level: "DEBUG",
			Settings: map[string]interface{}{
				"output": backend.StdErr,
			},
		},
		{
			Name:  "FILE",
			Level: "ERROR",
			Settings: map[string]interface{}{
				"base_dir": os.TempDir(),
			},
			Sweeper: &config.LogSweeperConfig{
				Duration: 5,
				Settings: map[string]interface{}{
					"work_dir": os.TempDir(),
				},
			},
		},
	}

	config.DefaultConfig.LoggerConfigs = []*config.LoggerConfig{
		{
			Name:  "STD_OUTPUT",
			Level: "DEBUG",
		},
	}

	if err := Init(ctx); err != nil {
		t.Fatal(err)
	}

	Debug("Verify logger init: case_13")
	Info("Verify logger init: case_13")
	Infof("Verify logger init: %s", "case_13")
	Error("Verify logger init: case_13")
	Errorf("Verify logger init: %s", "case_13")
}
