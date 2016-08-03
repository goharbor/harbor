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
	"github.com/vmware/harbor/utils/log"
)

// AddAccessLog persists the access logs
func AddAccessLog(accessLog models.AccessLog) error {
	o := GetOrmer()
	p, err := o.Raw(`insert into access_log
		 (user_id, project_id, repo_name, repo_tag, guid, operation, op_time)
		 values (?, ?, ?, ?, ?, ?, now())`).Prepare()
	if err != nil {
		return err
	}
	defer p.Close()

	_, err = p.Exec(accessLog.UserID, accessLog.ProjectID, accessLog.RepoName, accessLog.RepoTag, accessLog.GUID, accessLog.Operation)

	return err
}

//GetAccessLogs gets access logs according to different conditions
func GetAccessLogs(accessLog models.AccessLog) ([]models.AccessLog, error) {

	o := GetOrmer()
	sql := `select a.log_id, u.username, a.repo_name, a.repo_tag, a.operation, a.op_time
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
	if accessLog.RepoName != "" {
		sql += ` and a.repo_name = ? `
		queryParam = append(queryParam, accessLog.RepoName)
	}
	if accessLog.RepoTag != "" {
		sql += ` and a.repo_tag = ? `
		queryParam = append(queryParam, accessLog.RepoTag)
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
func AccessLog(username, projectName, repoName, repoTag, action string) error {
	o := GetOrmer()
	sql := "insert into  access_log (user_id, project_id, repo_name, repo_tag, operation, op_time) " +
		"select (select user_id as user_id from user where username=?), " +
		"(select project_id as project_id from project where name=?), ?, ?, ?, now() "
	_, err := o.Raw(sql, username, projectName, repoName, repoTag, action).Exec()

	if err != nil {
		log.Errorf("error in AccessLog: %v ", err)
	}
	return err
}

//GetRecentLogs returns recent logs according to parameters
func GetRecentLogs(userID, linesNum int, startTime, endTime string) ([]models.AccessLog, error) {
	var recentLogList []models.AccessLog
	queryParam := make([]interface{}, 1)

	sql := "select log_id, access_log.user_id, project_id, repo_name, repo_tag, GUID, operation, op_time, username from access_log left join  user on access_log.user_id=user.user_id where project_id in (select distinct project_id from project_member where user_id = ?)"
	queryParam = append(queryParam, userID)
	if startTime != "" {
		sql += " and op_time >= ?"
		queryParam = append(queryParam, startTime)
	}

	if endTime != "" {
		sql += " and op_time <= ?"
		queryParam = append(queryParam, endTime)
	}

	sql += " order by op_time desc"
	if linesNum != 0 {
		sql += " limit ?"
		queryParam = append(queryParam, linesNum)
	}
	o := GetOrmer()
	_, err := o.Raw(sql, queryParam).QueryRows(&recentLogList)
	if err != nil {
		return nil, err
	}
	return recentLogList, nil
}

//GetTopRepos return top  accessed public repos
func GetTopRepos(countNum int) ([]models.TopRepo, error) {

	o := GetOrmer()
	// hide the where condition: project.public = 1, Can add to the sql when necessary.
	sql := "select repo_name, COUNT(repo_name) as access_count from access_log left join project on access_log.project_id=project.project_id where access_log.operation = 'pull' group by repo_name order by access_count desc limit ? "
	queryParam := []interface{}{}
	queryParam = append(queryParam, countNum)
	var list []models.TopRepo
	_, err := o.Raw(sql, queryParam).QueryRows(&list)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return list, nil
	}
	placeHolder := make([]string, len(list))
	repos := make([]string, len(list))
	for i, v := range list {
		repos[i] = v.RepoName
		placeHolder[i] = "?"
	}
	placeHolderStr := strings.Join(placeHolder, ",")
	queryParam = nil
	queryParam = append(queryParam, repos)
	var usrnameList []models.TopRepo
	sql = `select a.username as creator, a.repo_name from (select access_log.repo_name, user.username,
	access_log.op_time from user left join access_log on user.user_id = access_log.user_id where 
	access_log.operation = 'push' and access_log.repo_name in (######) order by access_log.repo_name,
	access_log.op_time ASC) a group by a.repo_name`
	sql = strings.Replace(sql, "######", placeHolderStr, 1)
	_, err = o.Raw(sql, queryParam).QueryRows(&usrnameList)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(list); i++ {
		for _, v := range usrnameList {
			if v.RepoName == list[i].RepoName {
				//			list[i].Creator = v.Creator
				break
			}
		}
	}
	return list, nil
}
