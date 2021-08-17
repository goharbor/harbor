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
	"github.com/goharbor/harbor/src/common/security/robot"
	robotCtr "github.com/goharbor/harbor/src/controller/robot"
	pkgModels "github.com/goharbor/harbor/src/pkg/project/models"

	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/controller/event/metadata"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification"
	repomodel "github.com/goharbor/harbor/src/pkg/repository/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/repository"
)

func newRepositoryAPI() *repositoryAPI {
	return &repositoryAPI{
		proCtl:  project.Ctl,
		repoCtl: repository.Ctl,
		artCtl:  artifact.Ctl,
	}
}

type repositoryAPI struct {
	BaseAPI
	proCtl  project.Controller
	repoCtl repository.Controller
	artCtl  artifact.Controller
}

func (r *repositoryAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	if err := unescapePathParams(params, "RepositoryName"); err != nil {
		r.SendError(ctx, err)
	}

	return nil
}

func (r *repositoryAPI) ListAllRepositories(ctx context.Context, params operation.ListAllRepositoriesParams) middleware.Responder {
	// set query
	query, err := r.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return r.SendError(ctx, err)
	}
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return r.SendError(ctx, errors.UnauthorizedError(errors.New("security context not found")))
	}
	if !secCtx.IsSysAdmin() && !secCtx.IsSolutionUser() {
		projectIDs, err := r.listAuthorizedProjectIDs(ctx)
		if err != nil {
			return r.SendError(ctx, err)
		}
		// no authorized projects, return nil directly
		if len(projectIDs) == 0 {
			return operation.NewListAllRepositoriesOK().
				WithXTotalCount(0).
				WithLink(r.Links(ctx, params.HTTPRequest.URL, 0, query.PageNumber, query.PageSize).String()).
				WithPayload(nil)
		}
		orList := &q.OrList{}
		for _, projectID := range projectIDs {
			orList.Values = append(orList.Values, projectID)
		}
		query.Keywords["ProjectID"] = orList
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
		repos = append(repos, r.assembleRepository(ctx, model.NewRepoRecord(repository)))
	}
	return operation.NewListAllRepositoriesOK().
		WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(repos)
}

func (r *repositoryAPI) listAuthorizedProjectIDs(ctx context.Context) ([]int64, error) {
	secCtx, ok := security.FromContext(ctx)
	if !ok {
		return nil, errors.UnauthorizedError(errors.New("security context not found"))
	}
	query := &q.Query{
		Keywords: map[string]interface{}{},
	}
	if secCtx.IsAuthenticated() {
		switch secCtx.(type) {
		case *local.SecurityContext:
			currentUser := secCtx.(*local.SecurityContext).User()
			query.Keywords["member"] = &project.MemberQuery{
				UserID:     currentUser.UserID,
				GroupIDs:   currentUser.GroupIDs,
				WithPublic: true,
			}
		case *robot.SecurityContext:
			// for the system level robot that covers all the project, see it as the system admin.
			var coverAll bool
			var names []string
			r := secCtx.(*robot.SecurityContext).User()
			for _, p := range r.Permissions {
				if p.Scope == robotCtr.SCOPEALLPROJECT {
					coverAll = true
					break
				}
				names = append(names, p.Namespace)
			}
			if !coverAll {
				namesQuery := &pkgModels.NamesQuery{
					Names:      names,
					WithPublic: true,
				}
				query.Keywords["names"] = namesQuery
			}
		default:
			query.Keywords["public"] = true
		}
	} else {
		query.Keywords["public"] = true
	}

	projects, err := r.proCtl.List(ctx, query)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for _, project := range projects {
		ids = append(ids, project.ProjectID)
	}
	return ids, nil
}

func (r *repositoryAPI) ListRepositories(ctx context.Context, params operation.ListRepositoriesParams) middleware.Responder {
	if err := r.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceRepository); err != nil {
		return r.SendError(ctx, err)
	}
	project, err := r.proCtl.GetByName(ctx, params.ProjectName)
	if err != nil {
		return r.SendError(ctx, err)
	}

	// set query
	query, err := r.BuildQuery(ctx, params.Q, params.Sort, params.Page, params.PageSize)
	if err != nil {
		return r.SendError(ctx, err)
	}
	query.Keywords["ProjectID"] = project.ProjectID

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
		repos = append(repos, r.assembleRepository(ctx, model.NewRepoRecord(repository)))
	}
	return operation.NewListRepositoriesOK().
		WithXTotalCount(total).
		WithLink(r.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(repos)
}

func (r *repositoryAPI) GetRepository(ctx context.Context, params operation.GetRepositoryParams) middleware.Responder {
	if err := r.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceRepository); err != nil {
		return r.SendError(ctx, err)
	}
	repository, err := r.repoCtl.GetByName(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName))
	if err != nil {
		return r.SendError(ctx, err)
	}
	return operation.NewGetRepositoryOK().WithPayload(r.assembleRepository(ctx, model.NewRepoRecord(repository)))
}

func (r *repositoryAPI) assembleRepository(ctx context.Context, repository *model.RepoRecord) *models.Repository {
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
	if err := r.repoCtl.Update(ctx, &repomodel.RepoRecord{
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

	// fire event
	notification.AddEvent(ctx, &metadata.DeleteRepositoryEventMetadata{
		Ctx:        ctx,
		Repository: repository.Name,
		ProjectID:  repository.ProjectID,
	})

	return operation.NewDeleteRepositoryOK()
}
