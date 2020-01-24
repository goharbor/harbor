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
	"github.com/goharbor/harbor/src/pkg/project"
	"github.com/goharbor/harbor/src/pkg/q"
	"github.com/goharbor/harbor/src/pkg/repository"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/artifact"
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

// TODO do auth in a separate middleware

func (a *artifactAPI) ListArtifacts(ctx context.Context, params operation.ListArtifactsParams) middleware.Responder {
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
	option := &artifact.Option{
		WithTag: true, // return the tag by default
	}
	if params.WithTag != nil {
		option.WithTag = *(params.WithTag)
	}
	if option.WithTag {
		if params.WithImmutableStatus != nil {
			option.TagOption = &artifact.TagOption{
				WithImmutableStatus: *(params.WithImmutableStatus),
			}
		}
	}
	if params.WithLabel != nil {
		option.WithLabel = *(params.WithLabel)
	}
	if params.WithScanOverview != nil {
		option.WithScanOverview = *(params.WithScanOverview)
	}
	if params.WithSignatrue != nil {
		option.WithSignature = *(params.WithSignatrue)
	}

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
	// TODO implement
	return operation.NewGetArtifactOK()
}

func (a *artifactAPI) DeleteArtifact(ctx context.Context, params operation.DeleteArtifactParams) middleware.Responder {
	// TODO implement
	return operation.NewDeleteArtifactOK()
}
