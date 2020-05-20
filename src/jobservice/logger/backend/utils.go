package backend

import (
	"strings"

	"github.com/goharbor/harbor/src/lib/log"
)

func parseLevel(lvl string) log.Level {

	var level = log.WarningLevel

	switch strings.ToLower(lvl) {
	case "debug":
		level = log.DebugLevel
	case "info":
		level = log.InfoLevel
	case "warning":
		level = log.WarningLevel
	case "error":
		level = log.ErrorLevel
	case "fatal":
		level = log.FatalLevel
	default:
	}

	return level
}
