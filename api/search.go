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
	"sort"
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	svc_utils "github.com/vmware/harbor/service/utils"
	"github.com/vmware/harbor/utils"

	"github.com/astaxie/beego"
)

type SearchAPI struct {
	BaseAPI
}

type SearchResult struct {
	Project    []map[string]interface{} `json:"project"`
	Repository []map[string]interface{} `json:"repository"`
}

func (n *SearchAPI) Get() {
	userID, ok := n.GetSession("userId").(int)
	if !ok {
		userID = dao.NON_EXIST_USER_ID
	}
	keyword := n.GetString("q")
	projects, err := dao.QueryRelevantProjects(userID)
	if err != nil {
		beego.Error("Failed to get projects of user id:", userID, ", error:", err)
		n.CustomAbort(http.StatusInternalServerError, "Failed to get project search result")
	}
	projectSorter := &utils.ProjectSorter{Projects: projects}
	sort.Sort(projectSorter)
	projectResult := []map[string]interface{}{}
	for _, p := range projects {
		match := true
		if len(keyword) > 0 && !strings.Contains(p.Name, keyword) {
			match = false
		}
		if match {
			entry := make(map[string]interface{})
			entry["id"] = p.ProjectId
			entry["name"] = p.Name
			entry["public"] = p.Public
			projectResult = append(projectResult, entry)
		}
	}

	repositories, err2 := svc_utils.GetRepoFromCache()
	if err2 != nil {
		beego.Error("Failed to get repos from cache, error :", err2)
		n.CustomAbort(http.StatusInternalServerError, "Failed to get repositories search result")
	}
	sort.Strings(repositories)
	repositoryResult := filterRepositories(repositories, projects, keyword)
	result := &SearchResult{Project: projectResult, Repository: repositoryResult}
	n.Data["json"] = result
	n.ServeJSON()
}

func filterRepositories(repositories []string, projects []models.Project, keyword string) []map[string]interface{} {
	var i, j int = 0, 0
	result := []map[string]interface{}{}
	for i < len(repositories) && j < len(projects) {
		r := &utils.Repository{Name: repositories[i]}
		d := strings.Compare(r.GetProject(), projects[j].Name)
		if d < 0 {
			i++
			continue
		} else if d == 0 {
			i++
			if len(keyword) != 0 && !strings.Contains(r.Name, keyword) {
				continue
			}
			entry := make(map[string]interface{})
			entry["repository_name"] = r.Name
			entry["project_name"] = projects[j].Name
			entry["project_id"] = projects[j].ProjectId
			entry["project_public"] = projects[j].Public
			result = append(result, entry)
		} else {
			j++
		}
	}
	return result
}
