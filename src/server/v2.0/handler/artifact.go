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
	"github.com/goharbor/harbor/src/common/rbac"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/artifact"
	"time"
)

func newArtifactAPI() *artifactAPI {
	return &artifactAPI{
		artCtl:  artifact.Ctl,
		proMgr:  project.Mgr,
		repoMgr: repository.Mgr,
	}
}

type artifactAPI struct {
	BaseAPI
	artCtl  artifact.Controller
	proMgr  project.Manager
	repoMgr repository.Manager
}

func (a *artifactAPI) ListArtifacts(ctx context.Context, params operation.ListArtifactsParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	// set query
	query := &q.Query{
		Keywords: map[string]interface{}{},
	}
	if params.Type != nil {
		query.Keywords["Type"] = *(params.Type)
	}
	if params.Page != nil {
		query.PageNumber = *(params.Page)
	}
	if params.PageSize != nil {
		query.PageSize = *(params.PageSize)
	}
	repository, err := a.repoMgr.GetByName(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName))
	if err != nil {
		return a.SendError(ctx, err)
	}
	query.Keywords["RepositoryID"] = repository.RepositoryID

	// set option
	option := option(params.WithTag, params.WithImmutableStatus,
		params.WithLabel, params.WithScanOverview, params.WithSignature)

	// list artifacts according to the query and option
	total, arts, err := a.artCtl.List(ctx, query, option)
	if err != nil {
		return a.SendError(ctx, err)
	}

	var artifacts []*models.Artifact
	for _, art := range arts {
		artifacts = append(artifacts, art.ToSwagger())
	}

	// TODO add link header
	return operation.NewListArtifactsOK().WithXTotalCount(total).WithLink("").WithPayload(artifacts)
}

func (a *artifactAPI) GetArtifact(ctx context.Context, params operation.GetArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionRead, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	// set option
	option := option(params.WithTag, params.WithImmutableStatus,
		params.WithLabel, params.WithScanOverview, params.WithSignature)

	// get the artifact
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, option)
	if err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewGetArtifactOK().WithPayload(artifact.ToSwagger())
}

func (a *artifactAPI) DeleteArtifact(ctx context.Context, params operation.DeleteArtifactParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceArtifact); err != nil {
		return a.SendError(ctx, err)
	}
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName), params.Reference, nil)
	if err != nil {
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.Delete(ctx, artifact.ID); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewDeleteArtifactOK()
}

func (a *artifactAPI) CreateTag(ctx context.Context, params operation.CreateTagParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionCreate, rbac.ResourceTag); err != nil {
		return a.SendError(ctx, err)
	}
	art, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName),
		params.Reference, &artifact.Option{
			WithTag: true,
		})
	if err != nil {
		return a.SendError(ctx, err)
	}
	tag := &artifact.Tag{}
	tag.RepositoryID = art.RepositoryID
	tag.ArtifactID = art.ID
	tag.Name = params.Tag.Name
	tag.PushTime = time.Now()
	if _, err = a.artCtl.CreateTag(ctx, tag); err != nil {
		return a.SendError(ctx, err)
	}
	// TODO set location header?
	return operation.NewCreateTagCreated()
}

func (a *artifactAPI) DeleteTag(ctx context.Context, params operation.DeleteTagParams) middleware.Responder {
	if err := a.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionDelete, rbac.ResourceTag); err != nil {
		return a.SendError(ctx, err)
	}
	artifact, err := a.artCtl.GetByReference(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName),
		params.Reference, &artifact.Option{
			WithTag: true,
		})
	if err != nil {
		return a.SendError(ctx, err)
	}
	var id int64
	for _, tag := range artifact.Tags {
		if tag.Name == params.TagName {
			id = tag.ID
			break
		}
	}
	// the tag not found
	if id == 0 {
		err = ierror.New(nil).WithCode(ierror.NotFoundCode).WithMessage(
			"tag %s attached to artifact %d not found", params.TagName, artifact.ID)
		return a.SendError(ctx, err)
	}
	if err = a.artCtl.DeleteTag(ctx, id); err != nil {
		return a.SendError(ctx, err)
	}
	return operation.NewDeleteTagOK()
}

func option(withTag, withImmutableStatus, withLabel, withScanOverview, withSignature *bool) *artifact.Option {
	option := &artifact.Option{
		WithTag: true, // return the tag by default
	}
	if withTag != nil {
		option.WithTag = *(withTag)
	}
	if option.WithTag {
		if withImmutableStatus != nil {
			option.TagOption = &artifact.TagOption{
				WithImmutableStatus: *(withImmutableStatus),
			}
		}
	}
	if withLabel != nil {
		option.WithLabel = *(withLabel)
	}
	if withScanOverview != nil {
		option.WithScanOverview = *(withScanOverview)
	}
	if withSignature != nil {
		option.WithSignature = *(withSignature)
	}
	return option
}
