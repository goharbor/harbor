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
	"strconv"
)

// NewLogger create a logger for a speicified job
func NewLogger(jobID int64) *log.Logger {
	logFile := GetJobLogPath(jobID)
	d := filepath.Dir(logFile)
	if _, err := os.Stat(d); os.IsNotExist(err) {
		err := os.MkdirAll(d, 0660)
		if err != nil {
			log.Errorf("Failed to create directory for log file %s, the error: %v", logFile, err)
		}
	}
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		log.Errorf("Failed to open log file %s, the log of job %d will be printed to standard output, the error: %v", logFile, jobID, err)
		f = os.Stdout
	}
	return log.New(f, log.NewTextFormatter(), log.InfoLevel)
}

// GetJobLogPath returns the absolute path in which the job log file is located.
func GetJobLogPath(jobID int64) string {
	f := fmt.Sprintf("job_%d.log", jobID)
	k := jobID / 1000
	p := ""
	var d string
	for k > 0 {
		d = strconv.FormatInt(k%1000, 10)
		k = k / 1000
		if k > 0 && len(d) == 1 {
			d = "00" + d
		}
		if k > 0 && len(d) == 2 {
			d = "0" + d
		}

		p = filepath.Join(d, p)
	}
	p = filepath.Join(config.LogDir(), p, f)
	return p
}
