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

package api

import (
	"net/http"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/utils/log"
)

//LogAPI handles request api/logs
type LogAPI struct {
	BaseAPI
	userID int
}

//Prepare validates the URL and the user
func (l *LogAPI) Prepare() {
	l.userID = l.ValidateUser()
}

//Get returns the recent logs according to parameters
func (l *LogAPI) Get() {
	lines, err := l.GetInt("lines")
	startTime := l.GetString("start_time")
	endTime := l.GetString("end_time")

	if err != nil {
		log.Errorf("Get parameters error--lines, err: %v", err)
		l.CustomAbort(http.StatusBadRequest, "bad request of lines")
	}
	if lines <= 0 {
		lines = 10
	}
	if len(startTime) <= 0 {
		log.Errorf("Get parameters error--startTime: %s", startTime)
		l.CustomAbort(http.StatusBadRequest, "bad request of startTime")
	}
	if len(endTime) <= 0 {
		log.Errorf("Get parameters error--endTime: %s", endTime)
		l.CustomAbort(http.StatusBadRequest, "bad request of endTime")
	}
	var logList []models.AccessLog
	logList, err = dao.GetRecentLogs(lines, startTime, endTime)
	if err != nil {
		l.CustomAbort(http.StatusInternalServerError, "Internal error")
		return
	}
	l.Data["json"] = logList
	l.ServeJSON()
}
