package backend

import "testing"

// Test std logger
func TestStdLogger(t *testing.T) {
	l := NewStdOutputLogger("DEBUG", StdErr, 4)
	l.Debug("TestStdLogger")
	l.Debugf("%s", "TestStdLogger")
	l.Info("TestStdLogger")
	l.Infof("%s", "TestStdLogger")
	l.Warning("TestStdLogger")
	l.Warningf("%s", "TestStdLogger")
	l.Error("TestStdLogger")
	l.Errorf("%s", "TestStdLogger")
}
