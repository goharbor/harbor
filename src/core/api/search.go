// Copyright 2018 Project Harbor Authors
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

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/orm"
	"github.com/goharbor/harbor/src/lib/q"
	"helm.sh/helm/v3/cmd/helm/search"
)

type chartSearchHandler func(string, []string) ([]*search.Result, error)

var searchHandler chartSearchHandler

// SearchAPI handles request to /api/search
type SearchAPI struct {
	BaseController
}

type searchResult struct {
	Project    []*models.Project        `json:"project"`
	Repository []map[string]interface{} `json:"repository"`
	Chart      *[]*search.Result        `json:"chart,omitempty"`
}

// Get ...
func (s *SearchAPI) Get() {
	keyword := s.GetString("q")

	query := q.New(q.KeyWords{})
	if keyword != "" {
		query.Keywords["name"] = &q.FuzzyMatchValue{Value: keyword}
	}

	if !s.SecurityCtx.IsSysAdmin() {
		if sc, ok := s.SecurityCtx.(*local.SecurityContext); ok && sc.IsAuthenticated() {
			user := sc.User()
			member := &project.MemberQuery{
				UserID:     user.UserID,
				GroupIDs:   user.GroupIDs,
				WithPublic: true,
			}
			query.Keywords["member"] = member
		} else {
			query.Keywords["public"] = true
		}
	}

	projects, err := s.ProjectCtl.List(s.Context(), query)
	if err != nil {
		s.ParseAndHandleError("failed to get projects", err)
		return
	}

	projectResult := []*models.Project{}
	proNames := []string{}
	for _, p := range projects {
		proNames = append(proNames, p.Name)

		if sc, ok := s.SecurityCtx.(*local.SecurityContext); ok && sc.IsAuthenticated() {
			roles, err := s.ProjectCtl.ListRoles(s.Context(), p.ProjectID, sc.User())
			if err != nil {
				s.SendInternalServerError(fmt.Errorf("failed to list roles: %v", err))
				return
			}
			p.RoleList = roles
			p.Role = highestRole(roles)
		}

		total, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
			ProjectIDs: []int64{p.ProjectID},
		})
		if err != nil {
			log.Errorf("failed to get total of repositories of project %d: %v", p.ProjectID, err)
			s.SendInternalServerError(fmt.Errorf("failed to get total of repositories of project %d: %v", p.ProjectID, err))
			return
		}

		p.RepoCount = total

		projectResult = append(projectResult, p)
	}

	repositoryResult, err := filterRepositories(projects, keyword)
	if err != nil {
		log.Errorf("failed to filter repositories: %v", err)
		s.SendInternalServerError(fmt.Errorf("failed to filter repositories: %v", err))
		return
	}

	result := &searchResult{
		Project:    projectResult,
		Repository: repositoryResult,
	}

	// If enable chart repository
	if config.WithChartMuseum() {
		if searchHandler == nil {
			searchHandler = chartController.SearchChart
		}

		chartResults, err := searchHandler(keyword, proNames)
		if err != nil {
			log.Errorf("failed to filter charts: %v", err)
			s.SendInternalServerError(err)
			return

		}
		result.Chart = &chartResults

	}

	s.Data["json"] = result
	s.ServeJSON()
}

func filterRepositories(projects []*models.Project, keyword string) (
	[]map[string]interface{}, error) {
	result := []map[string]interface{}{}
	if len(projects) == 0 {
		return result, nil
	}

	repositories, err := dao.GetRepositories(&models.RepositoryQuery{
		Name: keyword,
	})
	if err != nil {
		return nil, err
	}
	if len(repositories) == 0 {
		return result, nil
	}

	projectMap := map[string]*models.Project{}
	for _, project := range projects {
		projectMap[project.Name] = project
	}

	ctx := orm.NewContext(nil, dao.GetOrmer())
	for _, repository := range repositories {
		projectName, _ := utils.ParseRepository(repository.Name)
		project, exist := projectMap[projectName]
		if !exist {
			continue
		}
		entry := make(map[string]interface{})
		entry["repository_name"] = repository.Name
		entry["project_name"] = project.Name
		entry["project_id"] = project.ProjectID
		entry["project_public"] = project.IsPublic()
		entry["pull_count"] = repository.PullCount

		count, err := artifact.Ctl.Count(ctx, &q.Query{
			Keywords: map[string]interface{}{
				"RepositoryID": repository.RepositoryID,
			},
		})
		if err != nil {
			log.Errorf("failed to get the count of artifacts under the repository %s: %v",
				repository.Name, err)
		} else {
			entry["artifact_count"] = count
		}

		result = append(result, entry)
	}
	return result, nil
}

// Returns the highest role in the role list.
// This func should be removed once we deprecate the "current_user_role_id" in project API
// A user can have multiple roles and they may not have a strict ranking relationship
func highestRole(roles []int) int {
	if roles == nil {
		return 0
	}
	rolePower := map[int]int{
		common.RoleProjectAdmin: 50,
		common.RoleMaintainer:   40,
		common.RoleDeveloper:    30,
		common.RoleGuest:        20,
		common.RoleLimitedGuest: 10,
	}
	var highest, highestPower int
	for _, role := range roles {
		if p, ok := rolePower[role]; ok && p > highestPower {
			highest = role
			highestPower = p
		}
	}
	return highest
}
