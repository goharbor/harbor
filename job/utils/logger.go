package utils

import (
	"fmt"

	"github.com/vmware/harbor/job/config"
	"github.com/vmware/harbor/utils/log"
	"os"
	"path/filepath"
)

func NewLogger(jobID int64) *log.Logger {
	logFile := GetJobLogPath(jobID)
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Errorf("Failed to open log file %s, the log of job %d will be printed to standard output", logFile, jobID)
		f = os.Stdout
	}
	return log.New(f, log.NewTextFormatter(), log.InfoLevel)
}

func GetJobLogPath(jobID int64) string {
	fn := fmt.Sprintf("job_%d.log", jobID)
	return filepath.Join(config.LogDir(), fn)
}
