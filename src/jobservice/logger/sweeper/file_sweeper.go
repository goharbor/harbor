package sweeper

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"time"
)

const (
	oneDay = 24 * time.Hour
)

// FileSweeper is used to sweep the file logs
type FileSweeper struct {
	duration int
	workDir  string
}

// NewFileSweeper is constructor of FileSweeper
func NewFileSweeper(workDir string, duration int) *FileSweeper {
	return &FileSweeper{
		workDir:  workDir,
		duration: duration,
	}
}

// Sweep logs
func (fs *FileSweeper) Sweep() (int, error) {
	cleared := 0

	logFiles, err := ioutil.ReadDir(fs.workDir)
	if err != nil {
		return 0, fmt.Errorf("getting outdated log files under '%s' failed with error: %s", fs.workDir, err)
	}

	// Nothing to sweep
	if len(logFiles) == 0 {
		return 0, nil
	}

	// Start to sweep log files
	// Record all errors
	errs := make([]string, 0)
	for _, logFile := range logFiles {
		if logFile.ModTime().Add(time.Duration(fs.duration) * oneDay).Before(time.Now()) {
			logFilePath := path.Join(fs.workDir, logFile.Name())
			if err := os.Remove(logFilePath); err != nil {
				errs = append(errs, fmt.Sprintf("remove log file '%s' error: %s", logFilePath, err))
				continue // go on for next one
			}

			cleared++
		}
	}

	if len(errs) > 0 {
		err = fmt.Errorf("%s", strings.Join(errs, "\n"))
	}

	// Return error with high priority
	return cleared, err
}

// Duration for sweeping
func (fs *FileSweeper) Duration() int {
	return fs.duration
}
