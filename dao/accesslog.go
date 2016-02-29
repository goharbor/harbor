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

package dao

import (
	"strings"

	"github.com/vmware/harbor/models"

	"github.com/astaxie/beego/orm"
)

// AddAccessLog persists the access logs
func AddAccessLog(accessLog models.AccessLog) error {
	o := orm.NewOrm()
	p, err := o.Raw(`insert into access_log
		 (user_id, project_id, repo_name, guid, operation, op_time)
		 values (?, ?, ?, ?, ?, now())`).Prepare()
	if err != nil {
		return err
	}
	defer p.Close()

	_, err = p.Exec(accessLog.UserID, accessLog.ProjectID, accessLog.RepoName, accessLog.GUID, accessLog.Operation)

	return err
}

//GetAccessLogs gets access logs according to different conditions
func GetAccessLogs(accessLog models.AccessLog) ([]models.AccessLog, error) {

	o := orm.NewOrm()
	sql := `select a.log_id, u.username, a.repo_name, a.operation, a.op_time
		from access_log a left join user u on a.user_id = u.user_id
		where a.project_id = ? `
	queryParam := make([]interface{}, 1)
	queryParam = append(queryParam, accessLog.ProjectID)

	if accessLog.UserID != 0 {
		sql += ` and a.user_id = ? `
		queryParam = append(queryParam, accessLog.UserID)
	}
	if accessLog.Operation != "" {
		sql += ` and a.operation = ? `
		queryParam = append(queryParam, accessLog.Operation)
	}
	if accessLog.Username != "" {
		sql += ` and u.username like ? `
		queryParam = append(queryParam, accessLog.Username)
	}
	if accessLog.Keywords != "" {
		sql += ` and a.operation in ( `
		keywordList := strings.Split(accessLog.Keywords, "/")
		num := len(keywordList)
		for i := 0; i < num; i++ {
			if keywordList[i] != "" {
				if i == num-1 {
					sql += `?)`
				} else {
					sql += `?,`
				}
				queryParam = append(queryParam, keywordList[i])
			}
		}
	}
	if accessLog.BeginTimestamp > 0 {
		sql += ` and a.op_time >= ? `
		queryParam = append(queryParam, accessLog.BeginTime)
	}
	if accessLog.EndTimestamp > 0 {
		sql += ` and a.op_time <= ? `
		queryParam = append(queryParam, accessLog.EndTime)
	}

	sql += ` order by a.op_time desc `

	var accessLogList []models.AccessLog
	_, err := o.Raw(sql, queryParam).QueryRows(&accessLogList)
	if err != nil {
		return nil, err
	}
	return accessLogList, nil
}

// AccessLog ...
func AccessLog(username, projectName, repoName, action string) error {
	o := orm.NewOrm()
	sql := "insert into  access_log (user_id, project_id, repo_name, operation, op_time) " +
		"select (select user_id as user_id from user where username=?), " +
		"(select project_id as project_id from project where name=?), ?, ?, now() "
	_, err := o.Raw(sql, username, projectName, repoName, action).Exec()

	return err
}
