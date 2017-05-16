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

package dao

import (
	"strings"

	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils/log"
)

// AddAccessLog persists the access logs
func AddAccessLog(accessLog models.AccessLog) error {
	o := GetOrmer()
	_, err := o.Insert(&accessLog)
	return err
}

// GetTotalOfAccessLogs ...
func GetTotalOfAccessLogs(query models.AccessLog) (int64, error) {
	o := GetOrmer()

	queryParam := []interface{}{}

	sql := `select count(*) from access_log al 
		where al.project_id = ?`
	queryParam = append(queryParam, query.ProjectID)

	sql += genFilterClauses(query, &queryParam)

	var total int64
	if err := o.Raw(sql, queryParam).QueryRow(&total); err != nil {
		return 0, err
	}
	return total, nil
}

//GetAccessLogs gets access logs according to different conditions
func GetAccessLogs(query models.AccessLog, limit, offset int64) ([]models.AccessLog, error) {
	o := GetOrmer()

	queryParam := []interface{}{}
	sql := `select al.log_id, al.username, al.repo_name, 
			al.repo_tag, al.operation, al.op_time 
		from access_log al 
		where al.project_id = ? `
	queryParam = append(queryParam, query.ProjectID)

	sql += genFilterClauses(query, &queryParam)

	sql += ` order by al.op_time desc `

	sql = paginateForRawSQL(sql, limit, offset)

	logs := []models.AccessLog{}
	_, err := o.Raw(sql, queryParam).QueryRows(&logs)
	if err != nil {
		return logs, err
	}

	return logs, nil
}

func genFilterClauses(query models.AccessLog, queryParam *[]interface{}) string {
	sql := ""

	if query.Username != "" {
		sql += ` and al.username like ? `
		*queryParam = append(*queryParam, "%"+escape(query.Username)+"%")
	}

	if query.Operation != "" {
		sql += ` and al.operation = ? `
		*queryParam = append(*queryParam, query.Operation)
	}
	if query.RepoName != "" {
		sql += ` and al.repo_name = ? `
		*queryParam = append(*queryParam, query.RepoName)
	}
	if query.RepoTag != "" {
		sql += ` and al.repo_tag = ? `
		*queryParam = append(*queryParam, query.RepoTag)
	}
	if query.Keywords != "" {
		sql += ` and al.operation in ( `
		keywordList := strings.Split(query.Keywords, "/")
		num := len(keywordList)
		for i := 0; i < num; i++ {
			if keywordList[i] != "" {
				if i == num-1 {
					sql += `?)`
				} else {
					sql += `?,`
				}
				*queryParam = append(*queryParam, keywordList[i])
			}
		}
	}
	if query.BeginTimestamp > 0 {
		sql += ` and al.op_time >= ? `
		*queryParam = append(*queryParam, query.BeginTime)
	}
	if query.EndTimestamp > 0 {
		sql += ` and al.op_time <= ? `
		*queryParam = append(*queryParam, query.EndTime)
	}

	return sql
}

//GetRecentLogs returns recent logs according to parameters
func GetRecentLogs(username string, linesNum int, startTime, endTime string) ([]models.AccessLog, error) {
	logs := []models.AccessLog{}

	isAdmin, err := IsAdminRole(username)
	if err != nil {
		return logs, err
	}

	queryParam := []interface{}{}
	sql := `select log_id, username, project_id, repo_name, repo_tag, GUID, operation, op_time  
		from access_log `

	hasWhere := false
	if !isAdmin {
		sql += ` where project_id in 
			(select distinct project_id 
				from project_member pm
				join user u
				on  pm.user_id = u.user_id
				where u.username = ?) `
		queryParam = append(queryParam, username)
		hasWhere = true
	}

	if startTime != "" {
		if hasWhere {
			sql += " and op_time >= ?"
		} else {
			sql += " where op_time >= ?"
			hasWhere = true
		}

		queryParam = append(queryParam, startTime)
	}

	if endTime != "" {
		if hasWhere {
			sql += " and op_time <= ?"
		} else {
			sql += " where op_time <= ?"
			hasWhere = true
		}

		queryParam = append(queryParam, endTime)
	}

	sql += " order by op_time desc"
	if linesNum != 0 {
		sql += " limit ?"
		queryParam = append(queryParam, linesNum)
	}

	_, err = GetOrmer().Raw(sql, queryParam).QueryRows(&logs)
	if err != nil {
		return logs, err
	}
	return logs, nil
}

// CountPull ...
func CountPull(repoName string) (int64, error) {
	o := GetOrmer()
	num, err := o.QueryTable("access_log").Filter("repo_name", repoName).Filter("operation", "pull").Count()
	if err != nil {
		log.Errorf("error in CountPull: %v ", err)
		return 0, err
	}
	return num, nil
}
