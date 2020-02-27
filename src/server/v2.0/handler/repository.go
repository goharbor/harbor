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
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/api/artifact"
	"github.com/goharbor/harbor/src/api/repository"
	cmodels "github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/repository"
)

func newRepositoryAPI() *repositoryAPI {
	return &repositoryAPI{
		proMgr:  project.Mgr,
		repoCtl: repository.Ctl,
		artCtl:  artifact.Ctl,
	}
}

type repositoryAPI struct {
	BaseAPI
	// TODO replace proMgr with proCtl
	proMgr  project.Manager
	repoCtl repository.Controller
	artCtl  artifact.Controller
}

func (r *repositoryAPI) ListRepositories(ctx context.Context, params operation.ListRepositoriesParams) middleware.Responder {
	if err := r.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceRepository); err != nil {
		return r.SendError(ctx, err)
	}
	project, err := r.proMgr.Get(params.ProjectName)
	if err != nil {
		return r.SendError(ctx, err)
	}
	// set query
	query := &q.Query{
		Keywords: map[string]interface{}{
			"ProjectID": project.ProjectID,
		},
	}
	// TODO support fuzzy match
	if params.Name != nil {
		query.Keywords["Name"] = *(params.Name)
	}
	if params.Page != nil {
		query.PageNumber = *(params.Page)
	}
	if params.PageSize != nil {
		query.PageSize = *(params.PageSize)
	}
	total, err := r.repoCtl.Count(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	repositories, err := r.repoCtl.List(ctx, query)
	if err != nil {
		return r.SendError(ctx, err)
	}
	var repos []*models.Repository
	for _, repository := range repositories {
		repos = append(repos, r.assembleRepository(ctx, repository))
	}
	// TODO add link header
	return operation.NewListRepositoriesOK().WithXTotalCount(total).WithLink("").WithPayload(repos)
}

func (r *repositoryAPI) GetRepository(ctx context.Context, params operation.GetRepositoryParams) middleware.Responder {
	if err := r.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceRepository); err != nil {
		return r.SendError(ctx, err)
	}
	repository, err := r.repoCtl.GetByName(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName))
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetRepositoryOK().WithPayload(r.assembleRepository(ctx, repository))
}

func (r *repositoryAPI) assembleRepository(ctx context.Context, repository *cmodels.RepoRecord) *models.Repository {
	repo := repository.ToSwagger()
	total, err := r.artCtl.Count(ctx, &q.Query{
		Keywords: map[string]interface{}{
			"RepositoryID": repo.ID,
		},
	})
	if err != nil {
		log.Errorf("failed to get the count of artifacts under the repository %s: %v",
			repo.Name, err)
	}
	repo.ArtifactCount = total
	return repo
}

func (r *repositoryAPI) UpdateRepository(ctx context.Context, params operation.UpdateRepositoryParams) middleware.Responder {
	if err := r.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionUpdate, rbac.ResourceRepository); err != nil {
		return r.SendError(ctx, err)
	}
	repository, err := r.repoCtl.GetByName(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName))
	if err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.repoCtl.Update(ctx, &cmodels.RepoRecord{
		RepositoryID: repository.RepositoryID,
		Description:  params.Repository.Description,
	}, "Description"); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewDeleteRepositoryOK()
}

func (r *repositoryAPI) DeleteRepository(ctx context.Context, params operation.DeleteRepositoryParams) middleware.Responder {
	if err := r.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceRepository); err != nil {
		return r.SendError(ctx, err)
	}
	repository, err := r.repoCtl.GetByName(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName))
	if err != nil {
		return r.SendError(ctx, err)
	}
	if err := r.repoCtl.Delete(ctx, repository.RepositoryID); err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewDeleteRepositoryOK()
}
