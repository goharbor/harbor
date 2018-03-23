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

package api

import (
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/vmware/harbor/src/common"
	"github.com/vmware/harbor/src/common/dao"
	"github.com/vmware/harbor/src/common/models"
	"github.com/vmware/harbor/src/common/utils"
	"github.com/vmware/harbor/src/common/utils/log"
	uiutils "github.com/vmware/harbor/src/ui/utils"
)

// SearchAPI handles requesst to /api/search
type SearchAPI struct {
	BaseController
}

type searchResult struct {
	Project    []*models.Project        `json:"project"`
	Repository []map[string]interface{} `json:"repository"`
}

// Get ...
func (s *SearchAPI) Get() {
	keyword := s.GetString("q")
	isAuthenticated := s.SecurityCtx.IsAuthenticated()
	isSysAdmin := s.SecurityCtx.IsSysAdmin()

	var projects []*models.Project
	var err error

	if isSysAdmin {
		result, err := s.ProjectMgr.List(nil)
		if err != nil {
			s.ParseAndHandleError("failed to get projects", err)
			return
		}
		projects = result.Projects
	} else {
		projects, err = s.ProjectMgr.GetPublic()
		if err != nil {
			s.ParseAndHandleError("failed to get projects", err)
			return
		}
		if isAuthenticated {
			mys, err := s.SecurityCtx.GetMyProjects()
			if err != nil {
				s.HandleInternalServerError(fmt.Sprintf(
					"failed to get projects: %v", err))
				return
			}
			exist := map[int64]bool{}
			for _, p := range projects {
				exist[p.ProjectID] = true
			}

			for _, p := range mys {
				if !exist[p.ProjectID] {
					projects = append(projects, p)
				}
			}
		}
	}

	projectSorter := &models.ProjectSorter{Projects: projects}
	sort.Sort(projectSorter)
	projectResult := []*models.Project{}
	for _, p := range projects {
		if len(keyword) > 0 && !strings.Contains(p.Name, keyword) {
			continue
		}

		if isAuthenticated {
			roles := s.SecurityCtx.GetProjectRoles(p.ProjectID)
			if len(roles) != 0 {
				p.Role = roles[0]
			}

			if p.Role == common.RoleProjectAdmin || isSysAdmin {
				p.Togglable = true
			}
		}

		total, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
			ProjectIDs: []int64{p.ProjectID},
		})
		if err != nil {
			log.Errorf("failed to get total of repositories of project %d: %v", p.ProjectID, err)
			s.CustomAbort(http.StatusInternalServerError, "")
		}

		p.RepoCount = total

		projectResult = append(projectResult, p)
	}

	repositoryResult, err := filterRepositories(projects, keyword)
	if err != nil {
		log.Errorf("failed to filter repositories: %v", err)
		s.CustomAbort(http.StatusInternalServerError, "")
	}

	result := &searchResult{Project: projectResult, Repository: repositoryResult}
	s.Data["json"] = result
	s.ServeJSON()
}

func filterRepositories(projects []*models.Project, keyword string) (
	[]map[string]interface{}, error) {

	repositories, err := dao.GetRepositories()
	if err != nil {
		return nil, err
	}

	i, j := 0, 0
	result := []map[string]interface{}{}
	for i < len(repositories) && j < len(projects) {
		r := repositories[i]
		p, _ := utils.ParseRepository(r.Name)
		d := strings.Compare(p, projects[j].Name)
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
			entry["project_id"] = projects[j].ProjectID
			entry["project_public"] = projects[j].IsPublic()
			entry["pull_count"] = r.PullCount

			tags, err := getTags(r.Name)
			if err != nil {
				return nil, err
			}
			entry["tags_count"] = len(tags)

			result = append(result, entry)
		} else {
			j++
		}
	}
	return result, nil
}

func getTags(repository string) ([]string, error) {
	client, err := uiutils.NewRepositoryClientForUI("harbor-ui", repository)
	if err != nil {
		return nil, err
	}

	tags, err := client.ListTag()
	if err != nil {
		return nil, err
	}

	return tags, nil
}
