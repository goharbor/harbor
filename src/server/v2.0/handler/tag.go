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
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/controller/repository"
	"github.com/goharbor/harbor/src/controller/tag"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/tag"
)

func newTagAPI() *tagAPI {
	return &tagAPI{
		repoCtl: repository.Ctl,
		tagCtl:  tag.Ctl,
	}
}

type tagAPI struct {
	BaseAPI
	repoCtl repository.Controller
	tagCtl  tag.Controller
}

func (t *tagAPI) Prepare(ctx context.Context, operation string, params interface{}) middleware.Responder {
	if err := unescapePathParams(params, "RepositoryName"); err != nil {
		t.SendError(ctx, err)
	}
	return nil
}

func (t *tagAPI) ListTags(ctx context.Context, params operation.ListTagsParams) middleware.Responder {
	if err := t.RequireProjectAccess(ctx, params.ProjectName, rbac.ActionList, rbac.ResourceTag); err != nil {
		return t.SendError(ctx, err)
	}
	// set query
	query, err := t.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return t.SendError(ctx, err)
	}

	repository, err := t.repoCtl.GetByName(ctx, fmt.Sprintf("%s/%s", params.ProjectName, params.RepositoryName))
	if err != nil {
		return t.SendError(ctx, err)
	}
	query.Keywords["RepositoryID"] = repository.RepositoryID

	// get the total count of tags
	total, err := t.tagCtl.Count(ctx, query)
	if err != nil {
		return t.SendError(ctx, err)
	}

	// set option
	option := &tag.Option{}
	if params.WithSignature != nil {
		option.WithSignature = *params.WithSignature
	}
	if params.WithImmutableStatus != nil {
		option.WithImmutableStatus = *params.WithImmutableStatus
	}
	// list tags according to the query and option
	tags, err := t.tagCtl.List(ctx, query, option)
	if err != nil {
		return t.SendError(ctx, err)
	}

	var ts []*models.Tag
	for _, tag := range tags {
		ts = append(ts, tag.ToSwagger())
	}
	return operation.NewListTagsOK().
		WithXTotalCount(total).
		WithLink(t.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(ts)
}
