// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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

package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
    "github.com/vmware/harbor/src/common/api"
)

//LogAPI handles request api/logs
type LogAPI struct {
	api.BaseAPI
	userID int
}

//Prepare validates the URL and the user
func (l *LogAPI) Prepare() {
	l.userID = l.ValidateUser()
}

//Get returns the recent logs according to parameters
func (l *LogAPI) Get() {
	var err error
	startTime := l.GetString("start_time")
	if len(startTime) != 0 {
		i, err := strconv.ParseInt(startTime, 10, 64)
		if err != nil {
			log.Errorf("Parse startTime to int error, err: %v", err)
			l.CustomAbort(http.StatusBadRequest, "startTime is not a valid integer")
		}
		startTime = time.Unix(i, 0).String()
	}

	endTime := l.GetString("end_time")
	if len(endTime) != 0 {
		j, err := strconv.ParseInt(endTime, 10, 64)
		if err != nil {
			log.Errorf("Parse endTime to int error, err: %v", err)
			l.CustomAbort(http.StatusBadRequest, "endTime is not a valid integer")
		}
		endTime = time.Unix(j, 0).String()
	}

	var linesNum int
	lines := l.GetString("lines")
	if len(lines) != 0 {
		linesNum, err = strconv.Atoi(lines)
		if err != nil {
			log.Errorf("Get parameters error--lines, err: %v", err)
			l.CustomAbort(http.StatusBadRequest, "bad request of lines")
		}
		if linesNum <= 0 {
			log.Warning("lines must be a positive integer")
			l.CustomAbort(http.StatusBadRequest, "lines is 0 or negative")
		}
	} else if len(startTime) == 0 && len(endTime) == 0 {
		linesNum = 10
	}

	var logList []models.AccessLog
	logList, err = dao.GetRecentLogs(l.userID, linesNum, startTime, endTime)
	if err != nil {
		log.Errorf("Get recent logs error, err: %v", err)
		l.CustomAbort(http.StatusInternalServerError, "Internal error")
	}
	l.Data["json"] = logList
	l.ServeJSON()
}
