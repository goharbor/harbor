// +build windows

package utils

import "os"

// LogLevelSignalHandle will do nothing, because we aren't currently supporting signal handling in windows
func LogLevelSignalHandle(sig os.Signal) {
}
