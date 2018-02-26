// +build !windows

package utils

import (
	"fmt"
	"os"
	"syscall"

	"github.com/Sirupsen/logrus"
)

// LogLevelSignalHandle will increase/decrease the logging level via the signal we get.
func LogLevelSignalHandle(sig os.Signal) {
	switch sig {
	case syscall.SIGUSR1:
		if err := AdjustLogLevel(true); err != nil {
			fmt.Printf("Attempt to increase log level failed, will remain at %s level, error: %s\n", logrus.GetLevel(), err)
			return
		}
	case syscall.SIGUSR2:
		if err := AdjustLogLevel(false); err != nil {
			fmt.Printf("Attempt to decrease log level failed, will remain at %s level, error: %s\n", logrus.GetLevel(), err)
			return
		}
	}

	fmt.Println("Successfully setting log level to", logrus.GetLevel())
}
