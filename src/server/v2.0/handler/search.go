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

package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/core/api"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/config"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/search"
	"helm.sh/helm/v3/cmd/helm/search"
)

func newSearchAPI() *searchAPI {
	return &searchAPI{
		artifactCtl:   artifact.Ctl,
		projectCtl:    project.Ctl,
		repositoryCtl: repository.Ctl,

		chartMuseumEnabled: config.WithChartMuseum(),
		searchCharts: func(q string, namespaces []string) ([]*search.Result, error) {
			return api.GetChartController().SearchChart(q, namespaces)
		},
	}
}

type searchAPI struct {
	BaseAPI
	artifactCtl   artifact.Controller
	projectCtl    project.Controller
	repositoryCtl repository.Controller

	chartMuseumEnabled bool
	searchCharts       func(string, []string) ([]*search.Result, error)
}

func (s *searchAPI) Search(ctx context.Context, params operation.SearchParams) middleware.Responder {
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return s.SendError(ctx, fmt.Errorf("security not found in the context"))
	}

	kw := q.KeyWords{}

	if !secCtx.IsSysAdmin() {
		if sc, ok := secCtx.(*local.SecurityContext); ok && sc.IsAuthenticated() {
			user := sc.User()
			kw["member"] = &project.MemberQuery{
				UserID:     user.UserID,
				GroupIDs:   user.GroupIDs,
				WithPublic: true,
			}
		} else {
			kw["public"] = true
		}
	}

	projects, err := s.projectCtl.List(ctx, q.New(kw))
	if err != nil {
		return s.SendError(ctx, err)
	}

	projectResult := []*models.Project{}
	proNames := []string{}
	for _, p := range projects {
		proNames = append(proNames, p.Name)

		if params.Q != "" && !strings.Contains(p.Name, params.Q) {
			continue
		}

		if sc, ok := secCtx.(*local.SecurityContext); ok && sc.IsAuthenticated() {
			roles, err := s.projectCtl.ListRoles(ctx, p.ProjectID, sc.User())
			if err != nil {
				return s.SendError(ctx, errors.Wrap(err, "failed to list roles"))
			}
			p.RoleList = roles
			p.Role = highestRole(roles)
		}

		total, err := s.repositoryCtl.Count(ctx, q.New(q.KeyWords{"project_id": p.ProjectID}))
		if err != nil {
			log.Errorf("failed to get total of repositories of project %d: %v", p.ProjectID, err)
			return s.SendError(ctx, errors.Wrapf(err, "failed to get total of repositories of project %d", p.ProjectID))
		}

		p.RepoCount = total

		projectResult = append(projectResult, model.NewProject(p).ToSwagger())
	}

	repositoryResult, err := s.filterRepositories(ctx, projects, params.Q)
	if err != nil {
		log.Errorf("failed to filter repositories: %v", err)
		return s.SendError(ctx, errors.Wrap(err, "failed to filter repositories"))
	}

	chartResult, err := s.filterCharts(ctx, params.Q, proNames)
	if err != nil {
		log.Errorf("failed to filter charts: %v", err)
		return s.SendError(ctx, errors.Wrap(err, "failed to filter charts"))
	}

	return newSearchOK().WithPayload(&models.Search{
		Project:    projectResult,
		Repository: repositoryResult,
		Chart:      chartResult,
	})
}

func (s *searchAPI) filterRepositories(ctx context.Context, projects []*project.Project, keyword string) ([]*models.SearchRepository, error) {
	result := []*models.SearchRepository{}
	if len(projects) == 0 {
		return result, nil
	}

	repositories, err := s.repositoryCtl.List(ctx, q.New(q.KeyWords{"name": &q.FuzzyMatchValue{Value: keyword}}))
	if err != nil {
		return nil, err
	}

	if len(repositories) == 0 {
		return result, nil
	}

	projectMap := map[string]*project.Project{}
	for _, project := range projects {
		projectMap[project.Name] = project
	}

	for _, repository := range repositories {
		projectName, _ := utils.ParseRepository(repository.Name)
		project, exist := projectMap[projectName]
		if !exist {
			continue
		}

		entry := models.SearchRepository{
			RepositoryName: repository.Name,
			ProjectName:    project.Name,
			ProjectID:      repository.ProjectID,
			ProjectPublic:  project.IsPublic(),
			PullCount:      repository.PullCount,
		}

		count, err := s.artifactCtl.Count(ctx, q.New(q.KeyWords{"RepositoryID": repository.RepositoryID}))
		if err != nil {
			log.Errorf("failed to get the count of artifacts under the repository %s: %v",
				repository.Name, err)
		} else {
			entry.ArtifactCount = count
		}

		result = append(result, &entry)
	}

	return result, nil
}

func (s *searchAPI) filterCharts(ctx context.Context, q string, namespaces []string) ([]*models.SearchResult, error) {
	if !s.chartMuseumEnabled {
		return nil, nil
	}

	result := []*models.SearchResult{}
	if len(namespaces) == 0 {
		return result, nil
	}

	charts, err := s.searchCharts(q, namespaces)
	if err != nil {
		return nil, err
	}

	for _, chart := range charts {
		var entry models.SearchResult
		if err := lib.JSONCopy(&entry, chart); err != nil {
			return nil, err
		}

		result = append(result, &entry)
	}

	return result, nil
}

// searchOK removing the chart from the response when the chartmuseum is disabled
type searchOK struct {
	Payload interface{}
}

func (o *searchOK) WithPayload(payload *models.Search) *searchOK {
	if payload != nil {
		p := &struct {
			Chart      *[]*models.SearchResult    `json:"chart,omitempty"`
			Project    []*models.Project          `json:"project"`
			Repository []*models.SearchRepository `json:"repository"`
		}{
			Project:    payload.Project,
			Repository: payload.Repository,
		}

		if payload.Chart != nil {
			p.Chart = &payload.Chart
		}

		o.Payload = p
	}

	return o
}

func (o *searchOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {
	rw.WriteHeader(200)

	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

func newSearchOK() *searchOK {
	return &searchOK{}
}
