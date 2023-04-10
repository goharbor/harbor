// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package sweeper

import (
	"fmt"
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

	logFiles, err := os.ReadDir(fs.workDir)
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
		logFileInfo, ise := logFile.Info()
		if ise != nil {
			continue
		}
		if logFileInfo.ModTime().Add(time.Duration(fs.duration) * oneDay).Before(time.Now()) {
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
