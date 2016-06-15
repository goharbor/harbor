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
	"strings"

	"github.com/vmware/harbor/dao"
	"github.com/vmware/harbor/models"
	"github.com/vmware/harbor/service/cache"
	"github.com/vmware/harbor/utils/log"
)

// StatisticAPI handles request to /api/statistics/
type StatisticAPI struct {
	BaseAPI
	userID int
}

//Prepare validates the URL and the user
func (s *StatisticAPI) Prepare() {
	s.userID = s.ValidateUser()
}

// Get total projects and repos of the user
func (s *StatisticAPI) Get() {
	isAdmin, err := dao.IsAdminRole(s.userID)
	if err != nil {
		log.Errorf("Error occured in check admin, error: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	var projectList []models.Project
	if isAdmin {
		projectList, err = dao.GetAllProjects("")
	} else {
		projectList, err = dao.GetUserRelevantProjects(s.userID, "")
	}
	if err != nil {
		log.Errorf("Error occured in QueryProject, error: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	proMap := map[string]int{}
	proMap["my_project_count"] = 0
	proMap["my_repo_count"] = 0
	proMap["public_project_count"] = 0
	proMap["public_repo_count"] = 0
	var publicProjects []models.Project
	publicProjects, err = dao.GetPublicProjects("")
	if err != nil {
		log.Errorf("Error occured in QueryPublicProject, error: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "Internal error.")
	}
	proMap["public_project_count"] = len(publicProjects)
	for i := 0; i < len(publicProjects); i++ {
		proMap["public_repo_count"] += getRepoCountByProject(publicProjects[i].Name)
	}
	if isAdmin {
		proMap["total_project_count"] = len(projectList)
		proMap["total_repo_count"] = getTotalRepoCount()
	}
	for i := 0; i < len(projectList); i++ {
		if isAdmin {
			projectList[i].Role = models.PROJECTADMIN
		}
		if projectList[i].Role == models.PROJECTADMIN || projectList[i].Role == models.DEVELOPER ||
			projectList[i].Role == models.GUEST {
			proMap["my_project_count"]++
			proMap["my_repo_count"] += getRepoCountByProject(projectList[i].Name)
		}
	}
	s.Data["json"] = proMap
	s.ServeJSON()
}

//getReposByProject returns repo numbers of specified project
func getRepoCountByProject(projectName string) int {
	repoList, err := cache.GetRepoFromCache()
	if err != nil {
		log.Errorf("Failed to get repo from cache, error: %v", err)
		return 0
	}
	var resp int
	if len(projectName) > 0 {
		for _, r := range repoList {
			if strings.Contains(r, "/") && r[0:strings.LastIndex(r, "/")] == projectName {
				resp++
			}
		}
		return resp
	}
	return 0
}

//getTotalRepoCount returns total repo count
func getTotalRepoCount() int {
	repoList, err := cache.GetRepoFromCache()
	if err != nil {
		log.Errorf("Failed to get repo from cache, error: %v", err)
		return 0
	}
	return len(repoList)

}
