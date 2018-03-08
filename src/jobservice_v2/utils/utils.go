// Copyright 2018 The Harbor Authors. All rights reserved.

//Package utils provides reusable and sharable utilities for other packages and components.
package utils

import (
	"os"
	"strings"
)

//IsEmptyStr check if the specified str is empty (len ==0) after triming prefix and suffix spaces.
func IsEmptyStr(str string) bool {
	return len(strings.TrimSpace(str)) == 0
}

//ReadEnv return the value of env variable.
func ReadEnv(key string) string {
	return os.Getenv(key)
}

//FileExists check if the specified exists.
func FileExists(file string) bool {
	if !IsEmptyStr(file) {
		if _, err := os.Stat(file); err == nil {
			return true
		}
	}

	return false
}

//IsValidPort check if port is valid.
func IsValidPort(port uint) bool {
	return port != 0 && port < 65536
}
