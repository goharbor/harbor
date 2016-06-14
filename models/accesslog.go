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

package models

import (
	"time"
)

// AccessLog holds information about logs which are used to record the actions that user take to the resourses.
type AccessLog struct {
	LogID          int       `orm:"pk;column(log_id)" json:"log_id"`
	UserID         int       `orm:"column(user_id)"  json:"user_id"`
	ProjectID      int64     `orm:"column(project_id)"  json:"project_id"`
	RepoName       string    `orm:"column(repo_name)" json:"repo_name"`
	RepoTag        string    `orm:"column(repo_tag)" json:"repo_tag"`
	GUID           string    `orm:"column(GUID)"  json:"guid"`
	Operation      string    `orm:"column(operation)" json:"operation"`
	OpTime         time.Time `orm:"column(op_time)" json:"op_time"`
	Username       string    `json:"username"`
	Keywords       string    `json:"keywords"`
	BeginTime      time.Time
	BeginTimestamp int64 `json:"begin_timestamp"`
	EndTime        time.Time
	EndTimestamp   int64 `json:"end_timestamp"`
}
