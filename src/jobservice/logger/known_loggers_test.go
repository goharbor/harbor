package logger

import (
	"github.com/goharbor/harbor/src/jobservice/logger/backend"
	"github.com/stretchr/testify/require"
	"os"
	"path"
	"testing"
)

// TestKnownLoggers
func TestKnownLoggers(t *testing.T) {
	b := IsKnownLogger("Unknown")
	require.False(t, b)

	b = IsKnownLogger(LoggerNameFile)
	require.True(t, b)

	// no getter
	b = HasGetter(LoggerNameStdOutput)
	require.False(t, b)
	// has getter
	b = HasGetter(LoggerNameDB)
	require.True(t, b)

	// no sweeper
	b = HasSweeper(LoggerNameStdOutput)
	require.False(t, b)
	// has sweeper
	b = HasSweeper(LoggerNameDB)
	require.True(t, b)

	// unknown logger
	l := KnownLoggers("unknown")
	require.Nil(t, l)
	// known logger
	l = KnownLoggers(LoggerNameDB)
	require.NotNil(t, l)

	// unknown level
	b = IsKnownLevel("unknown")
	require.False(t, b)
	b = IsKnownLevel("")
	require.False(t, b)
	// known level
	b = IsKnownLevel(debugLevels[0])
	require.True(t, b)
}

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

	e := &Entry{}
	n := GetLoggerName(e)
	require.NotNil(t, n)
}
