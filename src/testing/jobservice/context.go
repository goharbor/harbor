package jobservice

import (
	"context"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
	"github.com/stretchr/testify/mock"
)

// MockJobContext mocks job context interface.
// TODO: Maybe moved to a separate `mock` pkg for sharing in future.
type MockJobContext struct {
	mock.Mock
}

// Build ...
func (mjc *MockJobContext) Build(tracker job.Tracker) (job.Context, error) {
	args := mjc.Called(tracker)
	c := args.Get(0)
	if c != nil {
		return c.(job.Context), nil
	}

	return nil, args.Error(1)
}

// Get ...
func (mjc *MockJobContext) Get(prop string) (interface{}, bool) {
	args := mjc.Called(prop)
	return args.Get(0), args.Bool(1)
}

// SystemContext ...
func (mjc *MockJobContext) SystemContext() context.Context {
	return context.TODO()
}

// Checkin ...
func (mjc *MockJobContext) Checkin(status string) error {
	args := mjc.Called(status)
	return args.Error(0)
}

// OPCommand ...
func (mjc *MockJobContext) OPCommand() (job.OPCommand, bool) {
	args := mjc.Called()
	return args.Get(0).(job.OPCommand), args.Bool(1)
}

// GetLogger ...
func (mjc *MockJobContext) GetLogger() logger.Interface {
	return &MockJobLogger{}
}

// Tracker ...
func (mjc *MockJobContext) Tracker() job.Tracker {
	args := mjc.Called()
	if t := args.Get(0); t != nil {
		return t.(job.Tracker)
	}

	return nil
}

// MockJobLogger mocks the job logger interface.
// TODO: Maybe moved to a separate `mock` pkg for sharing in future.
type MockJobLogger struct {
	mock.Mock
}

// Debug ...
func (mjl *MockJobLogger) Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Debugf ...
func (mjl *MockJobLogger) Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Info ...
func (mjl *MockJobLogger) Info(v ...interface{}) {
	logger.Info(v...)
}

// Infof ...
func (mjl *MockJobLogger) Infof(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Warning ...
func (mjl *MockJobLogger) Warning(v ...interface{}) {
	logger.Warning(v...)
}

// Warningf ...
func (mjl *MockJobLogger) Warningf(format string, v ...interface{}) {
	logger.Warningf(format, v...)
}

// Error ...
func (mjl *MockJobLogger) Error(v ...interface{}) {
	logger.Error(v...)
}

// Errorf ...
func (mjl *MockJobLogger) Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Fatal ...
func (mjl *MockJobLogger) Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Fatalf ...
func (mjl *MockJobLogger) Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}
