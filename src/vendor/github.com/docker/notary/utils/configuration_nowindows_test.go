// +build !windows

package utils

import (
	"syscall"
	"testing"

	"github.com/Sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestLogLevelSignalHandle(t *testing.T) {
	signalOperation := map[bool]syscall.Signal{
		optIncrement: syscall.SIGUSR1,
		optDecrement: syscall.SIGUSR2,
	}

	for _, expt := range logLevelExpectations {
		logrus.SetLevel(expt.startLevel)
		LogLevelSignalHandle(signalOperation[expt.increment])
		require.Equal(t, expt.endLevel, logrus.GetLevel())
	}
}
