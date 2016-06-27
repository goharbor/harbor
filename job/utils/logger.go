/*
    Copyright (c) 2016 VMware, Inc. All Rights Reserved.
    Licensed under the Apache License, Version 2.0 (the "License");
    you may not use this file except in compliance with the License.
    You may obtain a copy of the License at
        
        http://www.apache.org/licenses/LICENSE-2.0
        
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/

package utils

import (
	"fmt"

	"github.com/vmware/harbor/job/config"
	"github.com/vmware/harbor/utils/log"
	"os"
	"path/filepath"
)

// NewLogger create a logger for a speicified job
func NewLogger(jobID int64) *log.Logger {
	logFile := GetJobLogPath(jobID)
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Errorf("Failed to open log file %s, the log of job %d will be printed to standard output", logFile, jobID)
		f = os.Stdout
	}
	return log.New(f, log.NewTextFormatter(), log.InfoLevel)
}

// GetJobLogPath returns the absolute path in which the job log file is located.
func GetJobLogPath(jobID int64) string {
	fn := fmt.Sprintf("job_%d.log", jobID)
	return filepath.Join(config.LogDir(), fn)
}
