package backend

import (
	"testing"

	"github.com/goharbor/harbor/src/lib/log"
)

// Test parseLevel
func TestParseLevel(t *testing.T) {
	if l := parseLevel(""); l != log.WarningLevel {
		t.Errorf("expect level %d but got %d", log.WarningLevel, l)
	}
	if l := parseLevel("DEBUG"); l != log.DebugLevel {
		t.Errorf("expect level %d but got %d", log.DebugLevel, l)
	}
	if l := parseLevel("info"); l != log.InfoLevel {
		t.Errorf("expect level %d but got %d", log.InfoLevel, l)
	}
	if l := parseLevel("warning"); l != log.WarningLevel {
		t.Errorf("expect level %d but got %d", log.WarningLevel, l)
	}
	if l := parseLevel("error"); l != log.ErrorLevel {
		t.Errorf("expect level %d but got %d", log.ErrorLevel, l)
	}
	if l := parseLevel("FATAL"); l != log.FatalLevel {
		t.Errorf("expect level %d but got %d", log.FatalLevel, l)
	}
}
