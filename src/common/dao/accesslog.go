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

package dao

import (
	"github.com/astaxie/beego/orm"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/utils/log"
)

// AddAccessLog persists the access logs
func AddAccessLog(accessLog models.AccessLog) error {
	// the max length of username in database is 255, replace the last
	// three characters with "..." if the length is greater than 256
	if len(accessLog.Username) > 255 {
		accessLog.Username = accessLog.Username[:252] + "..."
	}

	o := GetOrmer()
	_, err := o.Insert(&accessLog)
	return err
}

// GetTotalOfAccessLogs ...
func GetTotalOfAccessLogs(query *models.LogQueryParam) (int64, error) {
	return logQueryConditions(query).Count()
}

// GetAccessLogs gets access logs according to different conditions
func GetAccessLogs(query *models.LogQueryParam) ([]models.AccessLog, error) {
	qs := logQueryConditions(query).OrderBy("-op_time")

	if query != nil && query.Pagination != nil {
		size := query.Pagination.Size
		if size > 0 {
			qs = qs.Limit(size)

			page := query.Pagination.Page
			if page > 0 {
				qs = qs.Offset((page - 1) * size)
			}
		}
	}

	logs := []models.AccessLog{}
	_, err := qs.All(&logs)
	return logs, err
}

func logQueryConditions(query *models.LogQueryParam) orm.QuerySeter {
	qs := GetOrmer().QueryTable(&models.AccessLog{})

	if query == nil {
		return qs
	}

	if len(query.ProjectIDs) > 0 {
		qs = qs.Filter("project_id__in", query.ProjectIDs)
	}
	if len(query.Username) != 0 {
		qs = qs.Filter("username__contains", query.Username)
	}
	if len(query.Repository) != 0 {
		qs = qs.Filter("repo_name__contains", query.Repository)
	}
	if len(query.Tag) != 0 {
		qs = qs.Filter("repo_tag__contains", query.Tag)
	}
	operations := []string{}
	for _, operation := range query.Operations {
		if len(operation) > 0 {
			operations = append(operations, operation)
		}
	}
	if len(operations) > 0 {
		qs = qs.Filter("operation__in", operations)
	}
	if query.BeginTime != nil {
		qs = qs.Filter("op_time__gte", query.BeginTime)
	}
	if query.EndTime != nil {
		qs = qs.Filter("op_time__lte", query.EndTime)
	}

	return qs
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
