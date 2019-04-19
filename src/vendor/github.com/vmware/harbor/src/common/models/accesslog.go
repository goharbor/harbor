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

package models

import (
	"time"
)

// AccessLog holds information about logs which are used to record the actions that user take to the resourses.
type AccessLog struct {
	LogID     int       `orm:"pk;auto;column(log_id)" json:"log_id"`
	Username  string    `orm:"column(username)"  json:"username"`
	ProjectID int64     `orm:"column(project_id)"  json:"project_id"`
	RepoName  string    `orm:"column(repo_name)" json:"repo_name"`
	RepoTag   string    `orm:"column(repo_tag)" json:"repo_tag"`
	GUID      string    `orm:"column(guid)"  json:"guid"`
	Operation string    `orm:"column(operation)" json:"operation"`
	OpTime    time.Time `orm:"column(op_time)" json:"op_time"`
}

// LogQueryParam is used to set query conditions when listing
// access logs.
type LogQueryParam struct {
	ProjectIDs []int64     // the IDs of projects to which the operation is done
	Username   string      // the operator's username of the log
	Repository string      // repository name
	Tag        string      // tag name
	Operations []string    // operations
	BeginTime  *time.Time  // the time after which the operation is done
	EndTime    *time.Time  // the time before which the operation is doen
	Pagination *Pagination // pagination information
}
