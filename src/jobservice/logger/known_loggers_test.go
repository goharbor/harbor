package logger

import (
	"github.com/goharbor/harbor/src/jobservice/logger/backend"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

// Test GetLoggerName
func TestGetLoggerName(t *testing.T) {
	uuid := "uuid_for_unit_test"
	l, err := backend.NewDBLogger(uuid, "DEBUG", 4)
	require.Nil(t, err)
	require.Equal(t, LoggerNameDB, GetLoggerName(l))

	stdLog := backend.NewStdOutputLogger("DEBUG", backend.StdErr, 4)
	require.Equal(t, LoggerNameStdOutput, GetLoggerName(stdLog))

	fileLog, err := backend.NewFileLogger("DEBUG", path.Join(os.TempDir(), "TestFileLogger.log"), 4)
	require.Nil(t, err)
	require.Equal(t, LoggerNameFile, GetLoggerName(fileLog))
}
