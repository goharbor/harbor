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

// Project holds the details of a project.
// TODO remove useless attrs
type Project struct {
	ProjectID       int64                  `orm:"pk;auto;column(project_id)" json:"project_id"`
	OwnerID         int                    `orm:"column(owner_id)" json:"owner_id"`
	Name            string                 `orm:"column(name)" json:"name"`
	CreationTime    time.Time              `orm:"column(creation_time)" json:"creation_time"`
	UpdateTime      time.Time              `orm:"update_time" json:"update_time"`
	Deleted         int                    `orm:"column(deleted)" json:"deleted"`
	CreationTimeStr string                 `orm:"-" json:"creation_time_str"`
	OwnerName       string                 `orm:"-" json:"owner_name"`
	Togglable       bool                   `orm:"-"`
	Role            int                    `orm:"-" json:"current_user_role_id"`
	RepoCount       int                    `orm:"-" json:"repo_count"`
	Metadata        map[string]interface{} `orm:"-" json:"metadata"`

	// TODO remove
	Public                                     int    `orm:"column(public)" json:"public"`
	EnableContentTrust                         bool   `orm:"-" json:"enable_content_trust"`
	PreventVulnerableImagesFromRunning         bool   `orm:"-" json:"prevent_vulnerable_images_from_running"`
	PreventVulnerableImagesFromRunningSeverity string `orm:"-" json:"prevent_vulnerable_images_from_running_severity"`
	AutomaticallyScanImagesOnPush              bool   `orm:"-" json:"automatically_scan_images_on_push"`
}

// ProjectSorter holds an array of projects
type ProjectSorter struct {
	Projects []*Project
}

// Len returns the length of array in ProjectSorter
func (ps *ProjectSorter) Len() int {
	return len(ps.Projects)
}

// Less defines the comparison rules of project
func (ps *ProjectSorter) Less(i, j int) bool {
	return ps.Projects[i].Name < ps.Projects[j].Name
}

// Swap swaps the position of i and j
func (ps *ProjectSorter) Swap(i, j int) {
	ps.Projects[i], ps.Projects[j] = ps.Projects[j], ps.Projects[i]
}

// ProjectQueryParam can be used to set query parameters when listing projects.
// The query condition will be set in the query if its corresponding field
// is not nil. Leave it empty if you don't want to apply this condition.
//
// e.g.
// List all projects: query := nil
// List all public projects: query := &QueryParam{Public: true}
// List projects the owner of which is user1: query := &QueryParam{Owner:"user1"}
// List all public projects the owner of which is user1: query := &QueryParam{Owner:"user1",Public:true}
// List projects which user1 is member of: query := &QueryParam{Member:&Member{Name:"user1"}}
// List projects which user1 is the project admin : query := &QueryParam{Memeber:&Member{Name:"user1",Role:1}}
type ProjectQueryParam struct {
	Name       string       // the name of project
	Owner      string       // the username of project owner
	Public     *bool        // the project is public or not, can be ture, false and nil
	Member     *MemberQuery // the member of project
	Pagination *Pagination  // pagination information
}

// MemberQuery fitler by member's username and role
type MemberQuery struct {
	Name string // the username of member
	Role int    // the role of the member has to the project
}

// Pagination ...
type Pagination struct {
	Page int64
	Size int64
}

// BaseProjectCollection contains the query conditions which can be used
// to get a project collection. The collection can be used as the base to
// do other filter
type BaseProjectCollection struct {
	Public bool
	Member string
}

// ProjectRequest holds informations that need for creating project API
type ProjectRequest struct {
	Name                                       string `json:"project_name"`
	Public                                     int    `json:"public"`
	EnableContentTrust                         bool   `json:"enable_content_trust"`
	PreventVulnerableImagesFromRunning         bool   `json:"prevent_vulnerable_images_from_running"`
	PreventVulnerableImagesFromRunningSeverity string `json:"prevent_vulnerable_images_from_running_severity"`
	AutomaticallyScanImagesOnPush              bool   `json:"automatically_scan_images_on_push"`
}

// ProjectQueryResult ...
type ProjectQueryResult struct {
	Total    int64
	Projects []*Project
}
