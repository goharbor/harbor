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
	"fmt"

	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
)

//LogAPI handles request api/logs
type LogAPI struct {
	BaseController
	username   string
	isSysAdmin bool
}

//Prepare validates the URL and the user
func (l *LogAPI) Prepare() {
	l.BaseController.Prepare()
	if !l.SecurityCtx.IsAuthenticated() {
		l.HandleUnauthorized()
		return
	}
	l.username = l.SecurityCtx.GetUsername()
	l.isSysAdmin = l.SecurityCtx.IsSysAdmin()
}

//Get returns the recent logs according to parameters
func (l *LogAPI) Get() {
	page, size := l.GetPaginationParams()
	query := &models.LogQueryParam{
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	if !l.isSysAdmin {
		projects, err := l.ProjectMgr.GetByMember(l.username)
		if err != nil {
			l.HandleInternalServerError(fmt.Sprintf(
				"failed to get projects of user %s: %v", l.username, err))
			return
		}

		if len(projects) == 0 {
			l.SetPaginationHeader(0, page, size)
			l.Data["json"] = nil
			l.ServeJSON()
			return
		}

		ids := []int64{}
		for _, project := range projects {
			ids = append(ids, project.ProjectID)
		}
		query.ProjectIDs = ids
	}

	total, err := dao.GetTotalOfAccessLogs(query)
	if err != nil {
		l.HandleInternalServerError(fmt.Sprintf(
			"failed to get total of access logs: %v", err))
		return
	}

	logs, err := dao.GetAccessLogs(query)
	if err != nil {
		l.HandleInternalServerError(fmt.Sprintf(
			"failed to get access logs: %v", err))
		return
	}

	l.SetPaginationHeader(total, page, size)

	l.Data["json"] = logs
	l.ServeJSON()
}
