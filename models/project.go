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

// Project holds the details of a project.
type Project struct {
	ProjectID       int64     `orm:"column(project_id)" json:"ProjectId"`
	OwnerID         int       `orm:"column(owner_id)" json:"OwnerId"`
	Name            string    `orm:"column(name)"`
	CreationTime    time.Time `orm:"column(creation_time)"`
	CreationTimeStr string
	Deleted         int `orm:"column(deleted)"`
	UserID          int `json:"UserId"`
	OwnerName       string
	Public          int `orm:"column(public)"`
	//This field does not have correspondent column in DB, this is just for UI to disable button
	Togglable bool

	UpdateTime time.Time `orm:"update_time" json:"update_time"`
}
